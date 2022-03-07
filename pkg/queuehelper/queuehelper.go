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

	err = queueBrokerChannel.Qos(settings.GetSettInt(QUEUEPREFETCHCOUNT), 0, false)

	if err != nil {
		log.Fatal(err.Error())
	}
}

func GetQueueBrokerChannel() *amqp.Channel {
	return queueBrokerChannel
}

func GetQueueBrokerConnection() *amqp.Connection {
	return queueBrokerConnection
}

func CloseQueue() {
	_ = queueBrokerChannel.Close()
	_ = queueBrokerConnection.Close()
}
