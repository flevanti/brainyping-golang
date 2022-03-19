package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/heartbeat"
	"brainyping/pkg/initapp"
	"brainyping/pkg/internalstatusmonitorapi"
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
var ctx context.Context
var cfunc context.CancelFunc

const STATUSINIT = "INIT"
const STATUSOK = "OK"
const STATUSNOK = "NOK"
const BULKSAVESIZE = 1000
const markerSourceResponses = "RESPONSES"
const markerSourceStatusChanges = "STATUSCHANGES"
const STMSAVEAUTOFLUSHMS = "STM_SAVE_AUTO_FLUSH_MS"
const STMAPIPORT = "STM_API_PORT"

func main() {
	var chReadResponses = make(chan dbhelper.CheckResponseRecordDb, 100)
	var chWriteStatusChanges = make(chan string, 100)
	var chWriteStatusCurrent = make(chan string, 100)

	initapp.InitApp("STATUSMONITOR")

	// start the listener for internal status monitoring
	internalstatusmonitorapi.StartListener(settings.GetSettStr(STMAPIPORT), initapp.GetAppRole())

	// start the beating..
	heartbeat.New(utilities.RetrieveHostName(), initapp.RetrieveHostNameFriendly(), initapp.GetAppRole(), "-", "-", time.Second*60, dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameHeartbeats, settings.GetSettStr(STMAPIPORT), utilities.RetrievePublicIP()).Start()

	// create the context
	ctx, cfunc = context.WithCancel(context.Background())
	defer cfunc()

	go closeHandler()

	fmt.Println("Loading last known checks statuses...")
	loadLastKnownStatus()

	retrieveLastStatusMonitorMarker()
	fmt.Printf("MARKER FOUND: RID %s MID %s SOURCE %s\n", marker.RequestId, marker.ResponseDbId, marker.source)

	// read responses...
	go readResponsesFromMarker(chReadResponses)

	// detect changes...
	go detectStatusChangesListener(chReadResponses, chWriteStatusChanges, chWriteStatusCurrent)

	// write changes to db
	go writeStatusChangesToDbBuffer(chWriteStatusChanges)

	go writeStatusCurrentToDbBuffer(chWriteStatusCurrent)

	go waitingForTheWorldToEnd()

	select {}

}

func waitingForTheWorldToEnd() {
	select {
	case <-ctx.Done():
		break
		// not having a default make sure this is a blocking select/case until context is done...
	}

	fmt.Print("\n\nEXITING....\n\n")

	time.Sleep(time.Second * 2)
	fmt.Print("\n\nBYE BYE\n\n")

	// this is it, it has been fun!
	os.Exit(0)

}

func closeHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	cfunc()
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
		case <-ctx.Done():
			return
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
				logChange(record.CheckId)
			}
		case <-ctx.Done():
			fmt.Println("Status change goroutine listener ended")
			return
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
		case <-ctx.Done():
			writeStatusChangesToDb(&recordsToSave)
			fmt.Println("Status changes buffer flushed")
			return
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
	statusRecord.PreviousStatusDurationSec = int64(time.Duration(statusRecord.PreviousStatusDuration) * time.Second)
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
