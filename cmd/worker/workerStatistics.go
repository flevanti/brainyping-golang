package main

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"strings"
	"time"
)

func ShowWorkerStats(ctx context.Context, ch chan amqp.Delivery) {
	type previousLoopsStatsForSpeedPurpose struct {
		totalMessages uint64
		samplingTime  time.Time
	}

	var previousLoopStats previousLoopsStatsForSpeedPurpose
	var msgPerSecondSpeed float64
	var deltaTime uint64
	var deltaMessages uint64
	var samplingInterval time.Duration = time.Millisecond * 3000 //this will ensure smoother statistics avoiding peaks and valleys....
	var endOfTheWorldMessage string
	var endOfTheWorldWorkerMessage string
	var successFailureRation float32

	//maybe we should reinvent the wheel and create a package for printing a nice table in the terminal....
	for {
		if endOfTheWorld {
			endOfTheWorldMessage = "COOLING DOWN......"
		}

		fmt.Printf("WORKERS STATS  (CONSUMER QUEUE SIZE %5d)\n", len(ch))

		for i := 0; i < WORKERSTOSTART; i++ {
			endOfTheWorldWorkerMessage = ""
			if endOfTheWorld {
				if workersMetadata.workerMetadata[i].inGracePeriod {
					endOfTheWorldWorkerMessage = endOfTheWorldWorkerMessage + "Cooling down... last message proc. " + workersMetadata.workerMetadata[i].lastMsgTime.Format(time.Stamp)
				}
				if workersMetadata.workerMetadata[i].stopped {
					endOfTheWorldWorkerMessage = endOfTheWorldWorkerMessage + " Stopped "
				}
			}
			successFailureRation = float32(workersMetadata.workerMetadata[i].msgFailed) / float32(workersMetadata.workerMetadata[i].msgReceived) * 100

			fmt.Printf("[%03d] %-6d %6d👍 👎%-13d ratio %.2f%%   %s\n", workersMetadata.workerMetadata[i].workerID, workersMetadata.workerMetadata[i].msgReceived, workersMetadata.workerMetadata[i].msgReceived-workersMetadata.workerMetadata[i].msgFailed, workersMetadata.workerMetadata[i].msgFailed, successFailureRation, endOfTheWorldWorkerMessage)
		}

		//prepare some statistics compared with the previous loop
		deltaTime = uint64(time.Since(previousLoopStats.samplingTime))                            //nanoseconds since last loop
		deltaMessages = workersMetadata.workersTotalMsgReceived - previousLoopStats.totalMessages //messages processed since last loop
		msgPerSecondSpeed = float64(deltaMessages) / float64(deltaTime) * 1000000000              //calculate messages/seconds speed...
		//store some info to be used later to calculate speed....
		if time.Since(previousLoopStats.samplingTime) > samplingInterval {
			previousLoopStats.samplingTime = time.Now()
			previousLoopStats.totalMessages = workersMetadata.workersTotalMsgReceived
		}
		fmt.Println("--------------------------")
		fmt.Printf("Total %d  (%.2f/s)     %s       \n", workersMetadata.workersTotalMsgReceived, msgPerSecondSpeed, time.Now().Format(time.Stamp))
		fmt.Println(endOfTheWorldMessage)

		//go to sleep, good boy!
		time.Sleep(time.Millisecond * 300)

		//go back to square one!
		fmt.Print(strings.Repeat("\033[F", WORKERSTOSTART+4)) //please note the +nth ... those are lines used for some other information that we need to consider....
	} //end outer infinit for loop
}
