package queuehelper

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"

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
	WorkerHostname            string                      `bson:"workerhostname"`
	WorkerHostnameFriendly    string                      `bson:"workerhostnamefriendly"`
	QueuedReturnUnix          int64                       `bson:"queuedreturnunix"`
	ReceivedByResponseHandler int64                       `bson:"receivedbyresponsehandler"`
	ErrorFatal                string                      `bson:"errorfatal"`
	RequestId                 string                      `bson:"requestid"`
}

func InitQueueWorker(region string, subRegion string) {
	region = strings.Trim(region, " ")
	subRegion = strings.Trim(subRegion, " ")

	if region == "" || subRegion == "" {
		utilities.FailOnError(errors.New("both region and subregion values need to be populated"))
	}
	initQueue(true, true, region, subRegion)
}

func InitQueueScheduler() {
	initQueue(true, false, "", "")
}

func InitQueueResponseCollector() {
	initQueue(false, true, "", "")

}

func initQueue(requests bool, responses bool, workerRegion string, workerSubRegion string) {
	var err error
	queueBrokerConnection, err = amqp.Dial(os.Getenv("QUEUEURL"))
	utilities.FailOnError(err)

	queueBrokerChannel, err = queueBrokerConnection.Channel()
	utilities.FailOnError(err)

	if requests {
		initQueuesRequests(workerRegion != "", workerRegion, workerSubRegion)
	}

	if responses {
		initQueuesResponses()
	}

	err = queueBrokerChannel.Qos(settings.GetSettInt(QUEUEPREFETCHCOUNT), 0, false)
	utilities.FailOnError(err)

}

func initQueuesRequests(worker bool, region, subRegion string) {
	var queueDeclared bool
	// REQUESTS QUEUES
	regions, err := settings.GetRegionsList()
	utilities.FailOnError(err)

	for _, r := range regions {
		// todo once system is stable implement check for region enabled flag here
		for _, sr := range r.SubRegions {
			queueFullName := BuildRequestsQueueName(r.Id, sr.Id)
			// todo once system is stable implement check for sub region enabled flag here

			// if declaring for a worker make sure we are declaring the queue for the righ region/subregion only....
			if worker && (region != r.Id || subRegion != sr.Id) {
				continue
			}

			// create the queue for the sub region. queue name is [queuebasename].[region].[subregion]
			_, err := queueBrokerChannel.QueueDeclare(queueFullName, true, false, false, false, nil)
			utilities.FailOnError(err)

			// create queue bindings with topic exchange
			err = queueBrokerChannel.QueueBind(queueFullName, BuildRequestsQueuebindingKey(r.Id, sr.Id), "amq.topic", false, nil)
			utilities.FailOnError(err)

			// at least one queue has been declared (this is particularly important for workers that are consuming only one specific queue....
			queueDeclared = true

		} // end for subregions loop
	} // end for regions loop

	if !queueDeclared {
		utilities.FailOnError(errors.New("unable to declare queue for worker, region/subregion configuration not found"))
	}
}

func BuildRequestsQueueName(region, subRegion string) string {
	queueBaseName := settings.GetSettStr(QUEUENAMEREQUEST)
	if queueBaseName == "" {
		utilities.FailOnError(errors.New("requests queue base name is empty in settings"))
	}
	return fmt.Sprintf("%s.%s.%s", queueBaseName, region, subRegion)
}

func BuildRequestsQueuebindingKey(region, subRegion string) string {
	return fmt.Sprintf("%s.%s", region, subRegion)
}

func initQueuesResponses() {
	// RESPONSES QUEUE
	_, err := queueBrokerChannel.QueueDeclare(settings.GetSettStr(QUEUENAMERESPONSE), true, false, false, false, nil)
	utilities.FailOnError(err)

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
