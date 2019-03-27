package core

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/json"
	"github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"gopkg.in/resty.v1"
	"io"
	"strings"
)

type Topic string
type SubscriptionId string

type TopicMessage struct {
	Topic   Topic       `json:"topic"`
	Payload interface{} `json:"payload"`
}

type Subscription struct {
	Id        SubscriptionId
	Topic     Topic  `json:"topic"`
	Filter    string `json:"filter"`
	Callback  string `json:"callback"`
	Queue     chan *TopicMessage
	SSEStream chan *TopicMessage
}

func (s *Subscription) processDelivery(m *TopicMessage) {
	evaluated := true

	if s.Filter != "" {
		bytes, err := json.Marshal(m.Payload)
		if err != nil {
			return
		}

		functions := map[string]govaluate.ExpressionFunction{
			"F": func(args ...interface{}) (interface{}, error) {
				fieldExp := args[0].(string)
				if strings.HasPrefix(fieldExp, "@") {
					fieldExp = "\\" + fieldExp
				}
				result := gjson.GetBytes(bytes, fieldExp)
				return result.Value(), nil
			},
		}

		expression, _ := govaluate.NewEvaluableExpressionWithFunctions(s.Filter, functions)
		result, _ := expression.Evaluate(nil)
		evaluated = result.(bool)
	}

	if !evaluated {
		return
	}

	if s.Callback != "" {
		_, err := resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(m).
			Post(s.Callback)
		if err != nil {
			fmt.Print(err)
		}
	} else if s.SSEStream != nil {
		s.SSEStream <- m
	}
}

func (s *Subscription) subscriptionWorker() {
	for {
		m, more := <-s.Queue
		if more {
			s.processDelivery(m)
		} else {
			break
		}
	}
}

type TopicMessageInbox struct {
	Subscriptions map[SubscriptionId]*Subscription
	Queue         chan *TopicMessage
}

func (i *TopicMessageInbox) topicMessageInboxWorker() {
	for {
		m, more := <-i.Queue
		if more {
			for _, s := range i.Subscriptions {
				s.Queue <- m
			}
		} else {
			break
		}
	}
}

func (i *TopicMessageInbox) addSubscription(s *Subscription) SubscriptionId {
	s.Queue = make(chan *TopicMessage, 100)
	if s.Callback == "" {
		s.SSEStream = make(chan *TopicMessage, 100)
	}

	s.Id = SubscriptionId(uuid.NewV4().String())
	i.Subscriptions[s.Id] = s
	go s.subscriptionWorker()

	return s.Id
}

func (i *TopicMessageInbox) delSubscription(s *Subscription) {
	close(s.Queue)
	if s.SSEStream != nil {
		close(s.SSEStream)
	}
	delete(i.Subscriptions, s.Id)
}

type HandlerContext struct {
	TopicMsgInbox map[Topic]*TopicMessageInbox
	Subscriptions map[SubscriptionId]*Subscription
}

func (h *HandlerContext) createTopicMessageInbox(t Topic) *TopicMessageInbox {
	inbox := TopicMessageInbox{
		make(map[SubscriptionId]*Subscription, 0),
		make(chan *TopicMessage, 1000),
	}
	h.TopicMsgInbox[t] = &inbox
	go inbox.topicMessageInboxWorker()
	return &inbox
}

func (h *HandlerContext) isSubscribedTopic(t Topic) bool {
	_, ok := h.TopicMsgInbox[t]
	return ok
}

type Server struct {
	router     *gin.Engine
	handlerCtx *HandlerContext
}

func InitServer() *Server {
	handlerCtx := HandlerContext{
		make(map[Topic]*TopicMessageInbox),
		make(map[SubscriptionId]*Subscription),
	}

	return &Server{
		setupRouter(&handlerCtx),
		&handlerCtx,
	}
}

func (s *Server) Run() (err error) {
	err = s.router.Run(":8080")
	return
}

func (h *HandlerContext) getTopicMessageInbox(t Topic) *TopicMessageInbox {
	if inbox, ok := h.TopicMsgInbox[t]; ok {
		return inbox
	}
	return nil
}

func (h *HandlerContext) registerSubscription(s *Subscription) SubscriptionId {
	inbox := h.getTopicMessageInbox(s.Topic)
	if inbox == nil {
		inbox = h.createTopicMessageInbox(s.Topic)
	}
	subId := inbox.addSubscription(s)
	h.Subscriptions[subId] = s
	return subId
}

func (h *HandlerContext) validateTopicMsg(message *TopicMessage) bool {
	return message.Topic != "" &&
		message.Payload != nil &&
		h.isSubscribedTopic(message.Topic)
}

func (h *HandlerContext) validateSubscription(s *Subscription) bool {
	return s.Topic != ""
}

func (h *HandlerContext) handlePublish(c *gin.Context) {
	var message TopicMessage
	err := c.BindJSON(&message)
	if err != nil || !h.validateTopicMsg(&message) {
		c.JSON(400, gin.H{
			"error": "invalid topic message",
		})
		return
	}

	if inbox := h.getTopicMessageInbox(message.Topic); inbox != nil {
		inbox.Queue <- &message
	}

	c.JSON(202, gin.H{})
}

func (h *HandlerContext) handleSubscribe(c *gin.Context) {
	var subscription Subscription
	err := c.BindJSON(&subscription)
	if err != nil || !h.validateSubscription(&subscription) {
		c.JSON(400, gin.H{
			"error": "invalid subscription",
		})
		return
	}

	subscriptionId := h.registerSubscription(&subscription)

	c.JSON(202, gin.H{"subscriptionId": subscriptionId})
}

func (h *HandlerContext) handleSubscribeSSE(c *gin.Context) {
	p := c.Params[0]
	if p.Key != "subId" || p.Value == "" {
		c.JSON(400, gin.H{
			"error": "invalid subscription",
		})
		return
	}

	subscriptionId := SubscriptionId(p.Value)
	s, ok := h.Subscriptions[subscriptionId]
	if !ok {
		c.JSON(400, gin.H{
			"error": "no such subscription",
		})
		return
	}

	c.Stream(func(w io.Writer) bool {
		if msg, more := <-s.SSEStream; more {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})
}

func (h *HandlerContext) closeSubscription(s *Subscription, subscriptionId SubscriptionId) {
	i, ok := h.TopicMsgInbox[s.Topic]
	if ok {
		i.delSubscription(s)
	}
	delete(h.Subscriptions, subscriptionId)
}

func (h *HandlerContext) handleSubscribeCancel(c *gin.Context) {
	p := c.Params[0]
	if p.Key != "subId" || p.Value == "" {
		c.JSON(400, gin.H{
			"error": "invalid subscription",
		})
		return
	}

	subscriptionId := SubscriptionId(p.Value)
	s, ok := h.Subscriptions[subscriptionId]
	if !ok {
		c.JSON(400, gin.H{
			"error": "no such subscription",
		})
		return
	}

	h.closeSubscription(s, subscriptionId)

	c.JSON(202, gin.H{})
}

func setupRouter(handlerCtx *HandlerContext) *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())
	r.POST("/publish", handlerCtx.handlePublish)
	r.POST("/subscribe", handlerCtx.handleSubscribe)
	r.GET("/subscribe/:subId/sse", handlerCtx.handleSubscribeSSE)
	r.POST("/subscribe/:subId/cancel", handlerCtx.handleSubscribeCancel)
	return r
}
