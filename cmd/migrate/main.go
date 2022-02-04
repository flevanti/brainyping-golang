package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/initapp"
	_ "brainyping/pkg/migrations"
	"brainyping/pkg/utilities"

	"github.com/flevanti/bisonmigration"
)

const collectionName = bisonmigration.MigrationAppDefaultCollection
const migrationsFilesPath = "pkg/migrations/"

var migrationsFolderExists bool
var sequenceStrictnessFlags = []string{
	bisonmigration.SequenceStrictnessNoLateComers,
	bisonmigration.SequenceStrictnessNoDuplicates,
}

func GetDatabaseName() string {
	// use main app database by default, use a wrapper to retrieve it so it is easy to change if we want.
	return dbhelper.GetDatabaseName()
}

func main() {
	initapp.InitApp()
	bisonmigration.MigrationEngineInitialise(GetDatabaseName(), collectionName, dbhelper.GetClient(), sequenceStrictnessFlags)
	migrationsFolderExists = checkIfMigrationsFolderExists()

	registerDbConnections()

	greetings()
	showPendingMigrations()

	// this function will start the interactive menu in the terminal
	// so no other logic should be added beyond this point, it won't be processed....
	userInteractionJourneyStartsHere()
}

func registerDbConnections() {
	bisonmigration.RegisterDbConnection("main", "Main connection used by the brainyping application", dbhelper.GetClient())
	// ...
	// ...

}

func checkIfMigrationsFolderExists() bool {
	_, err := os.Stat(migrationsFilesPath)
	if err == nil {
		return true
	}
	if !errors.Is(err, os.ErrNotExist) {
		// unexpected error
		utilities.FailOnError(err)
	}

	return false

}

func createNewMigrationFile(filename, sequence, connLabel, name string) error {
	body := bisonmigration.GetMigrationFileTemplate()
	body = strings.ReplaceAll(body, "{{sequence}}", sequence)
	body = strings.ReplaceAll(body, "{{name}}", name)
	body = strings.ReplaceAll(body, "{{connLabel}}", connLabel)

	return ioutil.WriteFile(fmt.Sprint(migrationsFilesPath, "/", filename), []byte(body), 0755)

}

//
// UP FUNCTIONS
//

func runPendingMigrations() error {
	return bisonmigration.RunPendingMigrations()
}

func runSpecificMigration(uniqueId string) error {
	return bisonmigration.RunSpecificMigration(uniqueId)
}

func runNextSingleMigration() error {
	return bisonmigration.RunNextSingleMigration()
}

func runUpToSpecificMigration(uniqueId string) error {
	return bisonmigration.RunUpToSpecificMigration(uniqueId)
}

//
// DOWN FUNCTIONS
//
func rollbackLastBatchMigrations() error {
	return bisonmigration.RollbackLastBatchMigrations()
}

func rollbackSingleLastMigration() error {
	return bisonmigration.RollbackSingleLastMigration()
}

func rollbackASpecificMigration(uniqueId string) error {
	return bisonmigration.RollbackASpecificMigration(uniqueId)
}

func rollbackToSpecificMigration(uniqueId string) error {
	return bisonmigration.RollbackToSpecificMigration(uniqueId)
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
