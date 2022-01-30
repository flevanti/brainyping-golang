package main

import (
	"brainyping/pkg/utilities"
	"bufio"
	"errors"
	"fmt"
	"github.com/flevanti/bisonmigration"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func readUserInput(textToShow string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(textToShow, " >  ")
	text, err := reader.ReadString('\n')
	utilities.FailOnError(err)
	return strings.Replace(text, "\n", "", -1)
}

func greetings() {
	fmt.Println("-------------------------------------")
	fmt.Println("   B I S O N   M I G R A T I O N S   ")
	fmt.Println("-------------------------------------")

	fmt.Printf("Migrations: %d pending, %d processed, %d registered\n", bisonmigration.GetMigrationsPendingCount(), bisonmigration.GetMigrationsProcessedCount(), bisonmigration.GetMigrationsRegisteredCount())
	if bisonmigration.GetMigrationAppDatabaseExists() {
		fmt.Printf("Migration database [%s] exists", databaseName)
	} else {
		fmt.Printf("Migration database [%s] does not exist and it will be created", databaseName)
	}
	fmt.Println()
	if bisonmigration.GetMigrationAppCollectionExists() {
		fmt.Printf("Migration collection [%s] exists", collectionName)
	} else {
		fmt.Printf("Migration collection [%s] does not exist and it will be created", collectionName)
	}
	fmt.Println()

}

func userInteractionJourneyStartsHere() {

	for {
		input := readUserInput("(m -> menu)")
		switch input {
		case "m":
			showMainMenu()
			break
		case "q":
			fmt.Println("Bye bye")
			os.Exit(0)
		case "1":
			showPendingMigrations()
			break
		case "2":
			showProcessedMigrations()
			break
		case "3":
			showRegisteredMigrations()
			break
		case "9":
			createNewStubMigrationFile()
			break
		default:
			fmt.Println("Option unknown, please try again")
		}
	}

}

func createNewStubMigrationFile() {
	if !migrationsFolderExists {
		fmt.Println("unable to create new stub migration, migration folder not found")
		fmt.Println("This could happend if you are not in a DEV/IDE environment or you are running the app outside of the project root folder")
		return
	}

	sequenceDefault, _ := strconv.Atoi(time.Now().Format("20060102150405")) //YYYYMMDDHHMMSS
	sequenceUser := readUserInput(fmt.Sprintf("Sequence? [leave blank to use %d]", sequenceDefault))
	var sequence int
	if sequenceUser == "" {
		sequence = sequenceDefault
	} else {
		var err error
		sequence, err = strconv.Atoi(sequenceUser)
		if err != nil {
			fmt.Println("Sequence entered is not numeric")
			fmt.Println("Creation of stub record aborted")
			return
		}
	}

	migrationName := readUserInput(fmt.Sprintf("Migration name?"))
	if migrationName == "" {
		fmt.Println("Migration name cannot be empty")
		fmt.Println("Creation of stub record aborted")
		return
	} else {
		regx, _ := regexp.Compile("[^A-Za-z0-9]+")
		migrationName = regx.ReplaceAllString(migrationName, "_")
	}
	migrationName = strings.ToLower(migrationName)
	if len(migrationName) < 10 {
		fmt.Println("Migration name is too short, it needs to be at least 10 characters after sanitisation")
		fmt.Printf("Current migration name after sanitisation is %d [%s]\n", len(migrationName), migrationName)
		fmt.Println("Creation of stub record aborted")
		return
	}

	filename := fmt.Sprintf("%d_%s.go", sequence, migrationName)

	if checkIfMigrationFileExists(filename) {
		fmt.Println("A migration file with the same details already exists")
		fmt.Println("Creation of stub record aborted")
		return
	}

	fmt.Printf("Filename generated is [%s]\n", filename)

	accept := readUserInput("Does it look ok? (y to accept) ")
	if accept != "y" {
		fmt.Println("Creation of stub record aborted")
		return
	}

	//we have a sequence, we have a migration name, file does not exists in target location...
	//let's go!
	err := createNewMigrationFile(filename, strconv.Itoa(sequence), migrationName)
	if err != nil {
		fmt.Println("Something went wrong while creating the migration file")
		utilities.FailOnError(err)
	}
	fmt.Println("Migration file created successfully")
}

func checkIfMigrationFileExists(filename string) bool {
	_, err := os.Stat(fmt.Sprint(migrationsFilesPath, "/", filename))
	if err == nil {
		return true
	}
	if !errors.Is(err, os.ErrNotExist) {
		//something unexpected... bye bye...
		utilities.FailOnError(err)
	}
	return false
}

func showMainMenu() {
	fmt.Println("\n\nMAIN MANU")
	fmt.Println("1 show pending migrations\n2 show processed migrations\n3 show registered migrations")
	fmt.Println("4 run pending migrations\n5 run specific migration")
	fmt.Println("6 rollback last batch\n7 rollback ONE specific migration")
	fmt.Println("8 rollback TO a specific migration\n9 create a new stub migration file")
	fmt.Println("q Quit")
}

func showPendingMigrations() {
	l := bisonmigration.GetMigrationsPending()
	fmt.Println("Pending migrations")
	for _, v := range l {
		fmt.Printf("%-15d%-70s%s\n", v.Sequence, v.Name, v.UniqueId)
	}
}

func showRegisteredMigrations() {
	l := bisonmigration.GetMigrationsRegistered()
	fmt.Println("Registered migrations")
	var pending, processedTime, batch string
	for _, v := range l {
		if !v.Processed {
			pending = "❕"
			processedTime = ""
			batch = ""
		} else {
			pending = ""
			processedTime = time.Unix(v.ProcessedTimeUnix, 0).Format(time.Stamp)
			batch = fmt.Sprintf("BATCH:%d", v.ProcessedBatch)
		}
		fmt.Printf("%-15d%-70s%-15s  %-3s%-22s%s\n", v.Sequence, v.Name, v.UniqueId, pending, processedTime, batch)
	}
}

func showProcessedMigrations() {
	l := bisonmigration.GetMigrationsProcessed()
	fmt.Println("Processed migrations")
	for _, v := range l {
		fmt.Printf("%-15d%-70s%-15s%-22s%s\n", v.Sequence, v.Name, v.UniqueId, time.Unix(v.ProcessedTimeUnix, 0).Format(time.Stamp), fmt.Sprintf("BATCH:%d", v.ProcessedBatch))
	}

}
