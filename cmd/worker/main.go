package main

import (
	"brainyping/pkg/checks"
	"brainyping/pkg/queuehelper"
	"brainyping/pkg/utilities"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type workerMetadataType struct {
	startTime     time.Time
	msgReceived   int64
	msgFailed     int64
	workerID      int
	workerReady   bool
	lastMsgTime   time.Time
	stopped       bool
	inGracePeriod bool
}
type workersMetadataType struct {
	workerMetadata          []workerMetadataType
	workersTotalMsgReceived uint64
}

var workerRegion string = os.Getenv("WORKER_REGION")
var workerSubRegion string = os.Getenv("WORKER_SUBREGION")
var workerIP string = utilities.RetrievePublicIP()
var workerHostName string = utilities.RetrieveHostName()
var endOfTheWorld bool
var workersMetadata workersMetadataType
var throttlerChannel chan int

const WORKERSTOSTART int = 50
const BUFFEREDCHANNELSIZE int = 100
const THROTTLERLIMIT int = 100
const GRACEPERIOD time.Duration = time.Second * 20

func main() {
	printGreetings()

	//initialise the throttle channel, please note that the variable already exists in the global scope
	throttlerChannel = make(chan int)

	//create the context
	ctx, cfunc := context.WithCancel(context.Background())
	defer cfunc()

	//create the channel used by the queue consumer to buffer fetched messages
	ch := make(chan amqp.Delivery, BUFFEREDCHANNELSIZE)

	//pass the context cancel function to the close hanlder
	closeHandler(cfunc)

	//start the workers!
	startTheWorkers(ctx, ch)

	//check if all workers are ready to work
	allWorkersReady()

	//start the throttler
	throttler()

	//start the queue consumer...
	go queuehelper.ConsumeQueueForPendingChecks(ctx, ch)

	//show some statistics about the current session
	go ShowWorkerStats(ctx, ch)

	//forever!!!
	waitingForTheWorldToEnd(ctx)

}

//above a certain limit it is probably better to multiply by [n] the sleep and multiply by [n] the elements pushed in the channel
//otherwise there's the risk that one element at the time will create a sort of cap
func throttler() {
	duration := time.Second / time.Duration(THROTTLERLIMIT)
	go func() {
		for {
			throttlerChannel <- 1
			time.Sleep(duration)
		}
	}()
}

func printGreetings() {
	fmt.Print("\033[H\033[2J") //clear the screen...
	fmt.Println("WORKER")
	fmt.Printf("region [%s] subregion [%s] hostname [%s] IP [%s]\n",
		workerRegion,
		workerSubRegion,
		workerHostName,
		workerIP)
	fmt.Printf("%d WORKERS CONFIGURED - THROTTLING SET TO % REQUESTS/SECOND\n", WORKERSTOSTART, THROTTLERLIMIT)
}

func startTheWorkers(ctx context.Context, ch chan amqp.Delivery) {
	fmt.Printf("STARTING %d WORKERS\n", WORKERSTOSTART)
	//crate the metadata array for all the workers THEN starts the go routine
	//this to avoid that a go routine could update the "ready" status while we perform the append
	//being part of the bootstrap of the app it is not really a problem and if something goes wrong we can see it immediately...
	//but we could have done it a bit better and more elegantly....
	for i := 0; i < WORKERSTOSTART; i++ {
		workersMetadata.workerMetadata = append(workersMetadata.workerMetadata, workerMetadataType{startTime: time.Now(), workerID: i})
	}
	for i := 0; i < WORKERSTOSTART; i++ {
		go worker(ctx, ch, i)
	}
}

func waitingForTheWorldToEnd(ctx context.Context) {
	select {
	case <-ctx.Done():
		break
		//not having a default make sure this is a blocking select/case until context is done...
	}

	//set a global flag to true to acknowledge the world is ending...
	endOfTheWorld = true

	//if we are here...The world is ending...
	//by now the consumer should have already received the message (pun intended) that the world is ending...
	//also the workers should have noticed it but we need to be sure to clear the channel before continuing...
	allWorkersGracefullyEnded()

	//this is it, it has been fun!
	os.Exit(0)

}
func worker(ctx context.Context, ch <-chan amqp.Delivery, metadataIndex int) {

	var messageQueued queuehelper.CheckRecordQueued
	var err error

	workersMetadata.workerMetadata[metadataIndex].workerReady = true
forloop:
	for {
		select {
		case check := <-ch:
			<-throttlerChannel
			workersMetadata.workerMetadata[metadataIndex].msgReceived++
			atomic.AddUint64(&workersMetadata.workersTotalMsgReceived, 1)
			workersMetadata.workerMetadata[metadataIndex].lastMsgTime = time.Now()

			err = unmarshalMessageBody(&check.Body, &messageQueued)
			utilities.FailOnError(err)

			messageQueued.ReceivedByWorkerUnix = time.Now().Unix()

			_ = check.Ack(false)
			err = checks.ProcessCheck(&messageQueued)
			if err != nil {
				workersMetadata.workerMetadata[metadataIndex].msgFailed++
			}
			messageQueued.QueuedReturnUnix = time.Now().Unix()
			messageQueued.RecordOutcome.Region = workerRegion

			jsonRecord, _ := json.Marshal(messageQueued)
			_ = queuehelper.PublishResponseForCheckProcessed(jsonRecord)

		case <-ctx.Done():
			workersMetadata.workerMetadata[metadataIndex].inGracePeriod = true
			if time.Since(workersMetadata.workerMetadata[metadataIndex].lastMsgTime) > GRACEPERIOD {
				workersMetadata.workerMetadata[metadataIndex].stopped = true
				break forloop
			}
		default:

		} //end select case
	} //end for/loop [forloop]

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

//we could have used a waitgroup....?
func allWorkersReady() {
	var ready int = 0
	var maxWait time.Duration = time.Second * 30
	var bootTime time.Time = time.Now()
	var percReady float32

	fmt.Println("CHECKING IF WORKERS ARE READY")

	//infinite loooooop
	for {
		ready = 0
		for _, w := range workersMetadata.workerMetadata {
			if w.workerReady {
				ready++
			}
		}
		if WORKERSTOSTART == ready {
			fmt.Printf("\r✅  (it took %.3fs)\n", float64(time.Since(bootTime))/float64(time.Second))
			os.Exit(0)
			return
		}
		if time.Since(bootTime) > maxWait {
			utilities.FailOnError(errors.New("workers not ready"))
		}
		percReady = float32(ready) / float32(WORKERSTOSTART) * 100
		time.Sleep(time.Millisecond * 300)
		fmt.Printf("%.2f%% (%d/%d)      \r", percReady, ready, WORKERSTOSTART)
	} // end for

	//nothing here!

}

//we could have used a waitgroup....!!
func allWorkersGracefullyEnded() bool {
	var stopped int = 0

	//infinite loooooop
	for {
		stopped = 0
		for _, w := range workersMetadata.workerMetadata {
			if w.stopped {
				stopped++
			}
		}
		if WORKERSTOSTART == stopped {
			//wait an extra couple of second to give time to the statistics to refresh on screen...
			time.Sleep(time.Second * 2)
			//this is it... bye bye....
			return true
		}
	} // end for

	//nothing here!

}

func closeHandler(cfunc context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cfunc()
	}()
}
