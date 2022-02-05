package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/initapp"
	"brainyping/pkg/queuehelper"
	"brainyping/pkg/settings"
	_ "brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func main() {
	initapp.InitApp()
	queuehelper.InitQueue()

	// create the context
	ctx, cfunc := context.WithCancel(context.Background())
	defer cfunc()

	// create the channel used by the queue consumer to buffer fetched messages
	chReceive := make(chan amqp.Delivery, settings.GetSettInt("RC_BUF_CH_SIZE"))
	chSave := make(chan dbhelper.CheckResponseRecordDb)

	// pass the context cancel function to the close handler
	closeHandler(cfunc)

	// start the queue consumer...
	go queuehelper.ConsumeQueueForResponsesToChecks(ctx, chReceive)

	dbhelper.Connect()
	defer dbhelper.Disconnect()

	if !dbhelper.CheckIfTableExists(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameResponse) {
		opts := options.CreateCollectionOptions{}
		utilities.FailOnError(dbhelper.CreateTable(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameResponse, &opts))
	}

	// waiting for the world to end - instructions to run before closing...
	go waitingForTheWorldToEnd(ctx)

	go ShowStatistics(ctx, chReceive, saveBuffer)

	go receiveResponses(ctx, chReceive, chSave)

	go saveResponseInBuffer(chSave)

	// wait forever!
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

			if time.Since(metadata.lastMsgTime) > settings.GetSettDuration("RC_GRACE_PERIOD_MS")*time.Millisecond {
				metadata.stopped = true
				break forloop
			}
		default:
			time.Sleep(100 * time.Millisecond)
		} // end select case
	} // end for/loop [forloop]

}

func saveResponseInBuffer(chsave chan dbhelper.CheckResponseRecordDb) {
	var lastSaved time.Time = time.Now()
	for {
		select {
		case record := <-chsave:
			saveBuffer = append(saveBuffer, record)
			if len(saveBuffer) >= settings.GetSettInt("RC_BULK_SAVE_SIZE") || time.Since(lastSaved) > settings.GetSettDuration("RC_SAVE_AUTO_FLUSH_MS")*time.Second {
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
		// not having a default make sure this is a blocking select/case until context is done...
	}

	// set a global flag to true to acknowledge the world is ending...
	endOfTheWorld = true

	for metadata.stopped == false {

	}

	time.Sleep(time.Second * 2)
	fmt.Print("\n\nBYE BYE\n\n")

	// this is it, it has been fun!
	os.Exit(0)

}
