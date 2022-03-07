package main

import (
	"context"

	"brainyping/pkg/queuehelper"
	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"github.com/streadway/amqp"
)

func PublishResponseForCheckProcessed(body []byte) error {
	err := queuehelper.GetQueueBrokerChannel().Publish("",
		settings.GetSettStr(queuehelper.QUEUENAMERESPONSE),
		false,
		false,
		amqp.Publishing{Body: body, DeliveryMode: 2})
	return err
}

func ConsumeQueueForPendingChecks(ctx context.Context, ch chan<- amqp.Delivery) {
	msgs, err := queuehelper.GetQueueBrokerChannel().Consume(settings.GetSettStr(queuehelper.QUEUENAMEREQUEST),
		QUEUECONSUMERNAME,
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
			// the context was cancelled, stop receiving messages from the queue
			// cancel the consumer and return
			// records prefetched - if any - will be returned to the queue by the rabbit client when the connection is closed
			queuehelper.GetQueueBrokerChannel().Cancel(QUEUECONSUMERNAME, false)
			return

		} // end select case
	} // end for

}
