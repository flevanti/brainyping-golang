package main

import (
	"brainyping/pkg/utilities"
	"bufio"
	"errors"
	"fmt"
	"github.com/flevanti/bisonmigration"
	"github.com/olekukonko/tablewriter"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func readUserInput(textToShow string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(textToShow, "> ")
	text, err := reader.ReadString('\n')
	utilities.FailOnError(err)

	return strings.Trim(text, " \n\t")
}

func greetings() {
	fmt.Println("-------------------------------------")
	fmt.Println("   B I S O N   M I G R A T I O N S   ")
	fmt.Println("-------------------------------------")

	fmt.Printf("Database used by the migration engine is [%s]", databaseName)
	if !bisonmigration.GetMigrationAppDatabaseExists() {
		fmt.Printf(" - It does not exist and it will be created")
	}
	fmt.Println()

	fmt.Printf("Collection used by the migration engine is [%s]", collectionName)
	if !bisonmigration.GetMigrationAppCollectionExists() {
		fmt.Printf(" - It does not exists and it will be created")
	}

	fmt.Println()
	fmt.Println()

	fmt.Printf("Migrations: %d pending, %d processed, %d registered\n", bisonmigration.GetMigrationsPendingCount(), bisonmigration.GetMigrationsProcessedCount(), bisonmigration.GetMigrationsRegisteredCount())
	fmt.Println()
}

func userInteractionJourneyStartsHere() {

	for {
		input := readUserInput("C:\\") //🪟 joke....? Dad joke?
		switch input {
		case "h", "help":
			showMainMenu()
			break
		case "q":
			fmt.Println("Bye bye")
			os.Exit(0)
		case "shopen":
			showPendingMigrations()
			break
		case "shopro":
			showProcessedMigrations()
			break
		case "shoreg":
			showRegisteredMigrations()
			break
		case "up":
			if messageIfDbNotInitialised() || messageIfDbConnectionsMissing() {
				//something is missing, break!
				break
			}
			_ = runPendingMigrations()
			break
		case "up1", "down", "down1", "downto":
			fmt.Println("Not yet implemented, sorry")
			break
		case "new":
			createNewStubMigrationFile()
			break
		case "conn":
			showConnectionsLabels()
			break
		case "dbinit":
			initialiseDb()
			break
		default:
			fmt.Println("Option unknown, please try again or `help`")
		}
	}

}

func initialiseDb() {
	if bisonmigration.CheckIfDbIsInitialised() {
		fmt.Println("Database already initialised")
		return
	}
	bisonmigration.InitialiseDatabase()
	fmt.Println("Database initialised")
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

	fmt.Println("Database connection label")
	connLabel := readUserInput("[leave blank for default connection] ")

	//check if the connection label is a shortcut for system labels...
	switch connLabel {
	case "":
		connLabel = bisonmigration.DbConnectionLabelDefault
		break
	}

	//we have a sequence, we have a migration name, file does not exists in target location...
	//let's go!
	err := createNewMigrationFile(filename, strconv.Itoa(sequence), connLabel, migrationName)
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
	var options [][]string
	options = append(options, []string{"shopen", "Show pending migrations"})
	options = append(options, []string{"shopro", "Show processed migrations"})
	options = append(options, []string{"shoreg", "Show registered migrations"})
	options = append(options, []string{"up", "process pending migrations"})
	options = append(options, []string{"up1", "process specific migration"})
	options = append(options, []string{"down", "Rollback last batch of migrations"})
	options = append(options, []string{"down1", "Rollback A specific migration"})
	options = append(options, []string{"downto", "Rollback TO a specific migration"})
	options = append(options, []string{"new", "Create a new migration file"})
	options = append(options, []string{"conn", "Show registered connections"})
	options = append(options, []string{"dbinit", "Initialise migration database"})
	options = append(options, []string{"q", "Quit"})

	printTable([]string{"CMD", "DESCRIPTION"}, options)

	//fmt.Println("\n\nMAIN MENU")
	//fmt.Println("1 show pending migrations\n2 show processed migrations\n3 show registered migrations")
	//fmt.Println("4 run pending migrations\n5 run specific migration")
	//fmt.Println("6 rollback last batch\n7 rollback ONE specific migration")
	//fmt.Println("8 rollback TO a specific migration\n9 create a new stub migration file")
	//fmt.Println("c show db connections labels\ndbinit initialise migration app database")
	//fmt.Println("q quit")
}

func showPendingMigrations() {
	l := bisonmigration.GetMigrationsPending()
	fmt.Println("Pending migrations")
	var tableData [][]string
	for _, v := range l {
		connMissing := ""
		if v.DbConnectionMissing {
			connMissing = "🔴"
		}
		tableData = append(tableData, []string{strconv.FormatInt(v.Sequence, 10), v.Name, v.UniqueId, fmt.Sprint(v.DbConnectionLabel, connMissing)})
	}
	printTable([]string{"SEQUENCE", "NAME", "UNIQUEID", "CONNECTION"}, tableData)
}

func showRegisteredMigrations() {
	l := bisonmigration.GetMigrationsRegistered()
	fmt.Println("Registered migrations")
	var pending, processedTime, batch string
	var tableData [][]string
	for _, v := range l {
		if !v.Processed {
			pending = "PENDING"
			processedTime = ""
			batch = ""
		} else {
			pending = ""
			processedTime = time.Unix(v.ProcessedTimeUnix, 0).Format(time.Stamp)
			batch = fmt.Sprintf("BATCH:%d", v.ProcessedBatch)
		}
		connMissing := ""
		if v.DbConnectionMissing {
			connMissing = "🔴"
		}
		tableData = append(tableData, []string{strconv.FormatInt(v.Sequence, 10), v.Name, v.UniqueId, fmt.Sprint(v.DbConnectionLabel, connMissing), pending, processedTime, batch})
	}
	printTable([]string{"SEQUENCE", "NAME", "UNIQUEID", "CONNECTION", "PENDING", "PROCESSED AT", "BATCH"}, tableData)

}

func showProcessedMigrations() {
	l := bisonmigration.GetMigrationsProcessed()
	fmt.Println("Processed migrations")
	var tableData [][]string
	for _, v := range l {
		tableData = append(tableData, []string{strconv.FormatInt(v.Sequence, 10), v.Name, v.UniqueId, v.DbConnectionLabel, time.Unix(v.ProcessedTimeUnix, 0).Format(time.Stamp), strconv.FormatInt(v.ProcessedBatch, 10), strconv.FormatInt(v.ProcessedTimeSpentMs, 10)})
	}
	printTable([]string{"SEQUENCE", "NAME", "UNIQUEID", "CONNECTION", "PROCESSED AT", "BATCH", "MS"}, tableData)

}

func showConnectionsLabels() {
	fmt.Println("Database connections labels registered")
	var tableData [][]string
	for _, v := range bisonmigration.GetConnectionsLabels() {
		tableData = append(tableData, []string{v.Label, v.Description})
	}
	printTable([]string{"CONNECTION LABEL", "DESCRIPTION"}, tableData)
}

func printTable(headers []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	//table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
}