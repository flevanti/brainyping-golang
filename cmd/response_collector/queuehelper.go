package main

import (
	"context"

	"brainyping/pkg/queuehelper"
	"brainyping/pkg/settings"
	"brainyping/pkg/utilities"

	"github.com/streadway/amqp"
)

func ConsumeQueueForResponsesToChecks(ctx context.Context, ch chan<- amqp.Delivery) {
	msgs, err := queuehelper.GetQueueBrokerChannel().Consume(settings.GetSettStr(queuehelper.QUEUENAMERESPONSE),
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
			// the context was cancelled, stop working
			// cancel the consumer...
			_ = queuehelper.GetQueueBrokerChannel().Cancel(QUEUECONSUMERNAME, false)
			return
		} // end select case
	} // end for

}
