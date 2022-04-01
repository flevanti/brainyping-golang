package main

import (
	"brainyping/pkg/queuehelper"
)

func PublishRequestForNewCheck(body []byte, region string, subRegion string) error {
	// please note that the messages are published in a durable way, this is probably more useful in dev phase than prod
	// we should probably create a flag to accomodate this... 🤠.....
	return queuehelper.PublishToTopicExchange(queuehelper.BuildRequestsQueueBindingKey(region, subRegion), body)
}
