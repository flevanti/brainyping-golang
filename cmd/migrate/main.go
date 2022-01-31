package main

import (
	"brainyping/pkg/dbhelper"
	_ "brainyping/pkg/migrations"
	"brainyping/pkg/utilities"
	"errors"
	"fmt"
	"github.com/flevanti/bisonmigration"
	"io/ioutil"
	"os"
	"strings"
)

const databaseName = dbhelper.Database //keep the migrations collection in the main app db
const collectionName = bisonmigration.MigrationAppDefaultCollection
const migrationsFilesPath = "pkg/migrations/"

var migrationsFolderExists bool
var sequenceStrictnessFlags = []string{
	bisonmigration.SequenceStrictnessNoLateComers,
	bisonmigration.SequenceStrictnessNoDuplicates,
}

func main() {

	bisonmigration.MigrationEngineInitialise(databaseName, collectionName, dbhelper.GetClient(), sequenceStrictnessFlags)
	migrationsFolderExists = checkIfMigrationsFolderExists()

	registerDbConnections()

	greetings()
	showPendingMigrations()

	//this function will start the interactive menu in the terminal
	//so no other logic should be added beyond this point, it won't be processed....
	userInteractionJourneyStartsHere()
}

func registerDbConnections() {
	bisonmigration.RegisterDbConnection("*MAIN*", "Main connection used by the application", dbhelper.GetClient())
	//...
	//...

}

func checkIfMigrationsFolderExists() bool {
	_, err := os.Stat(migrationsFilesPath)
	if err == nil {
		return true
	}
	if !errors.Is(err, os.ErrNotExist) {
		//unexpected error
		utilities.FailOnError(err)
	}

	return false

}

func createNewMigrationFile(filename, sequence, connLabel, name string) error {
	body := template
	body = strings.ReplaceAll(body, "{{sequence}}", sequence)
	body = strings.ReplaceAll(body, "{{name}}", name)
	body = strings.ReplaceAll(body, "{{connLabel}}", connLabel)

	return ioutil.WriteFile(fmt.Sprint(migrationsFilesPath, "/", filename), []byte(body), 0755)

}

func runPendingMigrations() error {
	_ = bisonmigration.RunPendingMigrations()
	return nil
}

func messageIfDbNotInitialised() bool {
	if !bisonmigration.CheckIfDbIsInitialised() {
		fmt.Println("Database not initialised, unable to proceed")
		return true
	}
	return false
}

func messageIfDbConnectionsMissing() bool {
	if len(bisonmigration.GetDbConnectionsMissing()) > 0 {
		fmt.Println("Database connection required for pending migrations has not been registered")
		fmt.Println("Unable to proceed")
		return true
	}
	return false
}
