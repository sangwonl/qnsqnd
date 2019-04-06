package core

import "github.com/satori/go.uuid"

type TopicMsgInbox struct {
	subscriptions *SubscriptionMap
	queue         chan *TopicMessage
}

func (i *TopicMsgInbox) addSubscription(s *Subscription) SubscriptionId {
	s.queue = make(chan *TopicMessage, 100)
	if s.Callback == "" {
		s.outStream = make(chan *TopicMessage, 100)
	}

	s.id = SubscriptionId(uuid.NewV4().String())
	i.subscriptions.put(s.id, s)

	go s.subscriptionQueueDispatcher()

	return s.id
}

func (i *TopicMsgInbox) delSubscription(s *Subscription) {
	close(s.queue)
	if s.outStream != nil {
		close(s.outStream)
	}
	i.subscriptions.del(s.id)
}
