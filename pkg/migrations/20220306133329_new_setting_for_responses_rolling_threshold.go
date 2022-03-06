package migrations

import (
	"brainyping/pkg/dbhelper"
	"brainyping/pkg/settings"

	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
)

//
// Please return an error if you want the migration to fail and the migration process to stop.
// Migration failed will continue to be pending ( or won't be rolled back if it was a down process)
// Don't exit, panic or try any other way to stop the process.
//
// just return a nice error
//
var key1_20220306133329 string = "RR_DAYS"
var key2_20220306133329 string = "RR_FREQUENCY_SEC"
var key3_20220306133329 string = "RR_BATCH_SIZE"

func up_20220306133329(db *mongo.Client) error {
	// remove the keys before inserting them again
	_ = down_20220306133329(db)

	settings.SaveNewSett(dbhelper.SettingType{Key: key1_20220306133329, Value: "7", Description: "retention period for responses collection records"})
	settings.SaveNewSett(dbhelper.SettingType{Key: key2_20220306133329, Value: "900", Description: "responses retention frequency"})
	settings.SaveNewSett(dbhelper.SettingType{Key: key3_20220306133329, Value: "1000", Description: "limit the number of responses to be deleted in each db operation. Multiple operations will be executed until retention period is satisfied."})

	return nil
}

func down_20220306133329(db *mongo.Client) error {
	settings.DeleteSettingByKey(key1_20220306133329)
	settings.DeleteSettingByKey(key2_20220306133329)
	settings.DeleteSettingByKey(key3_20220306133329)

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
	bisonmigration.RegisterMigration(20220306133329, "new_setting_for_responses_rolling_threshold", "*DEFAULT*", up_20220306133329, down_20220306133329)
}
