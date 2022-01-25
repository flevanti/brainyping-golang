package main

import (
	"awesomeProject/pkg/dbHelper"
	_ "awesomeProject/pkg/dotEnv"
	"awesomeProject/pkg/utilities"
	"bufio"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"runtime"
	"time"
)

var dataLoadingLimit int64
var recordsToSave []interface{}

const BULKSAVEBATCHSIZE = 1000 //speed up a bit the import with multiple inserts (please note that each imported record has nth records in the db)
const RECORDSPERSECOND = 50

func parseFlags() {
	//register flags and set defaults...
	limit := flag.Int64("limit", 9999999, "# of records to dataload")

	//parse flags...
	flag.Parse()

	//assign values
	dataLoadingLimit = *limit

}

func main() {
	parseFlags()

	timeStart := time.Now()
	fmt.Println("Process started at ", timeStart.Format(time.ANSIC))
	fmt.Printf("Dataloading limit is %d records\n", dataLoadingLimit)
	fmt.Println("Current memory usage: ", utilities.GetMemoryStats("AUTO")["AllocUnit"])
	dbHelper.Connect()
	defer dbHelper.Disconnect()

	if !dbHelper.CheckIfTableExists(dbHelper.TABLENAME_CHECKS) {
		utilities.FailOnError(dbHelper.CreateTable(dbHelper.TABLENAME_CHECKS))
	} else {
		dbHelper.EmptyTable(dbHelper.TABLENAME_CHECKS)
	}

	readAndWrite()
	timeElapsed := int(time.Since(timeStart).Seconds())
	fmt.Println("Process completed at ", time.Now().Format(time.ANSIC))
	fmt.Println("Current memory usage (before GC): ", utilities.GetMemoryStats("AUTO")["AllocUnit"])
	runtime.GC()
	fmt.Println("Current memory usage (after GC): ", utilities.GetMemoryStats("AUTO")["AllocUnit"])
	fmt.Printf("It took %d seconds... \n", timeElapsed)

	//openConnection()
	//truncateTable()
	//readAndWrite()

}

func readAndWrite() {

	var linesReadFromFile int  //records processed
	var recsSaved int64        //records saved to db
	var recsInBufferList int64 //records in list to be saved
	var creationTimeUnix int64 = time.Now().Unix()
	var startSchedTimeUnix int64 = time.Now().Unix()
	var record dbHelper.CheckRecord
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
		if recordsInCurrentSecond >= RECORDSPERSECOND {
			recordsInCurrentSecond = 0
			startSchedTimeUnix--
		}
		recordsInCurrentSecond++
		linesReadFromFile++ //Lines read from the file

		//make sure record is empty
		record = dbHelper.CheckRecord{}

		//HTTPS HEAD
		record = dbHelper.CheckRecord{
			CheckId:            fmt.Sprint("RECID-", recsSaved),
			Name:               scanner.Text() + "HTTPHEAD",
			Host:               "https://" + scanner.Text(),
			Port:               443,
			Type:               "HTTP",
			SubType:            "HEAD",
			Frequency:          900,
			Enabled:            true,
			Regions:            []string{"GLOBAL"},
			RegionsEachTime:    1,
			StartSchedTimeUnix: startSchedTimeUnix,
			CreatedUnix:        creationTimeUnix,
			UpdatedUnix:        creationTimeUnix,
			OwnerUid:           "OWNER1",
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

		////HTTPS ROBOTSTXT
		//record.CheckId = fmt.Sprint("RECID-", recsSaved)
		//record.SubType = "ROBOTSTXT"
		//record.Name = scanner.Text() + record.Type + record.SubType
		//recordsToSave = append(recordsToSave, record)
		//recsSaved++
		//recsInBufferList++
		//
		////NET
		//record.CheckId = fmt.Sprint("RECID-", recsSaved)
		//record.Type = "NET"
		//record.SubType = "NET"
		//record.Name = scanner.Text() + record.Type + record.SubType
		//recordsToSave = append(recordsToSave, record)
		//recsSaved++
		//recsInBufferList++

		if recsInBufferList >= BULKSAVEBATCHSIZE {
			err := dbHelper.SaveManyRecords(&recordsToSave, dbHelper.TABLENAME_CHECKS) // pass by reference to save some memory?
			utilities.FailOnError(err)
			//cleaning up...
			recordsToSave = nil  //empty slice - save some memory!
			recsInBufferList = 0 //reset list counter
		}

		//print some information to avoid thinking we are stuck... please not the "\r"
		fmt.Printf("\r%d lines read from file, %d records saved to db ", linesReadFromFile, recsSaved)

		//check if we reached the limit of records to import.....
		if recsSaved >= dataLoadingLimit {
			break
		}
	} // for scanner

	fmt.Println()

	//make sure to flush buffered records not yet saved...
	if recsInBufferList > 0 {
		err := dbHelper.SaveManyRecords(&recordsToSave, dbHelper.TABLENAME_CHECKS)
		utilities.FailOnError(err)
		fmt.Println("Buffered records left behind flushed to db!")
		recordsToSave = nil //empty slice - save some memory!
	}

	fmt.Println("Import completed")
	// The method os.File.Close() is called
	// on the os.File object to close the file
	_ = file.Close()

}
