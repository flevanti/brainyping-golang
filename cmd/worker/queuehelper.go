package main

import (
	"context"

	"brainyping/pkg/queuehelper"
	"brainyping/pkg/settings"

	"github.com/streadway/amqp"
)

func PublishResponseForCheckProcessed(body []byte) error {
	err := queuehelper.PublishToQueueDirectly(settings.GetSettStr(queuehelper.QUEUENAMERESPONSE), body)
	return err
}

func ConsumeQueueForPendingChecks(ctx context.Context, ch chan<- amqp.Delivery) error {
	queueName := queuehelper.BuildRequestsQueueName(settings.GetSettStr(WORKERREGION), settings.GetSettStr(WORKERSUBREGION))
	return queuehelper.StartConsumingMessages(ctx, QUEUECONSUMERNAME, queueName, ch)

}
