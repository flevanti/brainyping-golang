package initapp

import (
	"fmt"
	"log"
	"os"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/utilities"

	"github.com/joho/godotenv"
)

var bootTime time.Time

func InitApp() {
	bootTime = time.Now()
	importDotEnv()
	dbhelper.Connect()
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
	settings, err := dbhelper.GetSettings()
	utilities.FailOnError(err)

	for _, v := range settings {
		if _, exists := os.LookupEnv(v.Key); exists {
			fmt.Printf("Env variable [%s] already present, it won't be update with settings value\n", v.Key)
		} else {
			// create env variable only if it doesn't exist already
			err := os.Setenv(v.Key, v.Value)
			utilities.FailOnError(err)
		}
	} // end for loop

}
