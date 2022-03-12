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
// IMPORTANT FOR SAFETY REASONS AND AVOID STUPID CONFLICTS:
//
// DO NOT CREATE EXPORTED FUNCTIONS
// (translated, create only functions that start with lowercase characters)
//
// REMEMBER THAT ALL MIGRATIONS EXIST IN THE SAME PACKAGE, AVOID CREATING GLOBAL VARIABLES TO AVOID UNEXPECTED/HORRIBLE ERRORS
// IF YOU NEED GLOBAL VARIABLE MAKE SURE THEIR NAME IS UNIQUE, A GOOD IDEA IS TO USE THE MIGRATION SEQUENCE AS SUFFIX
// YOU HAVE BEEN WARNED

func up_20220310180244(db *mongo.Client) error {
	if dbhelper.CheckIfCollectionExists(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameHeartbeats) {
		return dbhelper.TruncateCollection(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameHeartbeats)
	}

	return dbhelper.CreateCollection(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameHeartbeats, &options.CreateCollectionOptions{})

}

func down_20220310180244(db *mongo.Client) error {
	if dbhelper.CheckIfCollectionExists(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameHeartbeats) {
		return dbhelper.DeleteCollection(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameHeartbeats)
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
	bisonmigration.RegisterMigration(20220310180244, "heartbeat_collection", "*DEFAULT*", up_20220310180244, down_20220310180244)
}
