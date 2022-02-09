package dbhelper

import (
	"context"
	"fmt"
	"log"

	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *mongo.Client
var ctx context.Context
var Initialised bool

type CheckRecord struct {
	CheckId            string   `bson:"checkid"`
	Name               string   `bson:"name"`
	Host               string   `bson:"host"`
	Port               int      `bson:"port"`
	Type               string   `bson:"type"`
	SubType            string   `bson:"subtype"`
	Frequency          int      `bson:"frequency"`
	UserAgent          string   `bson:"useragent"`
	Regions            []string `bson:"regions"`
	RegionsEachTime    int      `bson:"regionseachtime"`
	Enabled            bool     `bson:"enabled"`
	CreatedUnix        int64    `bson:"createdunix"`
	UpdatedUnix        int64    `bson:"updatedunix"`
	StartSchedTimeUnix int64    `bson:"startschedtimeunix"`
	OwnerUid           string   `bson:"owneruid"`
}

type CheckResponseRecordDb struct {
	MongoDbId            string            `bson:"_id,omitempty"`
	CheckId              string            `bson:"checkid"`
	Region               string            `bson:"region"`
	ScheduledTimeUnix    int64             `bson:"scheduledtimeunix"`
	ScheduledTimeDelay   int64             `bson:"scheduledtimedelay"`
	QueuedRequestUnix    int64             `bson:"queuedrequestunix"`
	ReceivedByWorkerUnix int64             `bson:"receivedbyworkerunix"`
	ProcessedUnix        int64             `bson:"processedunix"`
	TimeSpent            int64             `bson:"timespent"`
	QueuedResponseUnix   int64             `bson:"queuedresponseunix"`
	ReceivedResponseUnix int64             `bson:"receivedresponseunix"`
	CreatedUnix          int64             `bson:"createdunix"`
	OwnerUid             string            `bson:"owneruid"`
	Success              bool              `bson:"success"`
	ErrorOriginal        string            `bson:"errororiginal"`
	ErrorFriendly        string            `bson:"errorfriendly "`
	ErrorInternal        string            `bson:"errorinternal"`
	ErrorFatal           string            `bson:"errorfatal"`
	Message              string            `bson:"message"`
	Redirects            int               `bson:"redirects"`
	RedirectsHistory     []RedirectHistory `bson:"redirectshistory"`
	RequestId            string            `bson:"requestid""`
}

type CheckOutcomeRecord struct {
	TimeSpent        int64             `bson:"timespent"`
	Success          bool              `bson:"success"`
	ErrorOriginal    string            `bson:"errororiginal"`
	ErrorFriendly    string            `bson:"errorfriendly "`
	ErrorInternal    string            `bson:"errorinternal"`
	Message          string            `bson:"message"`
	Redirects        int               `bson:"redirects"`
	RedirectsHistory []RedirectHistory `bson:"redirectshistory"`
	CreatedUnix      int64             `bson:"createdunix"`
	Region           string            `bson:"region"`
}

type CheckStatusChangeRecordDb struct {
	CheckId             string
	OwnerUid            string
	ChangeUnix          int
	ChangePreviousUnix  int
	Status              string
	StatusPrevious      string
	RequestId           string
	ResponseDbId        string
	ChangeProcessedUnix int
}

type RedirectHistory struct {
	URL        string `bson:"url"`
	Status     string `bson:"status"`
	StatusCode int    `bson:"statuscode"`
}

type SettingType struct {
	Key         string `bosn:"key"`
	Value       string `bson:"value"`
	Description string `bson:"description"`
}

const TablenameChecks = "checks"
const TablenameResponse = "responses"
const TablenameSettings = "settings"
const TablenameChecksStatus = "checks_status"
const TablenameChecksStatusChanges = "checks_status_changes"

func GetDatabaseName() string {
	return settings.GetSettStr("DBDBNAME")
}

func GetSettings() ([]SettingType, error) {
	var result SettingType
	var results []SettingType
	var err error
	cursor, err := GetClient().Database(GetDatabaseName()).Collection(TablenameSettings).Find(nil, bson.M{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {

		err = cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	} // end for loop

	return results, nil
}

func DeleteTable(dbClient *mongo.Client, dbName string, tableName string) error {
	return dbClient.Database(dbName).Collection(tableName).Drop(ctx)
}

func CreateTable(dbClient *mongo.Client, dbName string, tableName string, opts *options.CreateCollectionOptions) error {
	return dbClient.Database(dbName).CreateCollection(ctx, tableName, opts)
}

func CreateIndexes(dbClient *mongo.Client, dbName string, tableName string, indexModels []mongo.IndexModel) error {
	_, err := dbClient.Database(dbName).Collection(tableName).Indexes().CreateMany(context.TODO(), indexModels)
	return err
}

func CheckIfTableExists(dbClient *mongo.Client, dbName string, tableName string) bool {
	for _, v := range TablesList(dbClient, dbName) {
		if tableName == v {
			return true
		}
	}
	return false
}

func TablesList(dbClient *mongo.Client, dbName string) []string {
	list, err := dbClient.Database(dbName).ListCollectionNames(ctx, bson.M{})
	utilities.FailOnError(err)
	return list
}

func CountEnabledChecks() int64 {
	coll := GetClient().Database("brainyping").Collection("checks")

	count, err := coll.CountDocuments(context.TODO(), bson.M{"enabled": true})
	_ = count
	if err != nil {
		log.Fatalf("OOOUCH " + err.Error())
	}
	// convert the cursor result to bson
	return count
}

func Connect() {
	if Initialised {
		return
	}
	var err error

	ctx = context.Background()
	connString := settings.GetSettStr("DBCONNSTRING")
	clientOptions := options.Client().ApplyURI(connString)

	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		utilities.FailOnError(err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

}

func Disconnect() {
	if err := client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func SaveManyRecords(db string, collection string, records *[]interface{}) error {
	coll := GetClient().Database(db).Collection(collection)

	_, err := coll.InsertMany(ctx, *records)

	if err != nil {
		return err
	}
	return nil
}

func SaveRecord(db string, collection string, document interface{}) error {
	coll := GetClient().Database(db).Collection(collection)

	_, err := coll.InsertOne(ctx, document)

	if err != nil {
		return err
	}
	return nil
}

func SaveNewSett(record SettingType) {
	err := SaveRecord(GetDatabaseName(), TablenameSettings, record)
	utilities.FailOnError(err)
}

func DeleteSettingByKey(key string) {
	_, err := DeleteRecordsByFieldValue(GetDatabaseName(), TablenameSettings, "key", key)
	utilities.FailOnError(err)
}

func DeleteRecordsByFieldValue(db string, collection string, field string, value interface{}) (*mongo.DeleteResult, error) {
	return GetClient().Database(db).Collection(collection).DeleteOne(context.TODO(), bson.M{field: value})
}

func GetRecords() {
	fmt.Println("reading records!!!!")
	coll := GetClient().Database("brainyping").Collection("checks")
	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatalf("OOOUCH " + err.Error())
	}
	// convert the cursor result to bson
	var result bson.M
	var i int64
	for cursor.Next(ctx) {
		i++
		_ = cursor.Decode(&result)
		fmt.Print("(", i, ")", result["name"], "  ---   ")
	}

}

func RetrieveEnabledChecksToBeScheduled(ch chan CheckRecord) {
	coll := GetClient().Database("brainyping").Collection("checks")
	opts := options.Find().SetProjection(bson.D{
		{"checkid", 1},
		{"name", 1},
		{"host", 1},
		{"port", 1},
		{"type", 1},
		{"subtype", 1},
		{"frequency", 1},
		{"regions", 1},
		{"regionseachtime", 1},
		{"owneruid", 1},
		{"startschedtimeunix", 1},
	})
	cursor, err := coll.Find(context.TODO(), bson.M{"enabled": true}, opts)
	if err != nil {
		log.Fatalf("OOOUCH " + err.Error())
	}
	// convert the cursor result to bson
	var result CheckRecord
	var i int64
	for cursor.Next(ctx) {
		i++
		err = cursor.Decode(&result)
		utilities.FailOnError(err)
		ch <- result
	}
	// tell caller we are done here....
	close(ch)
	return
}

func GetClient() *mongo.Client {
	return client
}
