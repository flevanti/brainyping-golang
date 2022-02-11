package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/initapp"
	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"
)

type checkStatusType struct {
	CheckId                   string        `bson:"checkid"`
	ResponseDbId              string        `bson:"responsedbid"`
	RequestId                 string        `bson:"requestid"`
	OwnerUid                  string        `bson:"owneruid"`
	CurrentStatus             string        `bson:"curreststatus"`
	CurrentStatusSince        time.Time     `bson:"currentstatussince"`
	CurrentStatusSinceUnix    int64         `bson:"currentstatussinceunix"`
	PreviousStatus            string        `bson:"previousstatus"`
	PreviousStatusSince       time.Time     `bson:"previousstatussince"`
	PreviousStatusSinceUnix   int64         `bson:"previousstatussinceunix"`
	PreviousStatusDuration    time.Duration `bson:"previousstatusduration"`
	PreviousStatusDurationSec int64         `bson:"previousstatusdurationsec"`
	ChangeProcessedUnix       int64         `bson:"changeprocessedunix"`
}
type markerType struct {
	RequestId    string
	ResponseDbId string
	source       string
}

var checksStatuses = map[string]checkStatusType{}
var checksStatusesMutex = sync.Mutex{}
var marker markerType

const STATUSINIT = "INIT"
const STATUSOK = "OK"
const STATUSNOK = "NOK"
const BULKSAVESIZE = 1000
const markerSourceResponses = "RESPONSES"
const markerSourceStatusChanges = "STATUSCHANGES"
const STMSAVEAUTOFLUSHMS = "STM_SAVE_AUTO_FLUSH_MS"

func main() {
	var chReadResponses = make(chan dbhelper.CheckResponseRecordDb, 100)
	var chWriteStatusChanges = make(chan string, 100)
	var chWriteStatusCurrent = make(chan string, 100)

	initapp.InitApp()
	// TODO load current checks statuses
	//  load last status monitor marker (to know where to continue from....)
	retrieveLastStatusMonitorMarker()
	fmt.Printf("MARKER FOUND: RID %s MID %s SOURCE %s\n", marker.RequestId, marker.ResponseDbId, marker.source)

	// read responses...
	go readResponsesFromMarker(chReadResponses)

	// detect changes...
	go detectStatusChangesListener(chReadResponses, chWriteStatusChanges, chWriteStatusCurrent)

	// write changes to db
	go writeStatusChangesToDbBuffer(chWriteStatusChanges)

	go writeStatusCurrentToDbBuffer(chWriteStatusCurrent)

	select {}

}

func writeStatusCurrentToDbBuffer(chWriteStatusCurrent chan string) {

	var recordI interface{}
	for {
		select {
		case checkId := <-chWriteStatusCurrent:

			checksStatusesMutex.Lock()
			recordI = checksStatuses[checkId]
			checksStatusesMutex.Unlock()

			saveCheckCurrentStatus(checkId, recordI)
		} // end select
	} // end for

}

func retrieveLastStatusMonitorMarker() {
	retrieveMarkerFromStatusChanges()
	if marker.ResponseDbId != "" {
		// found the marker no need to continue digging...
		return
	}
	// let's try to recorver the marker from the responses...
	retrieveMarkerFromResponses()
	if marker.ResponseDbId != "" {
		// found the marker in the responses, yeah! we are done here...
		return
	}

	// OK here we are again with a "no document found" error....
	// this means also the responses collection is empty... nothing bad with it but... we cannot continue...

	fmt.Println("No marker could be found or retrieved")
	fmt.Println("No data found in responses or status changes collection, unable to monitor the situation")
	fmt.Println("Make sure workers are working and try again")
	os.Exit(1)

}

func detectStatusChangesListener(chReadResponses chan dbhelper.CheckResponseRecordDb, chWriteStatusChanges chan string, chWriteStatusCurrent chan string) {

	for {

		select {
		case record := <-chReadResponses:
			if detectStatusChanges(&record) {
				chWriteStatusCurrent <- record.CheckId
				chWriteStatusChanges <- record.CheckId
			}

		default:

		} // end select loop

	} // end for loop

}

func writeStatusChangesToDbBuffer(chWriteStatusChangesToDb chan string) {
	var recordsToSave []interface{}
	var lastSaved time.Time = time.Now()

	for {
		select {
		case checkId := <-chWriteStatusChangesToDb:
			checksStatusesMutex.Lock()
			recordsToSave = append(recordsToSave, checksStatuses[checkId])
			checksStatusesMutex.Unlock()
		default:

		} // end select
		if len(recordsToSave) >= BULKSAVESIZE || time.Since(lastSaved) > settings.GetSettDuration(STMSAVEAUTOFLUSHMS)*time.Millisecond {
			writeStatusChangesToDb(&recordsToSave)
			recordsToSave = []interface{}{}
			lastSaved = time.Now()
		}
	} // end for loop
}

func writeStatusChangesToDb(records *[]interface{}) {
	if len(*records) == 0 {
		return
	}
	utilities.FailOnError(dbhelper.SaveManyRecords(dbhelper.GetDatabaseName(), dbhelper.TablenameChecksStatusChanges, records))
}

func detectStatusChanges(record *dbhelper.CheckResponseRecordDb) bool {
	var responseStatusString string
	if record.Success {
		responseStatusString = STATUSOK
	} else {
		responseStatusString = STATUSNOK
	}

	// create the element for the current checkID in the statuschanges element....
	checksStatusesMutex.Lock()
	initialiseCheckStatusElement(record)
	checksStatusesMutex.Unlock()

	if checksStatuses[record.CheckId].CurrentStatus != responseStatusString {
		// status change detected...
		checksStatusesMutex.Lock()
		updateCheckStatusElement(record, responseStatusString)
		checksStatusesMutex.Unlock()
		logChange(record.CheckId)
		return true
	}
	return false
}

func updateCheckStatusElement(record *dbhelper.CheckResponseRecordDb, newStatus string) {

	statusRecord := checksStatuses[record.CheckId]

	// please note that we are updating the "previous" metadata using the current metadata that is going to become ... old

	statusRecord.ResponseDbId = record.MongoDbId
	statusRecord.RequestId = record.RequestId
	statusRecord.PreviousStatus = statusRecord.CurrentStatus
	statusRecord.PreviousStatusSince = statusRecord.CurrentStatusSince
	statusRecord.PreviousStatusSinceUnix = statusRecord.PreviousStatusSince.Unix()
	statusRecord.CurrentStatus = newStatus
	statusRecord.CurrentStatusSince = time.Unix(record.ProcessedUnix, 0)
	statusRecord.CurrentStatusSinceUnix = statusRecord.CurrentStatusSince.Unix()
	statusRecord.ChangeProcessedUnix = time.Now().Unix()
	statusRecord.PreviousStatusDuration = statusRecord.CurrentStatusSince.Sub(statusRecord.PreviousStatusSince)
	checksStatuses[record.CheckId] = statusRecord
}

func initialiseCheckStatusElement(record *dbhelper.CheckResponseRecordDb) {
	if _, exists := checksStatuses[record.CheckId]; !exists {
		newStatusChange := checkStatusType{
			CheckId:                 record.CheckId,
			RequestId:               record.RequestId,
			OwnerUid:                record.OwnerUid,
			CurrentStatus:           STATUSINIT,
			CurrentStatusSince:      time.Unix(record.ProcessedUnix, 0),
			CurrentStatusSinceUnix:  record.ProcessedUnix,
			PreviousStatus:          "",
			PreviousStatusSince:     time.Now(),
			PreviousStatusSinceUnix: time.Now().Unix(),
			PreviousStatusDuration:  time.Duration(0),
			ChangeProcessedUnix:     time.Now().Unix(),
		}
		checksStatuses[record.CheckId] = newStatusChange
		// logChange(record.CheckId)
	}
}

func logChange(checkId string) {
	log.Printf("Status change detected at %s for CID [%s] RID [%s] new status [%s] previously was [%s] for [%s]\n",
		checksStatuses[checkId].CurrentStatusSince.Format(time.Stamp),
		checksStatuses[checkId].CheckId,
		checksStatuses[checkId].RequestId,
		checksStatuses[checkId].CurrentStatus,
		checksStatuses[checkId].PreviousStatus,
		checksStatuses[checkId].PreviousStatusDuration)
}
