package main

import (
	"errors"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/utilities"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func retrieveMarkerFromStatusChanges() {
	type recordType struct {
		Id        string `bson:"responsedbid"`
		RequestId string `bson:"requestid"`
	}

	var record recordType
	findOptions := options.FindOne()
	findOptions.SetSort(bson.M{"_id": -1}) // reverse order on _id to get the last one
	err := dbhelper.GetClient().Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameChecksStatusChanges).FindOne(nil, bson.M{}, findOptions).Decode(&record)

	// if no error it means we have found what we are looking for... assign value and return...
	if err == nil {
		marker.ResponseDbId = record.Id
		marker.RequestId = record.RequestId
		marker.source = markerSourceStatusChanges
		return
	}

	// if we are here we had an error....

	// make sure if we have an error IT IS NOT the "no document found" error....
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		utilities.FailOnError(err)
	}

}

func retrieveMarkerFromResponses() {
	type recordType struct {
		Id        string `bson:"_id"`
		RequestId string `bson:"requestid"`
	}

	var record recordType
	findOptions := options.FindOne()
	// so this error is actually ok, it means we couldn't find a marker, is it the first time we run this!?
	// so let's go back in time to the first response received ... and start from there....
	findOptions.SetSort(bson.M{"_id": 1}) // normal order on _id to get he first one...
	err := dbhelper.GetClient().Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameResponse).FindOne(nil, bson.M{}, findOptions).Decode(&record)
	// if no error it means we have found what we are looking for... assign value and return...
	if err == nil {
		marker.ResponseDbId = record.Id
		marker.RequestId = record.RequestId
		marker.source = markerSourceResponses
		return
	}

	// make sure if we have an error IT IS NOT the "no document found" error....
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		utilities.FailOnError(err)
	}

}

func readResponsesFromMarker(ch chan dbhelper.CheckResponseRecordDb) {
	var firstLoop = true
	var record dbhelper.CheckResponseRecordDb
	var err error
	var cursor *mongo.Cursor
	var objectId primitive.ObjectID
	var filterOperator string
	var recsProcessed int64
	for {
		if firstLoop && marker.source == markerSourceResponses {
			firstLoop = false
			filterOperator = "$gte"
		} else {
			filterOperator = "$gt"
		}
		objectId, err = primitive.ObjectIDFromHex(marker.ResponseDbId)
		utilities.FailOnError(err)
		findOptions := options.Find()
		findOptions.SetSort(bson.M{"_id": 1})
		findOptions.SetLimit(100)

		cursor, err = dbhelper.GetClient().Database(dbhelper.GetDatabaseName()).Collection(dbhelper.TablenameResponse).Find(nil, bson.M{"_id": bson.M{filterOperator: objectId}}, findOptions)
		utilities.FailOnError(err)

		for cursor.Next(nil) {
			recsProcessed++
			err = cursor.Decode(&record)
			utilities.FailOnError(err)
			ch <- record
			marker.RequestId = record.RequestId
			marker.ResponseDbId = record.MongoDbId
		} // end for cursor loop...

	} // end infinite for loop

}
