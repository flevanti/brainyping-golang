package migrations

import (
	"context"

	"brainyping/pkg/dbhelper"

	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func up_20220130181225(db *mongo.Client) error {
	if dbhelper.CheckIfTableExists(db, dbhelper.GetDatabaseName(), dbhelper.TablenameChecks) {
		return nil
	}

	opts := options.CreateCollectionOptions{}
	err := dbhelper.CreateTable(db, dbhelper.GetDatabaseName(), dbhelper.TablenameChecks, &opts)
	if err != nil {
		return err
	}

	idxUnique := true
	indexModels := []mongo.IndexModel{
		{Keys: bson.D{{"owneruid", 1}}},
		{Keys: bson.D{{"checkid", 1}}, Options: &options.IndexOptions{Unique: &idxUnique}},
	}

	_, err = db.Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameChecks).Indexes().CreateMany(context.TODO(), indexModels)
	if err != nil {
		return err
	}
	return nil
}

func down_20220130181225(db *mongo.Client) error {
	if !dbhelper.CheckIfTableExists(db, dbhelper.GetDatabaseName(), dbhelper.TablenameChecks) {
		return nil
	}
	err := db.Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameChecks).Drop(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

//
// this is adding the migration to the migration engine
//
func init() {
	bisonmigration.RegisterMigration(20220130181225, "create_checks_collection", bisonmigration.DbConnectionLabelDefault, up_20220130181225, down_20220130181225)
}
