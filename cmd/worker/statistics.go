package main

import (
	"fmt"
	"strconv"

	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"time"
)

func ShowWorkerStats(duration time.Duration) {
	var speedCalculator = utilities.CalculateSpeedPerSecond(time.Second * 3)
	// var successFailureRation float32
	var rows [][]string
	var row []string
	var tableHeaders = []string{"WRKID", "CHECKS", "OK", "NOK", "FAIL%", "LAST CHECK", "STATUS"}
	var md *workerMetadataType
	var failRatio string
	var startTime = time.Now()

	for {
		rows = [][]string{}
		for i := 0; i < settings.GetSettInt(WRKGOROUTINES); i++ {
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

		}
		utilities.PrintTable(tableHeaders, rows)

		fmt.Printf("Total %d  (%.2f/s)     %s       \n", workersMetadata.workersTotalMsgReceived, speedCalculator(workersMetadata.workersTotalMsgReceived), time.Now().Format(time.Stamp))

		// go to sleep, good boy!
		time.Sleep(time.Millisecond * 300)

		// check the duration before the clear screen so we leave the last refreshed statistics visible...
		// if for any reason the system is cooling down stay here until the end...
		if !endOfTheWorld && time.Since(startTime) > duration {
			break
		}

		utilities.ClearScreen()
	} // end outer infinite for loop
}
