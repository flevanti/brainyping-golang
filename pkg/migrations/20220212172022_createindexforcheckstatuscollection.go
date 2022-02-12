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

func up_20220212172022(db *mongo.Client) error {

	idxUnique := true
	idxName := "uk_checkid"
	indexModels := []mongo.IndexModel{{
		Keys:    bson.D{{"checkid", 1}},
		Options: &options.IndexOptions{Unique: &idxUnique, Name: &idxName},
	}}
	err := dbhelper.CreateIndexes(db, dbhelper.GetDatabaseName(), dbhelper.TablenameChecksStatus, indexModels)
	if err != nil {
		return err
	}
	return nil
}

func down_20220212172022(db *mongo.Client) error {
	_, err := db.Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameChecksStatus).Indexes().DropOne(nil, "uk_checkid")

	return err
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
	bisonmigration.RegisterMigration(20220212172022, "createindexforcheckstatuscollection", "*DEFAULT*", up_20220212172022, down_20220212172022)
}
