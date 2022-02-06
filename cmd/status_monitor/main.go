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
	checkId             string
	requestId           string
	ownerUid            string
	currentStatus       string
	currentStatusSince  time.Time
	previousStatus      string
	previousStatusSince time.Time
	changeprocessedUnix int64
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
	var chWriteStatusChanges = make(chan dbhelper.CheckStatusChangeRecordDb, 100)

	initapp.InitApp()
	// TODO load current checks statuses
	//  load last status monitor marker (to know where to continue from....)
	retrieveLastStatusMonitorMarker()
	fmt.Printf("MARKER FOUND: RID %s MID %s SOURCE %s\n", marker.RequestId, marker.ResponseDbId, marker.source)

	// read responses...
	go readResponsesFromMarker(chReadResponses)

	// detect changes...
	go detectStatusChangesListener(chReadResponses, chWriteStatusChanges)
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

func detectStatusChangesListener(chReadResponses chan dbhelper.CheckResponseRecordDb, chWriteStatusChanges chan dbhelper.CheckStatusChangeRecordDb) {

	for {

		select {
		case record := <-chReadResponses:
			detectStatusChanges(&record)

		default:

		} // end select loop

	} // end for loop

}

func detectStatusChanges(record *dbhelper.CheckResponseRecordDb) {
	var responseStatusString string
	if record.Success {
		responseStatusString = STATUSOK
	} else {
		responseStatusString = STATUSNOK
	}

	// create the element for the current checkID in the statuschanges element....
	initialiseCheckStatusElement(record)

	if checksStatuses[record.CheckId].currentStatus != responseStatusString {
		// status change detected...
		checksStatusesMutex.Lock()
		updateCheckStatusElement(record, responseStatusString)
		checksStatusesMutex.Unlock()
		logChange(record.CheckId)
	}

}

func updateCheckStatusElement(record *dbhelper.CheckResponseRecordDb, newStatus string) {

	statusRecord := checksStatuses[record.CheckId]

	statusRecord.requestId = record.RequestId
	statusRecord.previousStatus = statusRecord.currentStatus
	statusRecord.previousStatusSince = statusRecord.currentStatusSince
	statusRecord.currentStatus = newStatus
	statusRecord.currentStatusSince = time.Unix(record.ProcessedUnix, 0)
	statusRecord.changeprocessedUnix = time.Now().Unix()

	checksStatuses[record.CheckId] = statusRecord
}

func initialiseCheckStatusElement(record *dbhelper.CheckResponseRecordDb) {
	if _, exists := checksStatuses[record.CheckId]; !exists {
		newStatusChange := checkStatusType{
			checkId:             record.CheckId,
			requestId:           record.RequestId,
			ownerUid:            record.OwnerUid,
			currentStatus:       STATUSINIT,
			currentStatusSince:  time.Now(),
			previousStatus:      "",
			previousStatusSince: time.Now(),
			changeprocessedUnix: time.Now().Unix(),
		}
		checksStatuses[record.CheckId] = newStatusChange
		// logChange(record.CheckId)
	}
}

func logChange(checkId string) {

	log.Printf("Status change detected for CID [%s] RID [%s] new status [%s] previously was [%s] for [%s]\n",
		checksStatuses[checkId].checkId,
		checksStatuses[checkId].requestId,
		checksStatuses[checkId].currentStatus,
		checksStatuses[checkId].previousStatus,
		checksStatuses[checkId].currentStatusSince.Sub(checksStatuses[checkId].previousStatusSince))
}
