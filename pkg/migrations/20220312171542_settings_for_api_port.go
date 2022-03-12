package migrations

import (
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
//
// IMPORTANT FOR SAFETY REASONS AND AVOID STUPID CONFLICTS:
//
// DO NOT CREATE EXPORTED FUNCTIONS
// (translated, create only functions that start with lowercase characters)
//
// REMEMBER THAT ALL MIGRATIONS EXIST IN THE SAME PACKAGE, AVOID CREATING GLOBAL VARIABLES TO AVOID UNEXPECTED/HORRIBLE ERRORS
// IF YOU NEED GLOBAL VARIABLE MAKE SURE THEIR NAME IS UNIQUE, A GOOD IDEA IS TO USE THE MIGRATION SEQUENCE AS SUFFIX
// YOU HAVE BEEN WARNED

func up_20220312171542(db *mongo.Client) error {
	down_20220312171542(db) // remove keys before setting them to be sure they do not exist
	settings.SaveNewSettFriendly("WRK_API_PORT", "8080", "listening port for worker API")
	settings.SaveNewSettFriendly("SCH_API_PORT", "8081", "listening port for scheduler API")
	settings.SaveNewSettFriendly("RR_API_PORT", "8082", "listening port for response retention API")
	settings.SaveNewSettFriendly("RC_API_PORT", "8083", "listening port for response collection API")
	settings.SaveNewSettFriendly("STM_API_PORT", "8084", "listening port for status morning API")
	return nil
}

func down_20220312171542(db *mongo.Client) error {
	settings.DeleteSettingByKey("WRK_API_PORT")
	settings.DeleteSettingByKey("SCH_API_PORT")
	settings.DeleteSettingByKey("RR_API_PORT")
	settings.DeleteSettingByKey("RC_API_PORT")
	settings.DeleteSettingByKey("STM_API_PORT")

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
	bisonmigration.RegisterMigration(20220312171542, "settings_for_api_port", "*DEFAULT*", up_20220312171542, down_20220312171542)
}
