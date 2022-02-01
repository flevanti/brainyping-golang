package queuehelper

import (
	"brainyping/pkg/dbhelper"
	_ "brainyping/pkg/dotenv"
	"brainyping/pkg/utilities"
	"context"
	_ "github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"log"
	"os"
)

var queueNameRequest string = os.Getenv("QUEUENAME_REQUEST")
var queueNameResponse string = os.Getenv("QUEUENAME_RESPONSE")
var queueConsumerName string = "brainypingconsumer"
var queueBrokerConnection *amqp.Connection
var queueBrokerChannel *amqp.Channel

const PREFETCHCOUNT = 100

type CheckRecordQueued struct {
	Record                    dbhelper.CheckRecord        `bson:"record"`
	RecordOutcome             dbhelper.CheckOutcomeRecord `bson:"recordoutcome"`
	ScheduledUnix             int64                       `bson:"scheduledunix"`
	QueuedUnix                int64                       `bson:"queuedunix"`
	ReceivedByWorkerUnix      int64                       `bson:"receivedyworkerunix"`
	QueuedReturnUnix          int64                       `bson:"queuedreturnunix"`
	ReceivedByResponseHandler int64                       `bson:"receivedbyresponsehandler"`
	ErrorFatal                string                      `bson:"errorfatal"`
}

func init() {
	var err error
	queueBrokerConnection, err = amqp.Dial(os.Getenv("QUEUEURL"))
	if err != nil {
		log.Fatal(err.Error())
	}

	queueBrokerChannel, err = queueBrokerConnection.Channel()
	if err != nil {
		log.Fatal(err.Error())
	}

	//_ = queueBrokerChannel.Qos(1, 1, false)

	_, err = queueBrokerChannel.QueueDeclare(queueNameRequest, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = queueBrokerChannel.QueueDeclare(queueNameResponse, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	//prefecth is the quantity of records fetched from the queue.... it is important that they are processed and acknowledged... because they can't go back!
	//make sure that the number makes sense considering also the average number of go rountines workers and the buffered channel size...
	//basically we don't want to fetch too many messages, it could be risky and we could lose requests if the server for any reason crashed
	//on the other end we don't want that during the fetching of records the channel is starting to be empty and some workers have no work to do...
	//so ideally (in my humble opinion) considerig various numbers that are only in my mind...
	//PREFETCH = 2-3X the average speed
	err = queueBrokerChannel.Qos(1000, 0, false)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func PublishRequestForNewCheck(body []byte) error {
	//please note that the messages are published in a durable way, this is probably more useful in dev phase than prod
	//we should probably create a flag to accomodate this... 🤠.....
	err := queueBrokerChannel.Publish("",
		queueNameRequest,
		false,
		false,
		amqp.Publishing{Body: body, DeliveryMode: 2})
	return err
}

func PublishResponseForCheckProcessed(body []byte) error {
	err := queueBrokerChannel.Publish("",
		queueNameResponse,
		false,
		false,
		amqp.Publishing{Body: body, DeliveryMode: 2})
	return err
}

func ConsumeQueueForPendingChecks(ctx context.Context, ch chan<- amqp.Delivery) {
	msgs, err := queueBrokerChannel.Consume(queueNameRequest,
		queueConsumerName,
		false,
		false,
		false,
		false,
		nil)

	utilities.FailOnError(err)

	for {
		select {
		case msg := <-msgs:
			ch <- msg
		case <-ctx.Done():
			// the context was cancelled, stop working
			// cancel the consumer...
			_ = queueBrokerChannel.Cancel(queueConsumerName, false)
			return
		} //end select case
	} //end for

}

func ConsumeQueueForResponsesToChecks(ctx context.Context, ch chan<- amqp.Delivery) {
	msgs, err := queueBrokerChannel.Consume(queueNameResponse,
		queueConsumerName,
		false,
		false,
		false,
		false,
		nil)

	utilities.FailOnError(err)

	for {
		select {
		case msg := <-msgs:
			ch <- msg
		case <-ctx.Done():
			// the context was cancelled, stop working
			// cancel the consumer...
			_ = queueBrokerChannel.Cancel(queueConsumerName, false)
			return
		} //end select case
	} //end for

}
