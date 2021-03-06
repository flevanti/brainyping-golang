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
	CheckId            string     `bson:"checkid"`
	Name               string     `bson:"name"`
	NameFriendly       string     `bson:"namefriendly"`
	Host               string     `bson:"host"`
	Port               int        `bson:"port"`
	Type               string     `bson:"type"`
	SubType            string     `bson:"subtype"`
	Frequency          int        `bson:"frequency"`
	UserAgent          string     `bson:"useragent"`
	HttpHeaders        [][]string `bson:"httpheaders"`
	HttpBody           string     `bson:"httpbody"`
	HttpStatusCodeOK   int        `bson:"httpstatuscodeok"`
	ResponseString     string     `bson:"responsestring"`
	Regions            [][]string `bson:"regions"`
	Enabled            bool       `bson:"enabled"`
	CreatedUnix        int64      `bson:"createdunix"`
	UpdatedUnix        int64      `bson:"updatedunix"`
	StartSchedTimeUnix int64      `bson:"startschedtimeunix"`
	OwnerUid           string     `bson:"owneruid"`
}

type CheckResponseRecordDb struct {
	MongoDbId              string            `bson:"_id,omitempty"`
	CheckId                string            `bson:"checkid"`
	Region                 string            `bson:"region"`
	SubRegion              string            `bson:"subregion"`
	ScheduledTimeUnix      int64             `bson:"scheduledtimeunix"`
	ScheduledTimeDelay     int64             `bson:"scheduledtimedelay"`
	QueuedRequestUnix      int64             `bson:"queuedrequestunix"`
	ReceivedByWorkerUnix   int64             `bson:"receivedbyworkerunix"`
	ProcessedUnix          int64             `bson:"processedunix"`
	TimeSpent              int64             `bson:"timespent"`
	QueuedResponseUnix     int64             `bson:"queuedresponseunix"`
	ReceivedResponseUnix   int64             `bson:"receivedresponseunix"`
	CreatedUnix            int64             `bson:"createdunix"`
	OwnerUid               string            `bson:"owneruid"`
	Success                bool              `bson:"success"`
	ErrorOriginal          string            `bson:"errororiginal"`
	ErrorFriendly          string            `bson:"errorfriendly "`
	ErrorInternal          string            `bson:"errorinternal"`
	ErrorFatal             string            `bson:"errorfatal"`
	Message                string            `bson:"message"`
	Redirects              int               `bson:"redirects"`
	RedirectsHistory       []RedirectHistory `bson:"redirectshistory"`
	RequestId              string            `bson:"requestid"`
	WorkerHostname         string            `bson:"workerhostname"`
	WorkerHostnameFriendly string            `bson:"workerhostnamefriendly"`
	Attempts               int               `bson:"attempts"`
	ContentLength          int64             `bson:"contentlength"`
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
	SubRegion        string            `bson:"subregion"`
	ContentLength    int64             `bson:"contentlength"`
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

type RegionType struct {
	Id         string          `bson:"id" json:"id"`
	Name       string          `bson:"name" json:"name"`
	Flag       string          `bson:"flag" json:"flag"`
	Enabled    bool            `bson:"enabled" json:"enabled"`
	Continent  string          `bson:"continent" json:"continent"`
	SubRegions []SubRegionType `bson:"subregions" json:"subregions"`
}

type SubRegionType struct {
	Id       string `bson:"id" json:"id"`
	Name     string `bson:"name" json:"name"`
	Provider string `bson:"provider" json:"provider"`
	Enabled  bool   `bson:"enabled" json:"enabled"`
}

const TablenameChecks = "checks"
const TablenameResponse = "responses"
const TablenameSettings = "settings"
const TablenameChecksStatus = "checks_status"
const TablenameChecksStatusChanges = "checks_status_changes"
const TablenameChecksInFlight = "checks_inflight"
const TablenameHeartbeats = "heartbeats"

const DBDBNAME = "DBDBNAME"
const DBCONNSTRING = "DBCONNSTRING"

const GLOBREGIONS = "GLOB_REGIONS"

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
	case TablenameHeartbeats:
		idxUnique := true
		idxs = []mongo.IndexModel{
			{Keys: bson.D{{"hostname", 1}, {"approle", 1}}, Options: &options.IndexOptions{Unique: &idxUnique}},
			{Keys: bson.D{{"lasthb", 1}}},
		}

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
	err := dbClient.Database(dbName).CreateCollection(context.Background(), collectionName, opts)
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

func SaveRecord(dbClient *mongo.Client, db string, collection string, document interface{}, options *options.InsertOneOptions) error {
	coll := dbClient.Database(db).Collection(collection)

	_, err := coll.InsertOne(context.Background(), document, options)

	if err != nil {
		return err
	}
	return nil
}

func UpdateRecord(dbClient *mongo.Client, db string, collection string, filter interface{}, document interface{}, options *options.UpdateOptions) error {
	coll := dbClient.Database(db).Collection(collection)

	_, err := coll.UpdateOne(context.Background(), filter, document, options)

	if err != nil {
		return err
	}
	return nil
}

// TODO change function signature to receive also the db client
func DeleteRecordsByFieldValue(db string, collection string, field string, value interface{}) (*mongo.DeleteResult, error) {
	return GetClient().Database(db).Collection(collection).DeleteMany(context.TODO(), bson.M{field: value})
}

func GetClient() *mongo.Client {
	return client
}
