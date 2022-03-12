package main

import (
	"log"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/queuehelper"
	"brainyping/pkg/utilities"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CountEnabledChecks() int64 {
	coll := dbhelper.GetClient().Database("brainyping").Collection("checks")

	count, err := coll.CountDocuments(nil, bson.M{"enabled": true})
	_ = count
	if err != nil {
		log.Fatalf("OOOUCH " + err.Error())
	}
	// convert the cursor result to bson
	return count
}

func RetrieveEnabledChecksToBeScheduled(ch chan dbhelper.CheckRecord) {
	coll := dbhelper.GetClient().Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameChecks)
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
	cursor, err := coll.Find(nil, bson.M{"enabled": true}, opts)
	if err != nil {
		log.Fatalf("OOOUCH " + err.Error())
	}
	// convert the cursor result to bson
	var result dbhelper.CheckRecord
	var i int64
	for cursor.Next(nil) {
		i++
		err = cursor.Decode(&result)
		utilities.FailOnError(err)
		ch <- result
	}
	// tell caller we are done here....
	close(ch)
	return
}

func saveRecordAsInFlight(record queuehelper.CheckRecordQueued) error {

	var recToSave struct {
		CheckId           string `bson:"checkid"`
		Rid               string `bson:"rid"`
		InFlightSinceUnix int64  `bson:"inflightsinceunix"`
		InFlightSince     string `bson:"inflightsince"`
	}

	recToSave.Rid = record.RequestId
	recToSave.CheckId = record.Record.CheckId
	recToSave.InFlightSinceUnix = record.QueuedUnix
	recToSave.InFlightSince = time.Unix(record.QueuedUnix, 0).Format(time.Stamp)

	var recToSaveI interface{} = recToSave

	return dbhelper.SaveRecord(dbhelper.GetClient(), dbhelper.GetDatabaseName(), dbhelper.TablenameChecksInFlight, recToSaveI, &options.InsertOneOptions{})

}
