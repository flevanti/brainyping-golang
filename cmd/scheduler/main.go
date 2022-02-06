package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/initapp"
	"brainyping/pkg/queuehelper"
	_ "brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
)

var scheduler *gocron.Scheduler
var jobsQueuedSinceBoot int64
var jobsNotQueuedBecausePaused int64
var schedulerPaused bool // this is not interacting with the scheduler directly but preventing it to push new cheduled jobs in the queue to be processed

func main() {
	initapp.InitApp()
	queuehelper.InitQueue()
	fmt.Println("SCHEDULER")
	fmt.Printf("Boot time is %s\n", initapp.GetBootTime().Format(time.Stamp))

	// count enabled checks to plan
	count := dbhelper.CountEnabledChecks()
	if count == 0 {
		fmt.Println("No checks to plan, bye bye.... ")
		return
	}

	fmt.Println(count, " checks with enabled status")

	scheduler = gocron.NewScheduler(time.UTC)
	// enforce uniqueness of tags that we are using as a way to retrieve a scheduled job later...
	scheduler.TagsUnique()

	// put each check in the scheduler
	scheduleChecks()

	// wait for all jobs to be started before accepting scheduled jobs
	schedulerPaused = true
	// start the scheduler
	startScheduler()
	schedulerPaused = false

	// done, show some statistics.... forever!
	ShowMemoryStatsWhileSchedulerIsRunning()

}

func startScheduler() {
	doneSignal := make(chan int)
	go waitForSchedulerToStart(doneSignal)
	scheduler.StartAsync()
	doneSignal <- 1 // this should stop the go routing waiting for the scheduler to start...
	close(doneSignal)
}

func scheduleChecks() {
	var recScheduledTotal int64
	var record dbhelper.CheckRecord
	var recordQueued queuehelper.CheckRecordQueued
	var err error
	var printLine = func(rec int64, memAlloc string) {
		fmt.Printf("Checks scheduled %d (mem. %s)            \r", rec, memAlloc)
	}

	chRecords := make(chan dbhelper.CheckRecord)
	fmt.Println("Adding records to scheduler")
	go dbhelper.RetrieveEnabledChecksToBeScheduled(chRecords)

	for record = range chRecords {
		recScheduledTotal++
		// add start time to the record to have a point of reference for future checks (and be able to reference a planned scheduled time instead of the time the check occurs)
		recordQueued = queuehelper.CheckRecordQueued{Record: record}
		_, err = scheduler.Every(record.Frequency).Minute().StartAt(time.Unix(record.StartSchedTimeUnix, 0)).Tag(record.CheckId).Do(queue, recordQueued)
		utilities.FailOnError(err)
		printLine(recScheduledTotal, utilities.GetMemoryStats("MB")["AllocUnit"])

	} // end ch range
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
		} // end select case
		// nothing to do here I guess!
	} // end for
}

func queue(record queuehelper.CheckRecordQueued) {
	if schedulerPaused {
		atomic.AddInt64(&jobsNotQueuedBecausePaused, 1)
		return
	}
	atomic.AddInt64(&jobsQueuedSinceBoot, 1)
	record.RequestId = fmt.Sprintf("%d--%s", time.Now().UnixNano(), uuid.NewString())
	record.QueuedUnix = time.Now().Unix()
	record.ScheduledUnix = record.QueuedUnix // use the same time as the queued time, we don't have a better alternative right now.
	var recordJson, err = json.Marshal(record)
	if err != nil {
		fmt.Println("🔴")
		log.Fatal(err)
	}
	// TODO SEND THE SCHEDULED EVENT ALSO TO THE SCHEDULE PLAN...

	// for the moment we queue the whole record scheduled,
	// maybe later down the line we want to slim down...or enrich?
	err = queuehelper.PublishRequestForNewCheck(recordJson)
	if err != nil {
		log.Fatal(err)
	}

}

// Ok so this is tricky and probably not necessary.
// Unless the scheduled time is exactly the current timestamp 🍾
// we will consider the scheduled time the nearest one in the past
// we are basically assuming that we are a little behind schedule... not ahead....
//
//
// CURRENTLY DEPRECATED TO FIND A BETTER APPROACH OR IMPROVE THIS....
func calculateScheduledTime(startedAt *int64, frequency *int64) int64 {

	currentTimeUnix := time.Now().Unix()

	// add to the initial start time as many "frequency" as calculated dividing the difference in seconds between the start time and now
	// basically ... start time is 100, frequency is 20, current time is 230, the nearest scheduled time in the past is 220.

	// oooh boy so many *****
	return *startedAt + *frequency*((currentTimeUnix-*startedAt) / *frequency)
}

func ShowMemoryStatsWhileSchedulerIsRunning() {
	for {
		fmt.Print("\033[H\033[2J")
		memoryStats := utilities.GetMemoryStats("MB")
		if schedulerPaused {
			fmt.Printf("SCHEDULER IS PAUSED 🟠 - MISSED JOBS %d\n", jobsNotQueuedBecausePaused)
		} else {
			fmt.Println("SCHEDULER IS ACTIVE 🟢")
		}
		fmt.Printf("JOBS IN SCHEDULER %d JOBS QUEUED SO FAR %d MALLOC %s GC %s   (Uptime %s)",
			scheduler.Len(),
			jobsQueuedSinceBoot,
			memoryStats["AllocUnit"],
			memoryStats["NumGC"],
			time.Since(initapp.GetBootTime())/time.Second*time.Second)

		time.Sleep(time.Second * 1)
	}
}
