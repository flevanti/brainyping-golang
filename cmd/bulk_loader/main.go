package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
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
	initapp.InitApp("BULKLOADER")
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
	var requestsPerSecond int
	var frequencyMinutes int
	var linesReadFromFile int // records processed
	var recsSaved int         // records saved to db
	var recsInBufferList int  // records in list to be saved
	var creationTimeUnix int64 = time.Now().Unix()
	var startSchedTimeUnix int64 = time.Now().Unix()
	var record dbhelper.CheckRecord
	var recordsInCurrentSecond int

	fmt.Println("Please be sure you know in adavance how many rows will be pushed to avoid frequency overlapping")
	requestsPerSecond, err := strconv.Atoi(utilities.ReadUserInput("How many requests per seconds? "))
	utilities.FailOnError(err)
	frequencyMinutes, err = strconv.Atoi(utilities.ReadUserInput("Checks frequency (minutes)? "))
	utilities.FailOnError(err)

	// remove previous records added by the bulk loader....
	fmt.Println("Previous records removed")
	_, err = dbhelper.DeleteRecordsByFieldValue(dbhelper.GetDatabaseName(), dbhelper.TablenameChecks, "owneruid", settings.GetSettStr(BLOWNERUID))
	utilities.FailOnError(err)

	file, err := os.Open("siteslist.txt")
	if err != nil {
		log.Fatalf("failed to open file")
	}
	// create a new scanner
	scanner := bufio.NewScanner(file)
	// define the splitting function (lines)
	scanner.Split(bufio.ScanLines)

	regionsList := createRegionsList()

	for scanner.Scan() {
		if recordsInCurrentSecond >= requestsPerSecond {
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
			SubType:            "GET",
			Frequency:          frequencyMinutes,
			Enabled:            true,
			Regions:            regionsList,
			StartSchedTimeUnix: startSchedTimeUnix,
			CreatedUnix:        creationTimeUnix,
			UpdatedUnix:        creationTimeUnix,
			OwnerUid:           settings.GetSettStr(BLOWNERUID),
		}
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

func createRegionsList() [][]string {

	var regionsListToReturn [][]string
	regions, err := settings.GetRegionsList()
	utilities.FailOnError(err)

	for _, r := range regions {
		// todo check if region is enabled....
		for _, sr := range r.SubRegions {
			// todo check if subregion is enabled
			regionsListToReturn = append(regionsListToReturn, []string{r.Id, sr.Id})
		}
	}

	return regionsListToReturn
}
