package core

import (
	"errors"
	"sync"
)

var singletonInstance *PubSubManager
var once sync.Once

type PubSubManager struct {
	perTopicMsgInbox *TopicInboxMap
	allSubscriptions *SubscriptionMap
}

func GetPubSubManager() *PubSubManager {
	once.Do(func() {
		singletonInstance = &PubSubManager{
			createTopicInboxMap(),
			createSubscriptionMap(),
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
	mgr.allSubscriptions.put(subId, s)

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
	s := mgr.allSubscriptions.get(subId)
	if s == nil {
		return errors.New("no such subscription")
	}

	inbox := mgr.perTopicMsgInbox.get(s.Topic)
	if inbox == nil {
		return errors.New("msg inbox not found")
	}

	inbox.delSubscription(s)
	mgr.allSubscriptions.del(subId)

	return nil
}

func (mgr *PubSubManager) Stream(subId SubscriptionId) (<-chan *TopicMessage, error) {
	s := mgr.allSubscriptions.get(subId)
	if s == nil || s.outStream == nil {
		return nil, errors.New("no such subscription")
	}
	return s.outStream, nil
}

func (mgr *PubSubManager) newTopicMsgInbox(t Topic) *TopicMsgInbox {
	inbox := TopicMsgInbox{
		createSubscriptionMap(),
		make(chan *TopicMessage, 1000),
	}
	mgr.perTopicMsgInbox.put(t, &inbox)

	go inbox.topicQueueDispatcher()

	return &inbox
}

func (mgr *PubSubManager) getTopicMsgInbox(t Topic) *TopicMsgInbox {
	return mgr.perTopicMsgInbox.get(t)
}
