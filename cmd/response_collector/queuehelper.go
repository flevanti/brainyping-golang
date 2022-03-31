package main

import (
	"context"

	"brainyping/pkg/queuehelper"
	"brainyping/pkg/settings"

	"github.com/streadway/amqp"
)

func ConsumeQueueForResponsesToChecks(ctx context.Context, ch chan<- amqp.Delivery) error {
	// msgs, err := queuehelper.GetQueueBrokerChannel().Consume(settings.GetSettStr(queuehelper.QUEUENAMERESPONSE),
	// 	QUEUECONSUMERNAME,
	// 	false,
	// 	false,
	// 	false,
	// 	false,
	// 	nil)
	//
	// utilities.FailOnError(err)
	//
	// for {
	// 	select {
	// 	case msg := <-msgs:
	// 		ch <- msg
	// 	case <-ctx.Done():
	// 		// the context was cancelled, stop working
	// 		// cancel the consumer...
	// 		_ = queuehelper.GetQueueBrokerChannel().Cancel(QUEUECONSUMERNAME, false)
	// 		return
	// 	} // end select case
	// } // end for

	// queueName := queuehelper.BuildRequestsQueueName(settings.GetSettStr(WORKERREGION), settings.GetSettStr(WORKERSUBREGION))
	return queuehelper.StartConsumingMessages(ctx, QUEUECONSUMERNAME, settings.GetSettStr(queuehelper.QUEUENAMERESPONSE), ch)

}
