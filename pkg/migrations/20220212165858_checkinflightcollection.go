package migrations

import (
	"brainyping/pkg/dbhelper"

	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/bson"
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

func up_20220212165858(db *mongo.Client) error {
	if dbhelper.CheckIfTableExists(db, dbhelper.GetDatabaseName(), dbhelper.TablenameChecksInFlight) {
		return nil
	}

	err := dbhelper.CreateTable(db, dbhelper.GetDatabaseName(), dbhelper.TablenameChecksInFlight, &options.CreateCollectionOptions{})
	if err != nil {
		return err
	}

	idxUnique := true
	idxName := "uk_rid"
	indexModels := []mongo.IndexModel{{
		Keys:    bson.D{{"rid", 1}},
		Options: &options.IndexOptions{Unique: &idxUnique, Name: &idxName},
	}}
	err = dbhelper.CreateIndexes(db, dbhelper.GetDatabaseName(), dbhelper.TablenameChecksInFlight, indexModels)
	if err != nil {
		return err
	}

	return nil

}

func down_20220212165858(db *mongo.Client) error {
	err := db.Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameChecksInFlight).Drop(nil)
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
	bisonmigration.RegisterMigration(20220212165858, "checkinflightcollection", "*DEFAULT*", up_20220212165858, down_20220212165858)
}
