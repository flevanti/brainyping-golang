package initapp

import (
	"fmt"
	"log"
	"os"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

var bootTime time.Time

func InitApp() {
	bootTime = time.Now()
	importDotEnv()
	dbhelper.Connect(settings.GetSettStr(dbhelper.DBDBNAME), settings.GetSettStr(dbhelper.DBCONNSTRING))
	importSettings()
}

func GetBootTime() time.Time {
	return bootTime
}

func importDotEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error loading .env file", err.Error())
	}
}

func importSettings() {
	setts, err := GetSettings()
	utilities.FailOnError(err)

	for _, v := range setts {
		if _, exists := os.LookupEnv(v.Key); exists {
			fmt.Printf("Env variable [%s] already present, it won't be update with settings value\n", v.Key)
		} else {
			// create env variable only if it doesn't exist already
			err := os.Setenv(v.Key, v.Value)
			utilities.FailOnError(err)
		}
	} // end for loop

}

func GetSettings() ([]dbhelper.SettingType, error) {
	var result dbhelper.SettingType
	var results []dbhelper.SettingType
	var err error
	cursor, err := dbhelper.GetClient().Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameSettings).Find(nil, bson.M{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(nil) {

		err = cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	} // end for loop

	return results, nil
}
