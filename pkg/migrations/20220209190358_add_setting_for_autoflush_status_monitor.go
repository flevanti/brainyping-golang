package migrations

import (
	"brainyping/pkg/dbhelper"

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
var key string = "STM_SAVE_AUTO_FLUSH_MS"
var value string = "5000"
var descr string = "status monitor max wait between flushing the buffer to db"

func up_20220209190358(db *mongo.Client) error {
	dbhelper.DeleteSettingByKey(key)
	dbhelper.SaveNewSett(dbhelper.SettingType{Key: key, Value: value, Description: descr})
	return nil
}

func down_20220209190358(db *mongo.Client) error {
	dbhelper.DeleteSettingByKey(key)
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
	bisonmigration.RegisterMigration(20220209190358, "add_setting_for_autoflush_status_monitor", "*DEFAULT*", up_20220209190358, down_20220209190358)
}
