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

const databaseName = dbhelper.DATABASE //keep the migrations collection in the main db
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
	bisonmigration.RegisterDbConnection("main", dbhelper.GetClient())
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

func createNewMigrationFile(filename, sequence, name string) error {
	body := template
	body = strings.ReplaceAll(body, "{{sequence}}", sequence)
	body = strings.ReplaceAll(body, "{{name}}", name)

	return ioutil.WriteFile(fmt.Sprint(migrationsFilesPath, "/", filename), []byte(body), 0755)

}

func runPendingMigratoins() error {
	bisonmigration.RunPendingMigratoins()
	return nil
}
