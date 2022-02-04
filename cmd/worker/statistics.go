package main

import (
	"context"
	"fmt"
	"strconv"

	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"time"

	"github.com/streadway/amqp"
)

func ShowWorkerStats(ctx context.Context, ch chan amqp.Delivery) {
	var speedCalculator = utilities.CalculateSpeedPerSecond()
	// var successFailureRation float32
	var rows [][]string
	var row []string
	var tableHeaders = []string{"WRKID", "CHECKS", "OK", "NOK", "FAIL%", "LAST CHECK", "STATUS"}
	var md *workerMetadataType
	var failRatio string

	for {
		rows = [][]string{}
		for i := 0; i < settings.GetSettInt("WRK_GOROUTINES"); i++ {
			md = &workersMetadata.workerMetadata[i]
			if md.msgFailed > 0 {
				failRatio = fmt.Sprintf("%.1f", float64(md.msgFailed)/float64(md.msgReceived)*100)
			} else {
				failRatio = "0"
			}

			row = []string{
				strconv.Itoa(md.workerID),
				strconv.FormatInt(md.msgReceived, 10),
				strconv.FormatInt(md.msgReceived-md.msgFailed, 10),
				strconv.FormatInt(md.msgFailed, 10),
				failRatio,
				md.lastMsgTime.Format(time.Stamp),
				workerStatus[md.WorkerStatus].statusText,
			}
			rows = append(rows, row)
			// successFailureRation = float32(workersMetadata.workerMetadata[i].msgFailed) / float32(workersMetadata.workerMetadata[i].msgReceived) * 100

			// fmt.Printf("[%03d] %-6d %6d👍 👎%-13d ratio %.2f%%   %s\n", workersMetadata.workerMetadata[i].workerID, workersMetadata.workerMetadata[i].msgReceived, workersMetadata.workerMetadata[i].msgReceived-workersMetadata.workerMetadata[i].msgFailed, workersMetadata.workerMetadata[i].msgFailed, successFailureRation, endOfTheWorldWorkerMessage)
		}
		utilities.PrintTable(tableHeaders, rows)
		// prepare some statistics compared with the previous loop

		// fmt.Println("--------------------------")
		fmt.Printf("Total %d  (%.2f/s)     %s       \n", workersMetadata.workersTotalMsgReceived, speedCalculator(workersMetadata.workersTotalMsgReceived), time.Now().Format(time.Stamp))

		// go to sleep, good boy!
		time.Sleep(time.Millisecond * 300)

		utilities.ClearScreen()
	} // end outer infinit for loop
}
