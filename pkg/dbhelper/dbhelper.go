package dbhelper

import (
	"context"

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
	RequestId            string            `bson:"requestid"`
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

const DBDBNAME = "DBDBNAME"
const DBCONNSTRING = "DBCONNSTRING"

var mainDatabase string

func GetDatabaseName() string {
	return mainDatabase
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

func Connect(mainDatabaseLocal string, connString string) {
	if Initialised {
		return
	}
	mainDatabase = mainDatabaseLocal // database is not used in the connection string but stored for later use....
	var err error

	ctx = context.Background()
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

func DeleteRecordsByFieldValue(db string, collection string, field string, value interface{}) (*mongo.DeleteResult, error) {
	return GetClient().Database(db).Collection(collection).DeleteOne(context.TODO(), bson.M{field: value})
}

func GetClient() *mongo.Client {
	return client
}
