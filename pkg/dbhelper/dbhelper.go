package dbhelper

import (
	"context"
	"time"

	"brainyping/pkg/utilities"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *mongo.Client
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
const TablenameChecksInFlight = "checks_inflight"

const DBDBNAME = "DBDBNAME"
const DBCONNSTRING = "DBCONNSTRING"

var mainDatabase string

// GetDefaultIndexModelsByCollectionName returns the list of indexes that should be present in the collection. Used for rebuilding purposes.
func GetDefaultIndexModelsByCollectionName(c string) []mongo.IndexModel {
	var idxs []mongo.IndexModel

	switch c {

	case TablenameSettings:
		idxUnique := true
		idxs = []mongo.IndexModel{{Keys: bson.D{{"key", 1}}, Options: &options.IndexOptions{Unique: &idxUnique}}}
		break
	case TablenameResponse:
		idxUnique := true
		idxs = []mongo.IndexModel{
			{Keys: bson.D{{"checkid", 1}, {"processedunix", -1}}},
			{Keys: bson.D{{"requestid", 1}}, Options: &options.IndexOptions{Unique: &idxUnique}},
			{Keys: bson.D{{"receivedresponseunix", 1}}},
		}
		break
	case TablenameChecksInFlight:
		idxUnique := true
		idxName := "uk_rid"
		idxs = []mongo.IndexModel{
			{Keys: bson.D{{"rid", 1}}, Options: &options.IndexOptions{Unique: &idxUnique, Name: &idxName}},
			{Keys: bson.D{{"checkid", 1}}},
		}
		break
	case TablenameChecks:
		idxUnique := true
		idxs = []mongo.IndexModel{
			{Keys: bson.D{{"owneruid", 1}}},
			{Keys: bson.D{{"checkid", 1}}, Options: &options.IndexOptions{Unique: &idxUnique}},
		}
		break
	case TablenameChecksStatus:
		idxUnique := true
		idxName := "uk_checkid"
		idxs = []mongo.IndexModel{
			{Keys: bson.D{{"checkid", 1}}, Options: &options.IndexOptions{Unique: &idxUnique, Name: &idxName}},
			{Keys: bson.D{{"owneruid", 1}}}}
		break
	}

	return idxs
}

func GetDatabaseName() string {
	return mainDatabase
}

func DeleteCollection(dbClient *mongo.Client, dbName string, tableName string) error {
	return dbClient.Database(dbName).Collection(tableName).Drop(context.Background())
}

func CreateCollection(dbClient *mongo.Client, dbName string, collectionName string, opts *options.CreateCollectionOptions) error {
	return dbClient.Database(dbName).CreateCollection(context.Background(), collectionName, opts)
}

func CreateIndexes(dbClient *mongo.Client, dbName string, tableName string, indexModels []mongo.IndexModel) error {
	_, err := dbClient.Database(dbName).Collection(tableName).Indexes().CreateMany(context.TODO(), indexModels)
	return err
}

func CheckIfCollectionExists(dbClient *mongo.Client, dbName string, tableName string) bool {
	for _, v := range CollectionsList(dbClient, dbName) {
		if tableName == v {
			return true
		}
	}
	return false
}

func CollectionsList(dbClient *mongo.Client, dbName string) []string {
	list, err := dbClient.Database(dbName).ListCollectionNames(context.Background(), bson.M{})
	utilities.FailOnError(err)
	return list
}

// TruncateCollection delete and recreate a collection and rebuild the indexes. Indexes are retrieved from a list, all other indexes present in the collection not in the list will be lost
func TruncateCollection(dbClient *mongo.Client, dbName string, collectionName string) error {
	// todo add logic or new function to truncate a collection and maintain the current indexes in the db, not the one configured in the code.(a real truncate)
	err := DeleteCollection(dbClient, dbName, collectionName)
	if err != nil {
		return err
	}

	err = CreateCollection(dbClient, dbName, collectionName, &options.CreateCollectionOptions{})
	if err != nil {
		return err
	}

	indexModels := GetDefaultIndexModelsByCollectionName(collectionName)
	if len(indexModels) > 0 {
		err = CreateIndexes(dbClient, dbName, collectionName, indexModels)
		if err != nil {
			return err
		}
	}
	return nil
}

func Connect(mainDatabaseLocal string, connString string) {
	if Initialised {
		return
	}
	mainDatabase = mainDatabaseLocal // database is not used in the connection string but stored for later use....
	var err error
	ctxwt, cancCtxwt := context.WithTimeout(context.Background(), time.Second*5)
	defer cancCtxwt()
	clientOptions := options.Client().ApplyURI(connString)
	client, err = mongo.Connect(ctxwt, clientOptions)
	if err != nil {
		utilities.FailOnError(err)
	}
	if err := client.Ping(ctxwt, readpref.Primary()); err != nil {
		utilities.FailOnError(err)
	}

}

func Disconnect() {
	if err := client.Disconnect(context.Background()); err != nil {
		panic(err)
	}
}

func SaveManyRecords(db string, collection string, records *[]interface{}) error {
	coll := GetClient().Database(db).Collection(collection)

	_, err := coll.InsertMany(context.Background(), *records)

	if err != nil {
		return err
	}
	return nil
}

func SaveRecord(db string, collection string, document interface{}) error {
	coll := GetClient().Database(db).Collection(collection)

	_, err := coll.InsertOne(context.Background(), document)

	if err != nil {
		return err
	}
	return nil
}

func DeleteRecordsByFieldValue(db string, collection string, field string, value interface{}) (*mongo.DeleteResult, error) {
	return GetClient().Database(db).Collection(collection).DeleteMany(context.TODO(), bson.M{field: value})
}

func GetClient() *mongo.Client {
	return client
}
