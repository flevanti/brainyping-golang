package main

import (
	"context"
	"time"

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
	var lastMessageReceived time.Time
	var coolingDown bool
	msgs, err := queuehelper.GetQueueBrokerChannel().Consume(settings.GetSettStr(queuehelper.QUEUENAMEREQUEST),
		QUEUECONSUMERNAME,
		false,
		false,
		false,
		false,
		nil)

	utilities.FailOnError(err)

	for {
		a := len(msgs)
		_ = a
		select {
		case msg := <-msgs:
			ch <- msg
			lastMessageReceived = time.Now()

		case <-ctx.Done():
			// the context was cancelled, stop working
			// cancel the consumer...
			if !coolingDown {
				_ = queuehelper.GetQueueBrokerChannel().Cancel(QUEUECONSUMERNAME, false)
				coolingDown = true
				continue
			}
			// TODO THIS MECHANISM IS NOT YET COMPLETED, THE OVERALL SHUTTING DOWN PROCESS IS UNAWARE OF THIS PROCESS SO IT WON'T WAIT FOR IT.
			if coolingDown && time.Since(lastMessageReceived).Seconds() > 3 {
				return
			}

		} // end select case
	} // end for

}
