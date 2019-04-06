package core

import (
	"fmt"
	"gopkg.in/resty.v1"
)

func (i *TopicMsgInbox) topicQueueDispatcher() {
	for {
		m, more := <-i.queue
		if more {
			for _, s := range i.subscriptions.items() {
				s.queue <- m
			}
		} else {
			break
		}
	}
}

func (s *Subscription) processDelivery(m *TopicMessage) {
	if s.Filter != "" {
		if !evaluateExpression(m.Payload, s.Filter) {
			return
		}
	}

	if s.Callback != "" {
		_, err := resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(m).
			Post(s.Callback)
		if err != nil {
			fmt.Print(err)
		}
	} else if s.outStream != nil {
		s.outStream <- m
	}
}

func (s *Subscription) subscriptionQueueDispatcher() {
	for {
		m, more := <-s.queue
		if more {
			s.processDelivery(m)
		} else {
			break
		}
	}
}
