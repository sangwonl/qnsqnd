package core

import "github.com/satori/go.uuid"

type TopicMsgInbox struct {
	subscriptions *SubscriptionMap
	queue         chan *TopicMessage
}

func (i *TopicMsgInbox) addSubscription(s *Subscription) SubscriptionId {
	s.id = SubscriptionId(uuid.NewV4().String())
	s.queue = make(chan *TopicMessage, 100)
	if s.Callback == "" {
		s.outStream = make(chan *TopicMessage, 100)
	}

	i.subscriptions.put(s.id, s)

	go s.subscriptionQueueDispatcher()

	return s.id
}

func (i *TopicMsgInbox) delSubscription(s *Subscription) {
	i.subscriptions.del(s.id)
	if s.outStream != nil {
		close(s.outStream)
	}
	close(s.queue)
}
