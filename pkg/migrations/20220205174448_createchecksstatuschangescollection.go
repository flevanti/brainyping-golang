package migrations

import (
	"brainyping/pkg/dbhelper"

	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//
// Please return an error if you want the migration to fail and the migration process to stop.
// Migration failed will continue to be pending ( or won't be rolled back if it was a down process)
// Don't exit, panic or try any other way to stop the process.
//
// just return a nice error
//

//
// IMPORTANT FOR SAFETY REASONS AND AVOID STUPID CONFLICTS: DO NOT CREATE EXPORTED FUNCTIONS
// (translated, create only functions that start with lowercase characters)
//

func up_20220205174448(db *mongo.Client) error {
	if dbhelper.CheckIfCollectionExists(db, dbhelper.GetDatabaseName(), dbhelper.TablenameChecksStatusChanges) {
		return nil
	}

	opts := options.CreateCollectionOptions{}
	err := dbhelper.CreateCollection(db, dbhelper.GetDatabaseName(), dbhelper.TablenameChecksStatusChanges, &opts)
	if err != nil {
		return err
	}

	return nil
}

func down_20220205174448(db *mongo.Client) error {
	err := db.Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameChecksStatusChanges).Drop(nil)
	if err != nil {
		return err
	}
	return nil
}

//
//
// DON'T TOUCH ANYTHING BEYOND THIS POINT
//
//

//
// this is adding the migration to the migration engine
//
func init() {
	bisonmigration.RegisterMigration(20220205174448, "createchecksstatuschangescollection", "*DEFAULT*", up_20220205174448, down_20220205174448)
}
