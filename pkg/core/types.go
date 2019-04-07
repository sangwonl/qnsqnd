package core

import "sync"

type Topic string
type SubscriptionId string

type TopicMessage struct {
	Topic   Topic       `json:"topic"`
	Payload interface{} `json:"payload"`
}

type Subscription struct {
	id        SubscriptionId
	Topic     Topic  `json:"topic"`
	Filter    string `json:"filter"`
	Callback  string `json:"callback"`
	queue     chan *TopicMessage
	outStream chan *TopicMessage
}

func (m TopicMessage) validate() bool {
	return m.Topic != "" && m.Payload != nil
}

func (s Subscription) validate() bool {
	return s.Topic != ""
}

func (s *Subscription) enqueue(m *TopicMessage) {
	if !isTopicMsgChannelClosed(s.queue) {
		s.queue <- m
	}
}

type TopicInboxMap struct {
	mutex sync.RWMutex
	table map[Topic]*TopicMsgInbox
}

func createTopicInboxMap() *TopicInboxMap {
	return &TopicInboxMap{
		sync.RWMutex{},
		make(map[Topic]*TopicMsgInbox),
	}
}

func (m *TopicInboxMap) put(topic Topic, inbox *TopicMsgInbox) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.table[topic] = inbox
}

func (m *TopicInboxMap) get(topic Topic) *TopicMsgInbox {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.table[topic]
}

type SubscriptionMap struct {
	mutex sync.RWMutex
	table map[SubscriptionId]*Subscription
}

func createSubscriptionMap() *SubscriptionMap {
	return &SubscriptionMap{
		sync.RWMutex{},
		make(map[SubscriptionId]*Subscription),
	}
}

func (s *SubscriptionMap) put(id SubscriptionId, sub *Subscription) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.table[id] = sub
}

func (s *SubscriptionMap) get(id SubscriptionId) *Subscription {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.table[id]
}

func (s *SubscriptionMap) del(id SubscriptionId) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.table, id)
}

func (s *SubscriptionMap) items() []*Subscription {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	items := make([]*Subscription, len(s.table))
	i := 0
	for _, sub := range s.table {
		items[i] = sub
		i++
	}
	return items
}

