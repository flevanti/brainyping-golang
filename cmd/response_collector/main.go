package main

import (
	"brainyping/pkg/dbhelper"
	"brainyping/pkg/queuehelper"
	"brainyping/pkg/utilities"
	"context"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type metadataType struct {
	msgReceived   uint64
	msgFailed     uint64
	lastMsgTime   time.Time
	inGracePeriod bool
	stopped       bool
}

var endOfTheWorld bool
var metadata metadataType
var saveBuffer []interface{}

const BUFFEREDCHANNELSIZE int = 100
const GRACEPERIOD time.Duration = time.Second * 5
const BULKSAVESIZE int = 1000
const BULKSAVETHRESHOLDTIME time.Duration = time.Second * 10

func main() {
	//create the context
	ctx, cfunc := context.WithCancel(context.Background())
	defer cfunc()

	//create the channel used by the queue consumer to buffer fetched messages
	chReceive := make(chan amqp.Delivery, BUFFEREDCHANNELSIZE)
	chSave := make(chan dbhelper.CheckResponseRecordDb)

	//pass the context cancel function to the close handler
	closeHandler(cfunc)

	//start the queue consumer...
	go queuehelper.ConsumeQueueForResponsesToChecks(ctx, chReceive)

	dbhelper.Connect()
	defer dbhelper.Disconnect()

	if !dbhelper.CheckIfTableExists(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameResponse) {
		opts := options.CreateCollectionOptions{}
		utilities.FailOnError(dbhelper.CreateTable(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameResponse, &opts))
	}

	//waiting for the world to end - instructions to run before closing...
	go waitingForTheWorldToEnd(ctx)

	go ShowStatistics(ctx, chReceive, saveBuffer)

	go receiveResponses(ctx, chReceive, chSave)

	go saveResponseInBuffer(chSave)

	//wait forever!
	select {}

}

func receiveResponses(ctx context.Context, ch <-chan amqp.Delivery, chsave chan dbhelper.CheckResponseRecordDb) {

	var messageQueued queuehelper.CheckRecordQueued
	var err error

forloop:
	for {
		select {
		case response := <-ch:
			metadata.msgReceived++
			metadata.lastMsgTime = time.Now()
			err = json.Unmarshal(response.Body, &messageQueued)
			utilities.FailOnError(err)
			if messageQueued.RecordOutcome.Success {
				metadata.msgFailed++
			}
			messageQueued.ReceivedByResponseHandler = time.Now().Unix()
			chsave <- prepareRecordToBeSaved(messageQueued)
			_ = response.Ack(false)
		case <-ctx.Done():
			metadata.inGracePeriod = true
			if time.Since(metadata.lastMsgTime) > GRACEPERIOD {
				metadata.stopped = true
				break forloop
			}
		default:
		} //end select case
	} //end for/loop [forloop]

}

func saveResponseInBuffer(chsave chan dbhelper.CheckResponseRecordDb) {
	var lastSaved time.Time = time.Now()
	for {
		select {
		case record := <-chsave:
			saveBuffer = append(saveBuffer, record)
			if len(saveBuffer) >= BULKSAVESIZE || time.Since(lastSaved) > BULKSAVETHRESHOLDTIME {
				saveResponsesInDatabase()
			}
		default:
		}
	}
}

func saveResponsesInDatabase() {
	err := dbhelper.SaveManyRecords(dbhelper.GetDatabaseName(), dbhelper.TablenameResponse, &saveBuffer)
	utilities.FailOnError(err)
	saveBuffer = nil
}

func prepareRecordToBeSaved(record queuehelper.CheckRecordQueued) dbhelper.CheckResponseRecordDb {
	var response dbhelper.CheckResponseRecordDb

	response.CheckId = record.Record.CheckId
	response.OwnerUid = record.Record.OwnerUid
	response.ScheduledTimeUnix = record.ScheduledUnix
	response.ProcessedUnix = record.RecordOutcome.CreatedUnix
	response.ScheduledTimeDelay = record.RecordOutcome.CreatedUnix - record.ScheduledUnix
	response.Region = record.RecordOutcome.Region
	response.QueuedRequestUnix = record.QueuedUnix
	response.ReceivedByWorkerUnix = record.ReceivedByWorkerUnix
	response.QueuedResponseUnix = record.QueuedReturnUnix
	response.ReceivedResponseUnix = record.ReceivedByResponseHandler
	response.TimeSpent = record.RecordOutcome.TimeSpent
	response.Success = record.RecordOutcome.Success
	response.ErrorOriginal = record.RecordOutcome.ErrorOriginal
	response.ErrorFriendly = record.RecordOutcome.ErrorFriendly
	response.ErrorInternal = record.RecordOutcome.ErrorInternal
	response.ErrorFatal = record.ErrorFatal
	response.Message = record.RecordOutcome.Message
	response.Redirects = record.RecordOutcome.Redirects
	response.RedirectsHistory = record.RecordOutcome.RedirectsHistory
	response.CreatedUnix = time.Now().Unix()
	return response

}

func closeHandler(cfunc context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cfunc()
	}()
}

func waitingForTheWorldToEnd(ctx context.Context) {
	select {
	case <-ctx.Done():
		break
		//not having a default make sure this is a blocking select/case until context is done...
	}

	//set a global flag to true to acknowledge the world is ending...
	endOfTheWorld = true

	for metadata.stopped == false {

	}

	time.Sleep(time.Second * 2)
	fmt.Print("\n\nBYE BYE\n\n")

	//this is it, it has been fun!
	os.Exit(0)

}
