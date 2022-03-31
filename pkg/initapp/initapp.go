package initapp

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

var bootTime time.Time

var buildDateUnix string = "1"
var buildDate string
var build string
var version string = "developer"
var gitHash string = "########"
var appRole string

func InitApp(appRoleParam string) {
	appRole = appRoleParam
	generateBuildInfo()
	checkIfuserWantsJustToSeetheVersion()
	bootTime = time.Now()
	importDotEnv()
	dbhelper.Connect(settings.GetSettStr(dbhelper.DBDBNAME), settings.GetSettStr(dbhelper.DBCONNSTRING))
	importSettings()
}

func generateBuildInfo() {
	var buildDateUnix64, err = strconv.ParseInt(buildDateUnix, 10, 64)
	utilities.FailOnError(err)
	buildDate = time.Unix(buildDateUnix64, 0).Format(time.RFC850)
	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(buildDateUnix64, 10)))
	build = hex.EncodeToString(hasher.Sum(nil))[:7]
}

func checkIfuserWantsJustToSeetheVersion() {
	var versionFlag = flag.Bool("version", false, "show version")
	flag.Parse()
	if *versionFlag {
		printVersion()
		os.Exit(1)
	}
}

func GetBootTime() time.Time {
	return bootTime
}

func GetAppRole() string {
	return appRole
}

func importDotEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error loading .env file", err.Error())
	}
}
func RetrieveHostNameFriendly() string {
	return settings.GetSettStr("HOSTFRIENDLY")
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

func printVersion() {
	fmt.Printf("VERSION...... %s\n", version)
	fmt.Printf("BUILD DATE... %s\n", buildDate)
	fmt.Printf("BUILD UNIX... %s\n", buildDateUnix)
	fmt.Printf("BUILD HASH... %s\n", build)
	fmt.Printf("GIT HASH..... %s", gitHash)

}

func GetAppVersion() [][]string {
	return [][]string{
		{"BUILD DATE", buildDate},
		{"VERSION", version},
		{"BUILD UNIX", buildDateUnix},
		{"BUILD HASH", build},
		{"GIT HASH", gitHash},
	}
}
