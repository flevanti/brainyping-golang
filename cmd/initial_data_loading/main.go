package main

import (
	"brainyping/pkg/dbhelper"
	_ "brainyping/pkg/dotenv"
	"brainyping/pkg/utilities"
	"bufio"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"runtime"
	"time"
)

var recordsToSave []interface{}

const BULKSAVEBATCHSIZE int = 1000 //speed up a bit the import with multiple inserts (please note that each imported record has nth records in the db)
const RECORDSPERSECOND int = 10    //this will determine
const OWNERUID string = "INIT-DATA-LOAD"

func main() {

	timeStart := time.Now()
	fmt.Println("Process started at ", timeStart.Format(time.ANSIC))
	fmt.Println("Current memory usage: ", utilities.GetMemoryStats("AUTO")["AllocUnit"])
	dbhelper.Connect()
	defer dbhelper.Disconnect()

	if !dbhelper.CheckIfTableExists(dbhelper.TABLENAME_CHECKS) {
		utilities.FailOnError(dbhelper.CreateTable(dbhelper.TABLENAME_CHECKS))
	} else {
		dbhelper.DeleteAllChecksByOwnerUid(OWNERUID)
	}

	readAndWrite()
	timeElapsed := int(time.Since(timeStart).Seconds())
	fmt.Println("Process completed at ", time.Now().Format(time.ANSIC))
	fmt.Println("Current memory usage (before GC): ", utilities.GetMemoryStats("AUTO")["AllocUnit"])
	runtime.GC()
	fmt.Println("Current memory usage (after GC): ", utilities.GetMemoryStats("AUTO")["AllocUnit"])
	fmt.Printf("It took %d seconds... \n", timeElapsed)

}

func readAndWrite() {

	var linesReadFromFile int //records processed
	var recsSaved int         //records saved to db
	var recsInBufferList int  //records in list to be saved
	var creationTimeUnix int64 = time.Now().Unix()
	var startSchedTimeUnix int64 = time.Now().Unix()
	var record dbhelper.CheckRecord
	var recordsInCurrentSecond int

	file, err := os.Open("siteslist.txt")
	if err != nil {
		log.Fatalf("failed to open file")
	}
	//create a new scanner
	scanner := bufio.NewScanner(file)
	//define the splitting function (lines)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if recordsInCurrentSecond >= RECORDSPERSECOND/2 {
			recordsInCurrentSecond = 0
			startSchedTimeUnix--
		}
		recordsInCurrentSecond++
		linesReadFromFile++ //Lines read from the file

		//make sure record is empty
		record = dbhelper.CheckRecord{}

		//HTTPS HEAD
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
			OwnerUid:           OWNERUID,
		}
		record.Name = scanner.Text() + record.Type + record.SubType
		recordsToSave = append(recordsToSave, record)
		recsSaved++
		recsInBufferList++

		//HTTPS GET
		record.CheckId = fmt.Sprint("RECID-", recsSaved)
		record.SubType = "GET"
		record.Name = scanner.Text() + record.Type + record.SubType
		recordsToSave = append(recordsToSave, record)
		recsSaved++
		recsInBufferList++

		if recsInBufferList >= BULKSAVEBATCHSIZE {
			err := dbhelper.SaveManyRecords(&recordsToSave, dbhelper.TABLENAME_CHECKS) // pass by reference to save some memory?
			utilities.FailOnError(err)
			//cleaning up...
			recordsToSave = nil  //empty slice - save some memory!
			recsInBufferList = 0 //reset list counter
		}

		//print some information to avoid thinking we are stuck... please not the "\r"
		fmt.Printf("\r%d lines read from file, %d records saved to db ", linesReadFromFile, recsSaved)

	} // for scanner

	fmt.Println()

	//make sure to flush buffered records not yet saved...
	if recsInBufferList > 0 {
		err := dbhelper.SaveManyRecords(&recordsToSave, dbhelper.TABLENAME_CHECKS)
		utilities.FailOnError(err)
		fmt.Println("Buffered records left behind flushed to db!")
		recordsToSave = nil //empty slice - save some memory!
	}

	fmt.Println("Import completed")
	// The method os.File.Close() is called
	// on the os.File object to close the file
	_ = file.Close()

}
