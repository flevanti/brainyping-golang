package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"brainyping/pkg/checks"
	"brainyping/pkg/checks/httpcheck"
	"brainyping/pkg/dbhelper"
	"brainyping/pkg/heartbeat"
	"brainyping/pkg/initapp"
	"brainyping/pkg/internalstatusmonitorapi"
	"brainyping/pkg/queuehelper"
	"brainyping/pkg/settings"
	_ "brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"github.com/streadway/amqp"
)

type workerMetadataType struct {
	startTime    time.Time
	msgReceived  int64
	msgFailed    int64
	workerID     int
	lastMsgTime  time.Time
	WorkerStatus string
}

type workerStatusType struct {
	statusText string
	statusIcon string
}

type workersMetadataType struct {
	workerMetadata          []workerMetadataType
	workersTotalMsgReceived int64
}

var workerIP string = utilities.RetrievePublicIP()
var workerHostName string = utilities.RetrieveHostName()
var workerHostNameFriendly string
var endOfTheWorld bool
var workersMetadata workersMetadataType
var throttlerChannel chan int = make(chan int) // used to throttle the workers
var workerStatus = map[string]workerStatusType{
	"NEW":   {statusText: "NEW", statusIcon: "NEW"},
	"READY": {statusText: "READY", statusIcon: "READY"},
	"COOL":  {statusText: "COOLING DOWN", statusIcon: "COOL"},
	"STOP":  {statusText: "STOPPED", statusIcon: "STOP"},
	"UNK":   {statusText: "UNKNOWN", statusIcon: "STOP"}}

const WRKNEW = "NEW"
const WRKSTSREADY = "READY"
const WRKSTSCOOL = "COOL"
const WRKSTSSTOP = "STOP"
const STATISTICSTIMEOUT = 20 * time.Second
const WRKTHROTTLERPS = "WRK_THROTTLE_RPS"
const WRKGRACEPERIODMS = "WRK_GRACE_PERIOD_MS"
const WRKWRKSREADYTIMEOUTMS = "WRK_WRKS_READY_TIMEOUT_MS"
const WRKBUFCHSIZE = "WRK_BUF_CH_SIZE"
const WORKERREGION = "WORKER_REGION"
const WORKERSUBREGION = "WORKER_SUBREGION"
const WRKGOROUTINES = "WRK_GOROUTINES"
const WRKHTTPUSERAGENT = "WRK_HTTP_USER_AGENT"
const QUEUECONSUMERNAME = "worker"
const WRKAPIPORT = "WRK_API_PORT"

func main() {
	initapp.InitApp("WORKER")

	// check if region/subregion are valid
	utilities.FailOnError(checkRegionIsValid())

	utilities.FailOnError(queuehelper.InitQueueWorker(settings.GetSettStr(WORKERREGION), settings.GetSettStr(WORKERSUBREGION)))

	workerHostNameFriendly = initapp.RetrieveHostNameFriendly()

	// start the listener for internal status monitoring
	internalstatusmonitorapi.StartListener(settings.GetSettStr(WRKAPIPORT), initapp.GetAppRole())

	// start the beating..
	heartbeat.New(utilities.RetrieveHostName(), initapp.RetrieveHostNameFriendly(), initapp.GetAppRole(), settings.GetSettStr(WORKERREGION), settings.GetSettStr(WORKERSUBREGION), time.Second*60, dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameHeartbeats, settings.GetSettStr(WRKAPIPORT), workerIP).Start()

	printGreetings()
	httpcheck.HttpCheckDefaultUserAgent = settings.GetSettStr(WRKHTTPUSERAGENT)

	// create the context
	ctx, cfunc := context.WithCancel(context.Background())
	defer cfunc()

	// create the channel used by the queue consumer to buffer fetched messages
	ch := make(chan amqp.Delivery, settings.GetSettInt(WRKBUFCHSIZE))

	// pass the context cancel function to the close handler
	closeHandler(cfunc)

	// start the workers!
	startTheWorkers(ctx, ch)

	// check if all workers are ready to work
	allWorkersReady()

	// start the throttler
	throttler()

	// start to consume the queue...
	utilities.FailOnError(ConsumeQueueForPendingChecks(ctx, ch))

	go waitingForTheWorldToEnd(ctx)

	userInput()

}

func checkRegionIsValid() error {
	region := settings.GetSettStr(WORKERREGION)
	subRegion := settings.GetSettStr(WORKERSUBREGION)
	regions, err := settings.GetRegionsList()
	if err != nil {
		return err
	}
	// look for the region and if found look for the subregion...
	for _, r := range regions {
		if r.Id == region {
			for _, sr := range r.SubRegions {
				if sr.Id == subRegion {
					return nil // region/subregion are valid!
				}
			}
			return errors.New(fmt.Sprintf("sub region configured [%s] not valid", subRegion))
		}
	}
	return errors.New(fmt.Sprintf("region configured [%s] not valid", region))
}

// above a certain limit it is probably better to multiply by [n] the sleep and multiply by [n] the elements pushed in the channel
// otherwise there's the risk that one element at the time will create a sort of cap
func throttler() {
	duration := time.Second / settings.GetSettDuration(WRKTHROTTLERPS)
	go func() {
		for {
			throttlerChannel <- 1
			time.Sleep(duration)
		}
	}()
}

func printGreetings() {
	utilities.ClearScreen()
	headers := []string{"REGION", "SUBREGION", "HOSTNAME", "IP"}
	row := [][]string{{settings.GetSettStr(WORKERREGION), settings.GetSettStr(WORKERSUBREGION), workerHostName, workerIP}}
	utilities.PrintTable(headers, row)
}

func startTheWorkers(ctx context.Context, ch chan amqp.Delivery) {
	fmt.Printf("STARTING %d WORKERS\n", settings.GetSettInt(WRKGOROUTINES))
	// crate the metadata array for all the workers THEN starts the go routine
	// this to avoid that a go routine could update the "ready" status while we perform the append
	// being part of the bootstrap of the app it is not really a problem and if something goes wrong we can see it immediately...
	// but we could have done it a bit better and more elegantly....
	for i := 0; i < settings.GetSettInt(WRKGOROUTINES); i++ {
		workersMetadata.workerMetadata = append(workersMetadata.workerMetadata, workerMetadataType{WorkerStatus: WRKNEW, startTime: time.Now(), workerID: i})
	}
	for i := 0; i < settings.GetSettInt(WRKGOROUTINES); i++ {
		go worker(ctx, ch, i)
	}
}

func waitingForTheWorldToEnd(ctx context.Context) {
	select {
	case <-ctx.Done():
		break
		// not having a default make sure this is a blocking select/case until context is done...
	}

	// set a global flag to true to acknowledge the world is ending...
	endOfTheWorld = true

	// if we are here...The world is ending...
	// by now the consumer should have already received the message (pun intended) that the world is ending...
	// also the workers should have noticed it but we need to be sure to clear the channel before continuing...
	allWorkersGracefullyEnded()

	// Closing the queue
	queuehelper.CloseConsumerConnection()
	queuehelper.ClosePublisherConnection()

	// this is it, it has been fun!
	os.Exit(0)

}
func worker(ctx context.Context, ch <-chan amqp.Delivery, metadataIndex int) {

	var messageQueued queuehelper.CheckRecordQueued
	var err error

	workersMetadata.workerMetadata[metadataIndex].WorkerStatus = WRKSTSREADY
forloop:
	for {
		select {
		case check := <-ch:
			if len(check.Body) == 0 {
				// todo log this
				// this should never happen because we check for the same thing in the consumer
				// this is most likely caused by the queue empty or the consumer cancelled
				continue
			}
			err = check.Ack(false)
			if err != nil {
				// todo log this
				// it is possible that the connection dropped and the message was not acknowledged...
				// if this is true rabbitmq has put back the messages in the queue and they will be consumed shortly again...
				// so for the moment we just ignore this error and continue...
				continue
			}

			// make sure variable is clean/reset , not sure if the unmarshalling is replacing all keys values or not...
			messageQueued = queuehelper.CheckRecordQueued{}

			<-throttlerChannel
			workersMetadata.workerMetadata[metadataIndex].msgReceived++
			atomic.AddInt64(&workersMetadata.workersTotalMsgReceived, 1)
			workersMetadata.workerMetadata[metadataIndex].lastMsgTime = time.Now()
			err = unmarshalMessageBody(&check.Body, &messageQueued)
			if err != nil {
				// this is a very strange situation and we need to investigate for the moment print & die
				log.Println("Error unmarshalling the message from the queue!")
				log.Printf("This is the content of the message: %v", check)
				log.Println("-----------------------------SEE YOU SOON------------")
				utilities.FailOnError(err)
			}

			messageQueued.ReceivedByWorkerUnix = time.Now().Unix()
			messageQueued.WorkerHostname = workerHostName
			messageQueued.WorkerHostnameFriendly = workerHostNameFriendly

			// if the check fails make sure we try again just in case...

			for {
				messageQueued.Attempts++
				err = checks.ProcessCheckFromQueue(&messageQueued)
				if err != nil {
					workersMetadata.workerMetadata[metadataIndex].msgFailed++
					// if an error occurred stop trying....
					break
				}
				if messageQueued.RecordOutcome.Success {
					// if the check was successful we can break the loop
					break
				}
				// we don't want to try more than xx times
				if messageQueued.Attempts >= 3 {
					break
				}
				// sleep a little before trying again 😴
				time.Sleep(3 * time.Second)
			}

			messageQueued.QueuedReturnUnix = time.Now().Unix()
			messageQueued.RecordOutcome.Region = settings.GetSettStr(WORKERREGION)
			messageQueued.RecordOutcome.SubRegion = settings.GetSettStr(WORKERSUBREGION)

			jsonRecord, _ := json.Marshal(messageQueued)
			err = PublishResponseForCheckProcessed(jsonRecord)
			utilities.FailOnError(err)
		case <-ctx.Done():
			workersMetadata.workerMetadata[metadataIndex].WorkerStatus = WRKSTSCOOL
			if time.Since(workersMetadata.workerMetadata[metadataIndex].lastMsgTime) > settings.GetSettDuration(WRKGRACEPERIODMS)*time.Millisecond {
				workersMetadata.workerMetadata[metadataIndex].WorkerStatus = WRKSTSSTOP
				break forloop
			}
		default:
			time.Sleep(100 * time.Millisecond)
		} // end select case
	} // end for/loop [forloop]

}

func unmarshalMessageBody(body *[]byte, unmarshalledMessage *queuehelper.CheckRecordQueued) error {
	err := json.Unmarshal(*body, unmarshalledMessage)
	if err != nil {
		*unmarshalledMessage = queuehelper.CheckRecordQueued{}
		unmarshalledMessage.ErrorFatal = err.Error()
		return err
	}
	return nil
}

// we could have used a waitgroup....?
func allWorkersReady() {
	var readyCount int = 0
	var maxWait time.Duration = time.Millisecond * settings.GetSettDuration(WRKWRKSREADYTIMEOUTMS)
	var bootTime time.Time = time.Now()
	var percReady float32

	fmt.Println("Waiting for all workers to be ready")

	for {
		readyCount = 0
		for _, w := range workersMetadata.workerMetadata {
			if w.WorkerStatus == WRKSTSREADY {
				readyCount++
			}
		}
		if settings.GetSettInt(WRKGOROUTINES) == readyCount {
			fmt.Printf("\r✅  (it took %.3fs)\n", float64(time.Since(bootTime))/float64(time.Second))
			return
		}
		if time.Since(bootTime) > maxWait {
			utilities.FailOnError(errors.New("workers not ready"))
		}
		percReady = float32(readyCount) / float32(settings.GetSettInt(WRKGOROUTINES)) * 100
		fmt.Printf("%.2f%% (%d/%d)      \r", percReady, readyCount, settings.GetSettInt(WRKGOROUTINES))
	} // end for

	// nothing here!

}

// we could have used a waitgroup....!!
func allWorkersGracefullyEnded() bool {
	var stopped int
	// infinite loooooop
	for {
		stopped = 0
		for _, w := range workersMetadata.workerMetadata {
			if w.WorkerStatus == WRKSTSSTOP {
				stopped++
			}
		}
		if settings.GetSettInt(WRKGOROUTINES) == stopped {
			// wait an extra couple of second to give time to the statistics to refresh on screen...
			time.Sleep(time.Second * 2)
			// this is it... bye bye....
			return true
		}
	} // end for

	// nothing here!

}

func closeHandler(cfunc context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		endOfTheWorld = true
		cfunc()
	}()
}

func userInput() {
	for {
		userInput := utilities.ReadUserInput("h/q/s/enter")
		switch userInput {
		case "h":
			fmt.Println("h help, q quit, s statistics loop, enter statistics")
			break
		case "q":
			confirm := utilities.ReadUserInput("Are you sure? (yes to confirm)")
			if confirm == "yes" {
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
				// show the workers statists to have a better idea of what's going on during the cooling down..
				ShowWorkerStats(STATISTICSTIMEOUT)
			}
		case "s":
			ShowWorkerStats(STATISTICSTIMEOUT)
		case "":
			ShowWorkerStats(1 * time.Millisecond)
		} // end switch
	} // end for loop

}
