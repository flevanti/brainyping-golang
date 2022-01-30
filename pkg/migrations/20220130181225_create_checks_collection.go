package migrations

import (
	"brainyping/pkg/dbhelper"
	"context"
	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func up_20220130181225(db *mongo.Client) error {
	if dbhelper.CheckIfTableExists(dbhelper.TABLENAME_CHECKS) {
		return nil
	}

	opt := options.CreateCollectionOptions{}
	err := db.Database(dbhelper.DATABASE).CreateCollection(context.TODO(), dbhelper.TABLENAME_CHECKS, &opt)
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
	_, err = db.Database(dbhelper.DATABASE).Collection(dbhelper.TABLENAME_CHECKS).Indexes().CreateMany(context.TODO(), indexModels)
	if err != nil {
		return err
	}
	return nil
}

func down_20220130181225(db *mongo.Client) error {
	if !dbhelper.CheckIfTableExists(dbhelper.TABLENAME_CHECKS) {
		return nil
	}
	err := db.Database(dbhelper.DATABASE).Collection(dbhelper.TABLENAME_CHECKS).Drop(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

//
//this is adding the migration to the migration engine
//
func init() {
	bisonmigration.RegisterMigration(20220130181225, "create_checks_collection", "main", up_20220130181225, down_20220130181225)
}
