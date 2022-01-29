package main

import (
	"brainyping/pkg/dbhelper"
	_ "brainyping/pkg/migrations"
	"fmt"
	"github.com/flevanti/bisonmigration"
)

func main() {
	bisonmigration.MigrationAppConfig(bisonmigration.MigrationAppDefaultDatabase, bisonmigration.MigrationAppDefaultCollection, dbhelper.GetClient())
	greetings()
	showPendingMigrations()

}

func greetings() {
	//fmt.Println("🅱🅸🆂🅾🅽 🅼🅸🅶🆁🅰🆃🅸🅾🅽")
	fmt.Println("-------------------------------------")
	fmt.Println("   B I S O N   M I G R A T I O N S   ")
	fmt.Println("-------------------------------------")

	fmt.Printf("Migrations: %d pending, %d processed, %d registered\n", bisonmigration.GetMigrationsPendingCount(), bisonmigration.GetMigrationsProcessedCount(), bisonmigration.GetMigrationsRegisteredCount())
}

func showPendingMigrations() {
	l := bisonmigration.GetMigrationsPending()
	fmt.Println("Pending migrations")
	for _, v := range l {
		fmt.Printf("%-15d%-70s%s\n", v.Sequence, v.Name, v.UniqueId)
	}
}

// here we are writing a migration tool....
// it was Christmas 2018 the last time I wrote a migration tool from scratch in PHP
// for an application using the drupal 7 framework.
// the embedded system to deploy changes was so strange and prone to conflict during merge
// that it was literally blocking deployments as soon as the dev team started to grow
// At that time I took inspiration from the Laravel migration system and it was simple and worked (still working actually!)
// very well
//
// unfortunately here we have a problem, mainly because - as part of the tech stack - we are using mongodb
// and it doesn't really offer as far as I know native scripting (like sql) to interact with the db to insert/alter records or structure.
//
// So to avoid loosing too much time with something not part of the core system we will do something dirty but hopefully keeping it isolated
// I was initially thinking of having separate files with specific naming convention like "[timestamp]_somethingmeaninful.go"
// and then looping the files and call a function "[filename]_up()' to run the migration but then I realised that being go files they will be compiled/included
// in the binary
