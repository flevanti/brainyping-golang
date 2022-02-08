package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/initapp"
)

type checkStatusType struct {
	CheckId                string        `json:"checkid"`
	RequestId              string        `json:"requestid"`
	OwnerUid               string        `json:"owneruid"`
	CurrentStatus          string        `json:"curreststatus"`
	CurrentStatusSince     time.Time     `json:"currentstatussince"`
	PreviousStatus         string        `json:"previousstatus"`
	PreviousStatusSince    time.Time     `json:"previousstatussince"`
	PreviousStatusDuration time.Duration `json:"previousstatusduration"`
	ChangeprocessedUnix    int64         `json:"changeprocessedunix"`
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
const markerSourceStatusChanges = "STCHANGES"

func main() {
	var chReadResponses = make(chan dbhelper.CheckResponseRecordDb, 100)
	var chWriteStatusChanges = make(chan string, 100)

	initapp.InitApp()
	// TODO load current checks statuses
	//  load last status monitor marker (to know where to continue from....)
	retrieveLastStatusMonitorMarker()
	fmt.Printf("MARKER FOUND: RID %s MID %s SOURCE %s\n", marker.RequestId, marker.ResponseDbId, marker.source)

	// read responses...
	go readResponsesFromMarker(chReadResponses)

	// detect changes...
	go detectStatusChangesListener(chReadResponses, chWriteStatusChanges)

	// write changes to db
	writeStatusChangesToDb(chWriteStatusChanges)

	select {}

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

func detectStatusChangesListener(chReadResponses chan dbhelper.CheckResponseRecordDb, chWriteStatusChanges chan string) {

	for {

		select {
		case record := <-chReadResponses:
			if detectStatusChanges(&record) {
				chWriteStatusChanges <- record.CheckId
			}

		default:

		} // end select loop

	} // end for loop

}

func writeStatusChangesToDb(writeStatusChangesToDb chan string) {
	var recordsToSave []interface{}
	for {
		select {
		case checkId := <-writeStatusChangesToDb:
			checksStatusesMutex.Lock()
			recordsToSave = append(recordsToSave, checksStatuses[checkId])
			checksStatusesMutex.Unlock()
		}
	}
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

	statusRecord.RequestId = record.RequestId
	statusRecord.PreviousStatus = statusRecord.CurrentStatus
	statusRecord.PreviousStatusSince = statusRecord.CurrentStatusSince
	statusRecord.CurrentStatus = newStatus
	statusRecord.CurrentStatusSince = time.Unix(record.ProcessedUnix, 0)
	statusRecord.ChangeprocessedUnix = time.Now().Unix()
	statusRecord.PreviousStatusDuration = statusRecord.CurrentStatusSince.Sub(statusRecord.PreviousStatusSince)
	checksStatuses[record.CheckId] = statusRecord
}

func initialiseCheckStatusElement(record *dbhelper.CheckResponseRecordDb) {
	if _, exists := checksStatuses[record.CheckId]; !exists {
		newStatusChange := checkStatusType{
			CheckId:                record.CheckId,
			RequestId:              record.RequestId,
			OwnerUid:               record.OwnerUid,
			CurrentStatus:          STATUSINIT,
			CurrentStatusSince:     time.Unix(record.ProcessedUnix, 0),
			PreviousStatus:         "",
			PreviousStatusSince:    time.Now(),
			PreviousStatusDuration: time.Duration(0),
			ChangeprocessedUnix:    time.Now().Unix(),
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
