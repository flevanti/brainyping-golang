package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/initapp"
	"brainyping/pkg/settings"
	_ "brainyping/pkg/settings"
	"brainyping/pkg/utilities"
)

var recordsToSave []interface{}

const BLOWNERUID = "BL_OWNERUID"
const BLRPSSPREAD = "BL_RPS_SPREAD"
const BLBULKSAVESIZE = "BL_BULK_SAVE_SIZE"

func main() {
	initapp.InitApp()
	timeStart := time.Now()
	fmt.Println("Process started at ", timeStart.Format(time.ANSIC))
	fmt.Println("Current memory usage: ", utilities.GetMemoryStats("AUTO")["AllocUnit"])
	dbhelper.Connect(settings.GetSettStr(dbhelper.DBDBNAME), settings.GetSettStr(dbhelper.DBCONNSTRING))
	defer dbhelper.Disconnect()

	readAndWrite()
	timeElapsed := int(time.Since(timeStart).Seconds())
	fmt.Println("Process completed at ", time.Now().Format(time.ANSIC))
	fmt.Println("Current memory usage (before GC): ", utilities.GetMemoryStats("AUTO")["AllocUnit"])
	runtime.GC()
	fmt.Println("Current memory usage (after GC): ", utilities.GetMemoryStats("AUTO")["AllocUnit"])
	fmt.Printf("It took %d seconds... \n", timeElapsed)

}

func readAndWrite() {

	var linesReadFromFile int // records processed
	var recsSaved int         // records saved to db
	var recsInBufferList int  // records in list to be saved
	var creationTimeUnix int64 = time.Now().Unix()
	var startSchedTimeUnix int64 = time.Now().Unix()
	var record dbhelper.CheckRecord
	var recordsInCurrentSecond int

	file, err := os.Open("siteslist.txt")
	if err != nil {
		log.Fatalf("failed to open file")
	}
	// create a new scanner
	scanner := bufio.NewScanner(file)
	// define the splitting function (lines)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if recordsInCurrentSecond >= settings.GetSettInt(BLRPSSPREAD)/2 {
			recordsInCurrentSecond = 0
			startSchedTimeUnix--
		}
		recordsInCurrentSecond++
		linesReadFromFile++ // Lines read from the file

		// make sure record is empty
		record = dbhelper.CheckRecord{}

		// HTTPS HEAD
		record = dbhelper.CheckRecord{
			CheckId:            fmt.Sprint("RECID-", recsSaved),
			Host:               "https://" + scanner.Text(),
			Port:               443,
			Type:               "HTTP",
			SubType:            "HEAD",
			Frequency:          4,
			Enabled:            true,
			Regions:            []string{"GLOBAL"},
			RegionsEachTime:    1,
			StartSchedTimeUnix: startSchedTimeUnix,
			CreatedUnix:        creationTimeUnix,
			UpdatedUnix:        creationTimeUnix,
			OwnerUid:           settings.GetSettStr(BLOWNERUID),
		}
		record.Name = scanner.Text() + record.Type + record.SubType
		recordsToSave = append(recordsToSave, record)
		recsSaved++
		recsInBufferList++

		// HTTPS GET
		record.CheckId = fmt.Sprint("RECID-", recsSaved)
		record.SubType = "GET"
		record.Name = scanner.Text() + record.Type + record.SubType
		recordsToSave = append(recordsToSave, record)
		recsSaved++
		recsInBufferList++

		if recsInBufferList >= settings.GetSettInt(BLBULKSAVESIZE) {
			err := dbhelper.SaveManyRecords(dbhelper.GetDatabaseName(), dbhelper.TablenameChecks, &recordsToSave) // pass by reference to save some memory?
			utilities.FailOnError(err)
			// cleaning up...
			recordsToSave = nil  // empty slice - save some memory!
			recsInBufferList = 0 // reset list counter
		}

		// print some information to avoid thinking we are stuck... please not the "\r"
		fmt.Printf("\r%d lines read from file, %d records saved to db ", linesReadFromFile, recsSaved)

	} // for scanner

	fmt.Println()

	// make sure to flush buffered records not yet saved...
	if recsInBufferList > 0 {
		err := dbhelper.SaveManyRecords(dbhelper.GetDatabaseName(), dbhelper.TablenameChecks, &recordsToSave)
		utilities.FailOnError(err)
		fmt.Println("Buffered records left behind flushed to db!")
		recordsToSave = nil // empty slice - save some memory!
	}

	fmt.Println("Import completed")
	// The method os.File.Close() is called
	// on the os.File object to close the file
	_ = file.Close()

}
