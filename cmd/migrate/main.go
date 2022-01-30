package main

import (
	"brainyping/pkg/dbhelper"
	_ "brainyping/pkg/migrations"
	"fmt"
	"github.com/flevanti/bisonmigration"
)

const databaseName = bisonmigration.MigrationAppDefaultDatabase
const collectionName = bisonmigration.MigrationAppDefaultCollection

var sequenceStrictnessFlags = []string{
	bisonmigration.SequenceStrictnessNoLateComers,
	bisonmigration.SequenceStrictnessNoDuplicates,
}

func main() {

	bisonmigration.MigrationEngineInitialise(databaseName, collectionName, dbhelper.GetClient(), sequenceStrictnessFlags)
	greetings()
	showPendingMigrations()

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

func showPendingMigrations() {
	l := bisonmigration.GetMigrationsPending()
	fmt.Println("Pending migrations")
	for _, v := range l {
		fmt.Printf("%-15d%-70s%s\n", v.Sequence, v.Name, v.UniqueId)
	}
}

func showOptions() {
	//show a simple menu for interaction...
}
