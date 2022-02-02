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

func up_20220201220119(db *mongo.Client) error {
	if dbhelper.CheckIfTableExists(db, dbhelper.GetDatabaseName(), dbhelper.TablenameSettings) {
		return nil
	}

	opts := options.CreateCollectionOptions{}
	err := dbhelper.CreateTable(db, dbhelper.GetDatabaseName(), dbhelper.TablenameSettings, &opts)
	if err != nil {
		return err
	}

	idxUnique := true
	indexModels := []mongo.IndexModel{{
		Keys:    bson.D{{"key", 1}},
		Options: &options.IndexOptions{Unique: &idxUnique},
	}}
	err = dbhelper.CreateIndexes(db, dbhelper.GetDatabaseName(), dbhelper.TablenameSettings, indexModels)

	return nil
}

func down_20220201220119(db *mongo.Client) error {
	err := db.Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameSettings).Drop(nil)
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
//this is adding the migration to the migration engine
//
func init() {
	bisonmigration.RegisterMigration(20220201220119, "create_collection_for_centralised_configuration", "*DEFAULT*", up_20220201220119, down_20220201220119)
}
