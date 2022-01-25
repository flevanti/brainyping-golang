package main

import (
	"awesomeProject/pkg/dbHelper"
	"awesomeProject/pkg/queueHelper"
	"awesomeProject/pkg/utilities"
	"encoding/json"
	"fmt"
	"github.com/go-co-op/gocron"
	"log"
	"time"
)

var scheduler *gocron.Scheduler
var spreadTimeWindow int64 = 900
var bootTime time.Time = time.Now()
var jobsQueuedSinceBoot int64

func main() {
	fmt.Println("SCHEDULER")
	fmt.Printf("Boot time is %s\n", bootTime.Format(time.Stamp))

	//count enabled checks to plan
	count := dbHelper.CountEnabledChecks()
	if count == 0 {
		fmt.Println("No checks to plan, bye bye.... ")
		return
	}

	fmt.Println(count, " checks with enabled status")
	fmt.Println(spreadTimeWindow, " seconds used to spread the load")

	scheduler = gocron.NewScheduler(time.UTC)
	//enforce uniqueness of tags that we are using as a way to retrieve a scheduled job later...
	scheduler.TagsUnique()

	//put each check in the scheduler
	scheduleChecks()
	//start the scheduler
	startScheduler()

	//done, show some statistics.... forever!
	ShowMemoryStatsWhileSchedulerIsRunning()

}

func startScheduler() {
	doneSignal := make(chan int)
	go waitForSchedulerToStart(doneSignal)
	scheduler.StartAsync()
	doneSignal <- 1 //this should stop the go routing waiting for the scheduler to start...
	close(doneSignal)
}

func scheduleChecks() {
	var recScheduledTotal int64
	var record dbHelper.CheckRecord
	var recordQueued queueHelper.CheckRecordQueued
	var err error
	var printLine = func(rec int64, memAlloc string) {
		fmt.Printf("Checks scheduled %d (mem. %s)            \r", rec, memAlloc)
	}

	chRecords := make(chan dbHelper.CheckRecord)
	go dbHelper.RetrieveEnabledChecksToBeScheduled(chRecords)

	for record = range chRecords {
		recScheduledTotal++
		//add start time to the record to have a point of reference for future checks (and be able to reference a planned scheduled time instead of the time the check occurs)
		recordQueued = queueHelper.CheckRecordQueued{Record: record}
		_, err = scheduler.Every(record.Frequency).Second().StartAt(time.Unix(record.StartSchedTimeUnix, 0)).Tag(record.CheckId).Do(queue, recordQueued)
		utilities.FailOnError(err)
	} //end ch range
	printLine(recScheduledTotal, utilities.GetMemoryStats("MB")["AllocUnit"])
	fmt.Println()

}

func waitForSchedulerToStart(doneSignal <-chan int) {
	var startedWaiting time.Time = time.Now()
	var spinner = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	var spinnerPosition int
	for {
		select {
		case <-doneSignal:
			fmt.Println("\rStarting scheduler ✅        ")
			fmt.Printf("Starting the scheduler took %s\n\n", time.Since(startedWaiting)/time.Second*time.Second)
			return
		default:
			spinnerPosition++
			fmt.Printf("\rStarting scheduler %s     ", spinner[spinnerPosition%len(spinner)])
			time.Sleep(time.Millisecond * 150)
		} //end select case
		//nothing to do here I guess!
	} //end for
}

func queue(record queueHelper.CheckRecordQueued) {
	record.QueuedUnix = time.Now().Unix()
	record.ScheduledUnix = time.Now().Unix()
	var recordJson, err = json.Marshal(record)
	if err != nil {
		fmt.Println("🔴")
		log.Fatal(err)
	}

	//for the moment we queue the whole record scheduled,
	//maybe later down the line we want to slim down...or enrich?
	err = queueHelper.PublishRequestForNewCheck(recordJson)
	if err != nil {
		log.Fatal(err)
	}

	jobsQueuedSinceBoot++
}

func addScheduledJob() {

}

func ShowMemoryStatsWhileSchedulerIsRunning() {
	for {

		memoryStats := utilities.GetMemoryStats("MB")
		fmt.Printf("SCHED JOBS %d JOBS QUEUED %d MALLOC %s MALLOCTOT %s GC %s   (Uptime %s)\033[0K\r",
			scheduler.Len(),
			jobsQueuedSinceBoot,
			memoryStats["AllocUnit"],
			memoryStats["TotalAllocUnit"],
			memoryStats["NumGC"],
			time.Since(bootTime)/time.Second*time.Second)

		time.Sleep(time.Second * 3)
	}
}
