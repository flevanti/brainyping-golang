package queuehelper

import (
	"log"
	"os"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/settings"

	"github.com/streadway/amqp"
)

var queueBrokerConnection *amqp.Connection
var queueBrokerChannel *amqp.Channel

const QUEUENAMEREQUEST = "QUEUENAME_REQUEST"
const QUEUENAMERESPONSE = "QUEUENAME_RESPONSE"
const QUEUEPREFETCHCOUNT = "QUEUE_PREFETCH_COUNT"

type CheckRecordQueued struct {
	Record                    dbhelper.CheckRecord        `bson:"record"`
	RecordOutcome             dbhelper.CheckOutcomeRecord `bson:"recordoutcome"`
	ScheduledUnix             int64                       `bson:"scheduledunix"`
	QueuedUnix                int64                       `bson:"queuedunix"`
	ReceivedByWorkerUnix      int64                       `bson:"receivedyworkerunix"`
	QueuedReturnUnix          int64                       `bson:"queuedreturnunix"`
	ReceivedByResponseHandler int64                       `bson:"receivedbyresponsehandler"`
	ErrorFatal                string                      `bson:"errorfatal"`
	RequestId                 string                      `bson:"requestid"`
}

func InitQueue() {
	var err error
	queueBrokerConnection, err = amqp.Dial(os.Getenv("QUEUEURL"))
	if err != nil {
		log.Fatal(err.Error())
	}

	queueBrokerChannel, err = queueBrokerConnection.Channel()
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = queueBrokerChannel.QueueDeclare(settings.GetSettStr(QUEUENAMEREQUEST), true, false, false, false, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = queueBrokerChannel.QueueDeclare(settings.GetSettStr(QUEUENAMERESPONSE), true, false, false, false, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	// prefecth is the quantity of records fetched from the queue.... it is important that they are processed and acknowledged... because they can't go back!
	// make sure that the number makes sense considering also the average number of go rountines workers and the buffered channel size...
	// basically we don't want to fetch too many messages, it could be risky and we could lose requests if the server for any reason crashed
	// on the other end we don't want that during the fetching of records the channel is starting to be empty and some workers have no work to do...
	// so ideally (in my humble opinion) considerig various numbers that are only in my mind...
	// PREFETCH = 2-3X the average speed
	err = queueBrokerChannel.Qos(settings.GetSettInt(QUEUEPREFETCHCOUNT), 0, false)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func GetQueueBrokerChannel() *amqp.Channel {
	return queueBrokerChannel
}
