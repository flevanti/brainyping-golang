package main

import (
	"brainyping/pkg/queuehelper"
	"brainyping/pkg/settings"

	"github.com/streadway/amqp"
)

func PublishRequestForNewCheck(body []byte) error {
	// please note that the messages are published in a durable way, this is probably more useful in dev phase than prod
	// we should probably create a flag to accomodate this... 🤠.....
	err := queuehelper.GetQueueBrokerChannel().Publish("",
		settings.GetSettStr(queuehelper.QUEUENAMEREQUEST),
		false,
		false,
		amqp.Publishing{Body: body, DeliveryMode: 2})
	return err
}
