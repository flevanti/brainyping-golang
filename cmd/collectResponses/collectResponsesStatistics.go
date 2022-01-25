package main

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"strings"
	"time"
)

func ShowStatistics(ctx context.Context, ch chan amqp.Delivery, saveBuffer []interface{}) {
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
	var successFailureRation float32

	//maybe we should reinvent the wheel and create a package for printing a nice table in the terminal....
	for {
		if endOfTheWorld {
			endOfTheWorldMessage = "COOLING DOWN......"
		}

		fmt.Printf("STATISTICS  (CONSUMER QUEUE BUFFER %5d   DB SAVE BUFFER %5d)\n", len(ch), len(saveBuffer))

		if endOfTheWorld {
			if metadata.inGracePeriod {
				endOfTheWorldMessage = endOfTheWorldMessage + "Cooling down... last message proc. " + metadata.lastMsgTime.Format(time.Stamp)
			}
			if metadata.stopped {
				endOfTheWorldMessage = endOfTheWorldMessage + " Stopped "
			}
		}
		successFailureRation = float32(metadata.msgFailed) / float32(metadata.msgReceived) * 100

		fmt.Printf(" %-6d %6d👍 👎%-13d ratio %.2f%%   %s\n", metadata.msgReceived, metadata.msgReceived-metadata.msgFailed, metadata.msgFailed, successFailureRation, endOfTheWorldMessage)

		//prepare some statistics compared with the previous loop
		deltaTime = uint64(time.Since(previousLoopStats.samplingTime))               //nanoseconds since last loop
		deltaMessages = metadata.msgReceived - previousLoopStats.totalMessages       //messages processed since last loop
		msgPerSecondSpeed = float64(deltaMessages) / float64(deltaTime) * 1000000000 //calculate messages/seconds speed...
		//store some info to be used later to calculate speed....
		if time.Since(previousLoopStats.samplingTime) > samplingInterval {
			previousLoopStats.samplingTime = time.Now()
			previousLoopStats.totalMessages = metadata.msgReceived
		}
		fmt.Printf("(%.2f/s)     %s       \n", msgPerSecondSpeed, time.Now().Format(time.Stamp))
		fmt.Println(endOfTheWorldMessage)

		//go to sleep, good boy!
		time.Sleep(time.Millisecond * 300)

		//go back to square one!
		fmt.Print(strings.Repeat("\033[F", 4)) //please note the +nth ... those are lines used for some other information that we need to consider....
	}
}
