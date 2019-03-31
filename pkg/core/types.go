package core

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
