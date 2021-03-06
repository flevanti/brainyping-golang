package migrations

import (
	"context"

	"brainyping/pkg/dbhelper"

	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func up_20220130181238(db *mongo.Client) error {
	if dbhelper.CheckIfCollectionExists(db, dbhelper.GetDatabaseName(), dbhelper.TablenameResponse) {
		return nil
	}

	opt := options.CreateCollectionOptions{}
	err := db.Database(dbhelper.GetDatabaseName()).CreateCollection(context.TODO(), dbhelper.TablenameResponse, &opt)
	if err != nil {
		return err
	}
	indexModels := []mongo.IndexModel{{
		Keys: bson.D{
			{"owneruid", 1},
		}},
		{Keys: bson.D{
			{"checkid", 1},
		}},
	}
	err = dbhelper.CreateIndexes(db, dbhelper.GetDatabaseName(), dbhelper.TablenameResponse, indexModels)

	if err != nil {
		return err
	}
	return nil
}

func down_20220130181238(db *mongo.Client) error {
	if !dbhelper.CheckIfCollectionExists(db, dbhelper.GetDatabaseName(), dbhelper.TablenameResponse) {
		return nil
	}
	err := db.Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameResponse).Drop(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

//
// this is adding the migration to the migration engine
//
func init() {
	bisonmigration.RegisterMigration(20220130181238, "create_responses_collection", bisonmigration.DbConnectionLabelDefault, up_20220130181238, down_20220130181238)
}
