package core

import "github.com/satori/go.uuid"

type TopicMsgInbox struct {
	subscriptions map[SubscriptionId]*Subscription
	queue         chan *TopicMessage
}

func (i *TopicMsgInbox) addSubscription(s *Subscription) SubscriptionId {
	s.queue = make(chan *TopicMessage, 100)
	if s.Callback == "" {
		s.outStream = make(chan *TopicMessage, 100)
	}

	s.id = SubscriptionId(uuid.NewV4().String())
	i.subscriptions[s.id] = s
	go s.subscriptionQueueDispatcher()

	return s.id
}

func (i *TopicMsgInbox) delSubscription(s *Subscription) {
	close(s.queue)
	if s.outStream != nil {
		close(s.outStream)
	}
	delete(i.subscriptions, s.id)
}
