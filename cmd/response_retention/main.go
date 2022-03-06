package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/initapp"
	"brainyping/pkg/settings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var retentionDays int64
var retentionSeconds int64
var frequencySeconds int64
var batchSizeLimit int64
var totRecordsRemovedGlobal int64

const RRDAYS string = "RR_DAYS"
const RRFREQUENCYSEC string = "RR_FREQUENCY_SEC"
const RRBATCHSIZE string = "RR_BATCH_SIZE"

func main() {
	initapp.InitApp()
	retentionDays = settings.GetSettInt64(RRDAYS)
	retentionSeconds = retentionDays * 24 * 60 * 60
	frequencySeconds = settings.GetSettInt64(RRFREQUENCYSEC)
	batchSizeLimit = settings.GetSettInt64(RRBATCHSIZE)

	for {
		clean()
		wait()
		log.Println("--------------------------------------------------")
	}
}

// remove the response records older than the retention period
// records are removed from the older to the most recent
// records are removed in batches, each batch is max RR_BATCH_SIZE records (more or less)
// Mongodb doesn't support limit on delete so this is the approach we take:
// perform a query with...
// filters only records older than the retention period
// order them by receivedresponseunix field ASC (older on top)
// skip the first RR_BATCH_SIZE-1 records
// limit the query to 1 record
// the retrieved record should be the nth older from the bottom
// delete ALL records where receivedresponseunix >= the receivedresponseunix value found with the query
// this means that potentially we are removing few more records than the RR_BATCH_SIZE if more records have the same `receivedresponseunix` value
// to avoid this we could technically us the monbo _id value to delete the records once identified the one with the query because mongo _ids are sortable....
// todo check how things are going in the future if using mongo _id could be a better choice

func clean() {
	var record dbhelper.CheckResponseRecordDb
	var totRecordsRemoved int64

	log.Println("CLEANING \U0001F9F9")
	log.Printf("[retention [%d] days, batch size [%d] records, frequency [%d] seconds]\n", retentionDays, batchSizeLimit, frequencySeconds)

	unixThreshold := time.Now().Unix() - retentionSeconds

	log.Printf("Threshold is %s\n", time.Unix(unixThreshold, 0).Format(time.RFC850))

	// filter only records older than the treshold
	filter := bson.M{"receivedresponseunix": bson.M{"$lte": unixThreshold}}

	// sort by receivedresponseunix ASC
	// skip the first RR_BATCH_SIZE records
	// grab only the field we need receivedresponseunix
	opts := options.FindOne()
	opts.Sort = bson.M{"receivedresponseunix": 1}
	opts.Skip = &batchSizeLimit
	opts.Projection = bson.M{"receivedresponseunix": 1}

	for {

		err := dbhelper.GetClient().Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameResponse).FindOne(context.Background(), filter, opts).Decode(&record)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				break
			}
			log.Printf("ERROR RETRIEVING: %s", err.Error())
			// for the moment ... just return....
			return
		}

		log.Printf("Identified record [%s] with [receivedresponseunix] field value [%d] converted to [%s]\n", record.MongoDbId, record.ReceivedResponseUnix, time.Unix(record.ReceivedResponseUnix, 0).Format(time.RFC850))

		delRes, err := dbhelper.DeleteRecordsByFieldValue(dbhelper.GetDatabaseName(), dbhelper.TablenameResponse, "receivedresponseunix", bson.M{"$lte": record.ReceivedResponseUnix})
		if err != nil {
			log.Printf("ERROR DELETING: %s", err.Error())
			// for the moment ... just return....
			return
		}
		log.Printf("Records removed %d", delRes.DeletedCount)
		totRecordsRemoved += delRes.DeletedCount

	} // for loop

	if totRecordsRemoved == 0 {
		log.Println("Looks like no records have been removed this time, see you soon!")
	} else {
		log.Printf("Total records removed in this cleaning session %d\n", totRecordsRemoved)
		totRecordsRemovedGlobal += totRecordsRemoved
	}
	log.Printf("Total records removed since app started %d\n", totRecordsRemovedGlobal)

}

func wait() {
	var countDown = frequencySeconds
	for range time.Tick(time.Second) {
		countDown--
		fmt.Printf("Next retention cleaning in %d seconds     \r", countDown)
		if countDown <= 0 {
			break
		}
	}
	fmt.Println("")

}
