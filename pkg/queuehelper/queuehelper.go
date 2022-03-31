package queuehelper

// In this queueHelper package we try to abstract the connection,initialisation,binding of the queues.
// We are also trying to decide - based on the need - what and how should be configured
// The application only has two queue types "requests" and "responses"
// We can then be the consumer or the publisher of a queue
// it is probably an unneeded layer of complexity but it is easy to remove it if the app grows
// at that point we will choose simplicity over resource usage
// The application has a connection pool of two connection, one used to publish messages and one used to consume messages
// To each connection we can "attach" as many queues as we want so there's no need for the moment to have a bigger connection pool
// For this reason and for simpliciy-sake connections have dedicated variables.
//
// connection granularity (which queues to use and how) is an implementation used to learn the behaviour and try to save resources
// dedicated connection for consumer and publisher is suggested in rabbitMQ documentation
//
// we are also trying to implement a connection monitoring with reconnection feature
// for this reason we need to keep inside the package the queue consumer channel to be able to refresh it upon reconnection
//

import (
	"context"
	"sync"

	"brainyping/pkg/dbhelper"

	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/streadway/amqp"

	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"
)

type connectionInfoType struct {
	queueBrokerConnection   *amqp.Connection
	queueBrokerChannel      *amqp.Channel
	region                  string
	subRegion               string
	needRequestsQueue       bool
	needResponseQueue       bool
	allRequestsQueuesNeeded bool
	prefetchCount           int
	queuesDeclared          [][]string
	initialised             bool
	isConsumer              bool
	isPublisher             bool
	connectionFailure       bool
	channelMutex            sync.Mutex
}

var connectionConsumerInfo connectionInfoType
var connectionPublisherInfo connectionInfoType

var ConsumerChannelExposed = make(chan amqp.Delivery, 1)

// var connectionMonitoring bool

const QUEUENAMEREQUEST = "QUEUENAME_REQUEST"
const QUEUENAMERESPONSE = "QUEUENAME_RESPONSE"
const QUEUEPREFETCHCOUNT = "QUEUE_PREFETCH_COUNT"

type CheckRecordQueued struct {
	Record                    dbhelper.CheckRecord        `bson:"record"`
	RecordOutcome             dbhelper.CheckOutcomeRecord `bson:"recordoutcome"`
	ScheduledUnix             int64                       `bson:"scheduledunix"`
	QueuedUnix                int64                       `bson:"queuedunix"`
	ReceivedByWorkerUnix      int64                       `bson:"receivedyworkerunix"`
	WorkerHostname            string                      `bson:"workerhostname"`
	WorkerHostnameFriendly    string                      `bson:"workerhostnamefriendly"`
	QueuedReturnUnix          int64                       `bson:"queuedreturnunix"`
	ReceivedByResponseHandler int64                       `bson:"receivedbyresponsehandler"`
	ErrorFatal                string                      `bson:"errorfatal"`
	RequestId                 string                      `bson:"requestid"`
}

func InitQueueWorker(region string, subRegion string) error {
	if connectionConsumerInfo.initialised {
		return errors.New("consumer connection already initialised")
	}
	if connectionConsumerInfo.initialised {
		return errors.New("publisher connection already initialised")
	}
	region = strings.Trim(region, " ")
	subRegion = strings.Trim(subRegion, " ")

	if region == "" || subRegion == "" {
		utilities.FailOnError(errors.New("both region and subregion values need to be populated"))
	}

	connectionConsumerInfo = connectionInfoType{}
	connectionConsumerInfo.region = region
	connectionConsumerInfo.subRegion = subRegion
	connectionConsumerInfo.needRequestsQueue = true
	connectionConsumerInfo.needResponseQueue = false
	connectionConsumerInfo.isConsumer = true

	err := connectionConsumerInfo.initQueue()
	if err != nil {
		return err
	}

	connectionPublisherInfo = connectionInfoType{}
	connectionPublisherInfo.region = region
	connectionPublisherInfo.subRegion = subRegion
	connectionPublisherInfo.needRequestsQueue = false
	connectionPublisherInfo.needResponseQueue = true
	connectionPublisherInfo.isPublisher = true

	err = connectionPublisherInfo.initQueue()
	if err != nil {
		return err
	}

	return nil

}

func InitQueueScheduler() error {
	if connectionPublisherInfo.initialised {
		return errors.New("connection already initialised")
	}

	connectionPublisherInfo = connectionInfoType{}
	connectionPublisherInfo.needRequestsQueue = true
	connectionPublisherInfo.allRequestsQueuesNeeded = true
	connectionPublisherInfo.isPublisher = true

	return connectionPublisherInfo.initQueue()
}

func InitQueueResponseCollector() error {
	if connectionConsumerInfo.initialised {
		return errors.New("connection already initialised")
	}
	connectionConsumerInfo = connectionInfoType{}
	connectionConsumerInfo.needResponseQueue = true
	connectionConsumerInfo.isConsumer = true

	return connectionConsumerInfo.initQueue()

}

func (ci *connectionInfoType) initQueue() error {
	var err error
	ci.initialised = true

	ci.queueBrokerConnection, err = amqp.Dial(os.Getenv("QUEUEURL"))
	if err != nil {
		return err
	}

	ci.queueBrokerChannel, err = ci.queueBrokerConnection.Channel()
	if err != nil {
		return err
	}

	if ci.needRequestsQueue {
		err = ci.initQueuesRequests()
		if err != nil {
			return err
		}
	}

	if ci.needResponseQueue {
		err = ci.initQueuesResponses()
		if err != nil {
			return err
		}
	}

	if ci.isConsumer {
		err = ci.queueBrokerChannel.Qos(settings.GetSettInt(QUEUEPREFETCHCOUNT), 0, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ci *connectionInfoType) initQueuesRequests() error {
	var queueDeclared bool
	// REQUESTS QUEUES
	regions, err := settings.GetRegionsList()
	utilities.FailOnError(err)

	for _, r := range regions {
		// todo once system is stable implement here the check to verify that the region flag is enabled?
		for _, sr := range r.SubRegions {
			queueFullName := BuildRequestsQueueName(r.Id, sr.Id)
			// todo once system is stable implement here the check to verify that the subregion flag is enabled?

			// if declaring for a worker make sure we are declaring only the queue for the right region/subregion....
			if ci.allRequestsQueuesNeeded == false && (ci.region != r.Id || ci.subRegion != sr.Id) {
				continue
			}

			// TODO do we need to declare a queue even if we are only consuming it? it should already exists, created by "a publisher" before us... 🤔

			// create the queue for the sub region. queue name is [queuebasename].[region].[subregion]
			_, err := ci.queueBrokerChannel.QueueDeclare(queueFullName, true, false, false, false, nil)
			if err != nil {
				return err
			}
			fmt.Println(queueFullName)
			// create queue bindings with topic exchange
			err = ci.queueBrokerChannel.QueueBind(queueFullName, BuildRequestsQueueBindingKey(r.Id, sr.Id), "amq.topic", false, nil)
			if err != nil {
				return err
			}
			fmt.Println(BuildRequestsQueueBindingKey(r.Id, sr.Id))
			// at least one queue has been declared (this is particularly important for workers that are consuming only one specific queue....
			queueDeclared = true

		} // end for subregions loop
	} // end for regions loop

	if !queueDeclared {
		return errors.New("unable to declare any queue! is it a worker? is the region/subregion configured correctly")
	}

	return nil
}

func BuildRequestsQueueName(region, subRegion string) string {
	queueBaseName := settings.GetSettStr(QUEUENAMEREQUEST)
	if queueBaseName == "" {
		utilities.FailOnError(errors.New("requests queue base name is empty in settings"))
	}
	return fmt.Sprintf("%s.%s.%s", queueBaseName, region, subRegion)
}

func BuildRequestsQueueBindingKey(region, subRegion string) string {
	return fmt.Sprintf("%s.%s", region, subRegion)
}

func (ci *connectionInfoType) initQueuesResponses() error {
	// RESPONSES QUEUE
	_, err := ci.queueBrokerChannel.QueueDeclare(settings.GetSettStr(QUEUENAMERESPONSE), true, false, false, false, nil)
	if err != nil {
		return err
	}
	return nil
}

func (ci *connectionInfoType) GetQueueBrokerChannel() *amqp.Channel {
	return ci.queueBrokerChannel
}

func (ci *connectionInfoType) GetQueueBrokerConnection() *amqp.Connection {
	return ci.queueBrokerConnection
}

func (ci *connectionInfoType) Close() {
	ci.queueBrokerChannel.Close()
	ci.queueBrokerConnection.Close()
	ci.initialised = false
}

func (ci *connectionInfoType) CancelConsumer(consumerName string) {
	ci.queueBrokerChannel.Cancel(consumerName, false)
}

func CancelConsumer(consumerName string) {
	connectionConsumerInfo.CancelConsumer(consumerName)
}

func CloseConsumerConnection() {
	connectionConsumerInfo.Close()
}

func ClosePublisherConnection() {
	connectionPublisherInfo.Close()
}

func StartConsumingMessages(ctx context.Context, consumerName, queueName string, ch chan<- amqp.Delivery) error {
	msgs, err := connectionConsumerInfo.GetQueueBrokerChannel().Consume(queueName,
		consumerName,
		false,
		false,
		false,
		false,
		nil)

	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case msg := <-msgs:
				if len(msg.Body) == 0 {
					continue
				}
				// todo check if connection is still up
				ch <- msg
			case <-ctx.Done():
				CancelConsumer(consumerName)
				if len(msgs) == 0 {
					return
				}
			} // end select case
		} // end for
	}()

	return err
}

func GetConsumerChannel() chan amqp.Delivery {
	return ConsumerChannelExposed
}

func PublishToQueueDirectly(queueName string, body []byte) error {
	connectionPublisherInfo.channelMutex.Lock()
	defer connectionPublisherInfo.channelMutex.Unlock()
	err := connectionPublisherInfo.GetQueueBrokerChannel().Publish("",
		queueName,
		false,
		false,
		amqp.Publishing{Body: body, DeliveryMode: 2})
	return err
}

func PublishToTopicExchange(topicKey string, body []byte) error {
	connectionPublisherInfo.channelMutex.Lock()
	defer connectionPublisherInfo.channelMutex.Unlock()
	err := connectionPublisherInfo.GetQueueBrokerChannel().Publish("amq.topic",
		topicKey,
		false,
		false,
		amqp.Publishing{Body: body, DeliveryMode: 2})
	return err
}

//
// // for the moment we don't expose the whole connection monitor/reconnect status to the caller application, we keep in inside the queuehelper and see if that's enough
// // if for any reason we want/think it is needed to warn the app we will implement it
// func connectionMonitor() {
//
// 	// for the moment monitor only the connection, not the channel
// 	// if the channel monitor is re-enable please be aware that reconnecting triggers close notification
// 	// make sure you check if the connection is already up and that the error received is actually an error (!=nil)
// 	// errCh1 := GetQueueBrokerChannel().NotifyClose(make(chan *amqp.Error, 10))
// 	errCh2 := GetQueueBrokerConnection().NotifyClose(make(chan *amqp.Error, 10))
// 	for {
// 		select {
// 		// case err := <-errCh1:
// 		// 	fmt.Println("1--" + err.Error())
// 		// 	errInit := connectionInfo.initQueue()
// 		// 	if errInit != nil {
// 		//
// 		// 	}
// 		case err := <-errCh2:
// 			if err == nil {
// 				break // not a real error/connection issue?
// 			}
// 			connIsClosed := connectionInfo.queueBrokerConnection.IsClosed()
// 			if connIsClosed == false {
// 				break // we already have a connection open, for the moment break, in the future maybe reset connection?
// 			}
// 			fmt.Println("2--" + err.Error())
// 			errInit := connectionInfo.initQueue()
// 			if errInit != nil {
//
// 			}
// 		default:
// 			fmt.Print("")
//
// 		}
// 	}
// }
