package dbhelper

import (
	_ "brainyping/pkg/dotenv"
	"brainyping/pkg/utilities"
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
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
	Regions            []string `bson:"regions"`
	RegionsEachTime    int      `bson:"regionseachtime"`
	Enabled            bool     `bson:"enabled"`
	CreatedUnix        int64    `bson:"createdunix"`
	UpdatedUnix        int64    `bson:"updatedunix"`
	StartSchedTimeUnix int64    `bson:"startschedtimeunix"`
	OwnerUid           string   `bson:"owneruid"`
}

type CheckResponseRecordDb struct {
	CheckId string `bson:"checkid"`

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
	//UpdatedUnix          int64             `bson:"updatedunix"`
	//Name                 string            `bson:"name"`
	//Host                 string            `bson:"host"`
	//Port                 int               `bson:"port"`
	//Type                 string            `bson:"type"`
	//SubType              string            `bson:"subtype"`
	//Frequency            int               `bson:"frequency"`
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

func init() {
	Connect()
}

const DATABASE = "brainyping"
const TABLENAME_CHECKS = "checks"
const TABLENAME_RESPONSES = "responses"
const TABLENAME_HB = "heartbeats"

func EmptyTable(t string) {
	if !CheckIfTableExists(t) {
		utilities.FailOnError(errors.New("Unable to empty table [" + t + "], table not found"))
	}
	//mongo does not have a "truncate" command and the common approach is to delete everything, it could be slow... so we remove and recreate the collection
	//delete collection
	err := DeleteTable(t)
	utilities.FailOnError(err)

	//create collection
	err = CreateTable(t)
	utilities.FailOnError(err)

	//TOO SLOW...WHEN COLLECTION IS BIG
	//_, err := client.Database(DATABASE).Collection(t).DeleteMany(ctx, bson.M{})

	utilities.FailOnError(err)
}

func DeleteAllChecksByOwnerUid(ownerUid string) {
	d, err := client.Database(DATABASE).Collection(TABLENAME_CHECKS).DeleteMany(ctx, bson.M{"owneruid": bson.M{"$eq": ownerUid}})
	_ = d
	utilities.FailOnError(err)
}

func DeleteTable(t string) error {
	return client.Database(DATABASE).Collection(t).Drop(ctx)
}

func CreateTable(t string) error {
	return client.Database(DATABASE).CreateCollection(ctx, t)
}

func CheckIfTableExists(t string) bool {
	for _, v := range TablesList() {
		if t == v {
			return true
		}
	}
	return false
}

func TablesList() []string {
	list, err := client.Database(DATABASE).ListCollectionNames(ctx, bson.M{})
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
	var dbUser = os.Getenv("DBUSER")
	var dbPass = os.Getenv("DBPASS")
	var dbUrl = os.Getenv("DBURL")
	var dbProtDnsSeedFlag = os.Getenv("DBPROTOCOLWITHDNSSEED")
	var dbProtocol string

	if dbProtDnsSeedFlag == "1" {
		dbProtocol = "mongodb+srv://"
	} else {
		dbProtocol = "mongodb://"
	}

	ctx = context.Background()
	//var dbDbName = os.Getenv("DBDBNAME")
	clientOptions := options.Client().
		ApplyURI(dbProtocol + dbUser + ":" + dbPass + "@" + dbUrl + "/?retryWrites=true&w=majority")

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

func SaveManyRecords(records *[]interface{}, table string) error {
	coll := GetClient().Database(DATABASE).Collection(table)

	_, err := coll.InsertMany(ctx, *records)
	if err != nil {
		return err
	}
	return nil
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
	//tell caller we are done here....
	close(ch)
	return
}

func GetClient() *mongo.Client {
	return client
}