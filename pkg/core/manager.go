package core

import (
	"errors"
	"sync"
)

var singletonInstance *PubSubManager
var once sync.Once

type PubSubManager struct {
	perTopicMsgInbox map[Topic]*TopicMsgInbox
	allSubscriptions map[SubscriptionId]*Subscription
}

func GetPubSubManager() *PubSubManager {
	once.Do(func() {
		singletonInstance = &PubSubManager{
			make(map[Topic]*TopicMsgInbox),
			make(map[SubscriptionId]*Subscription),
		}
	})
	return singletonInstance
}

func (mgr *PubSubManager) Subscribe(s *Subscription) (SubscriptionId, error) {
	if !s.validate() {
		return "", errors.New("invalid subscription data")
	}

	inbox := mgr.getTopicMsgInbox(s.Topic)
	if inbox == nil {
		inbox = mgr.newTopicMsgInbox(s.Topic)
	}
	subId := inbox.addSubscription(s)
	mgr.allSubscriptions[subId] = s
	return subId, nil
}

func (mgr *PubSubManager) Publish(m *TopicMessage) error {
	if !m.validate() || mgr.getTopicMsgInbox(m.Topic) == nil {
		return errors.New("invalid Topic message data")
	}

	if inbox := mgr.getTopicMsgInbox(m.Topic); inbox != nil {
		inbox.queue <- m
	}
	return nil
}

func (mgr *PubSubManager) Unsubscribe(subId SubscriptionId) error {
	s, ok := mgr.allSubscriptions[subId]
	if !ok {
		return errors.New("no such subscription")
	}

	i, ok := mgr.perTopicMsgInbox[s.Topic]
	if !ok {
		return errors.New("msg inbox not found")
	}

	i.delSubscription(s)
	delete(mgr.allSubscriptions, subId)
	return nil
}

func (mgr *PubSubManager) Stream(subId SubscriptionId) (<-chan *TopicMessage, error) {
	s, ok := mgr.allSubscriptions[subId]
	if !ok || s.outStream == nil {
		return nil, errors.New("no such subscription")
	}
	return s.outStream, nil
}

func (mgr *PubSubManager) newTopicMsgInbox(t Topic) *TopicMsgInbox {
	inbox := TopicMsgInbox{
		make(map[SubscriptionId]*Subscription, 0),
		make(chan *TopicMessage, 1000),
	}
	mgr.perTopicMsgInbox[t] = &inbox
	go inbox.topicQueueDispatcher()
	return &inbox
}

func (mgr *PubSubManager) getTopicMsgInbox(t Topic) *TopicMsgInbox {
	if inbox, ok := mgr.perTopicMsgInbox[t]; ok {
		return inbox
	}
	return nil
}
