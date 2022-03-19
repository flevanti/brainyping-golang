package migrations

import (
	"encoding/json"

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
//
// IMPORTANT FOR SAFETY REASONS AND AVOID STUPID CONFLICTS:
//
// DO NOT CREATE EXPORTED FUNCTIONS
// (translated, create only functions that start with lowercase characters)
//
// REMEMBER THAT ALL MIGRATIONS EXIST IN THE SAME PACKAGE, AVOID CREATING GLOBAL VARIABLES TO AVOID UNEXPECTED/HORRIBLE ERRORS
// IF YOU NEED GLOBAL VARIABLE MAKE SURE THEIR NAME IS UNIQUE, A GOOD IDEA IS TO USE THE MIGRATION SEQUENCE AS SUFFIX
// YOU HAVE BEEN WARNED

func up_20220317174954(db *mongo.Client) error {
	var regions = []dbhelper.RegionType{
		{
			Name:       "Germany",
			Id:         "germany",
			Enabled:    true,
			Flag:       "🇩🇪",
			Continent:  "EU",
			SubRegions: []dbhelper.SubRegionType{{Id: "dusseldorf", Name: "Dusseldorf", Enabled: true, Provider: ""}},
		},
		{
			Name:       "United Kingdom",
			Id:         "unitedkingdom",
			Enabled:    true,
			Flag:       "🇬🇧",
			Continent:  "EU",
			SubRegions: []dbhelper.SubRegionType{{Id: "london", Name: "London", Enabled: true, Provider: ""}},
		},
		{
			Name:       "New Zealand",
			Id:         "newzealand",
			Enabled:    true,
			Flag:       "🇳🇿",
			Continent:  "OC",
			SubRegions: []dbhelper.SubRegionType{{Id: "auckland", Name: "Auckland", Enabled: true, Provider: ""}},
		},
		{
			Name:       "India",
			Id:         "india",
			Enabled:    true,
			Flag:       "🇮🇳",
			Continent:  "AS",
			SubRegions: []dbhelper.SubRegionType{{Id: "mumbai", Name: "Mumbai", Enabled: true, Provider: ""}},
		},
		{
			Name:       "United States",
			Id:         "unitedstates",
			Enabled:    true,
			Flag:       "🇺🇸",
			Continent:  "NA",
			SubRegions: []dbhelper.SubRegionType{{Id: "oregon", Name: "Oregon", Enabled: true, Provider: ""}, {Id: "virginia", Name: "Virginia", Enabled: true, Provider: ""}},
		},
	}

	regionsJson, err := json.Marshal(regions)
	if err != nil {
		return err
	}
	settings.SaveNewSettFriendly(dbhelper.GLOBREGIONS, string(regionsJson), "Application regions")

	return nil
}

func down_20220317174954(db *mongo.Client) error {
	settings.DeleteSettingByKey(dbhelper.GLOBREGIONS)
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
	bisonmigration.RegisterMigration(20220317174954, "add_regions_settings", "*DEFAULT*", up_20220317174954, down_20220317174954)
}
