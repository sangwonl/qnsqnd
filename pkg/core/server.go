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
	Topic     Topic  `json:"topic"`
	Filter    string `json:"filter"`
	Callback  string `json:"callback"`
	Queue     chan TopicMessage
	SSEStream chan TopicMessage
}

func (s *Subscription) subscriptionWorker() {
	for m := range s.Queue {
		evaluated := true
		if s.Filter != "" {
			bytes, err := json.Marshal(m.Payload)
			if err != nil {
				continue
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

		if evaluated {
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
	}
}

type TopicMessageInbox struct {
	Queue         chan TopicMessage
	Subscriptions []*Subscription
}

func (i *TopicMessageInbox) topicMessageInboxWorker() {
	for m := range i.Queue {
		for _, s := range i.Subscriptions {
			s.Queue <- m
		}
	}
}

func (i *TopicMessageInbox) addSubscription(sub *Subscription) SubscriptionId {
	sub.Queue = make(chan TopicMessage, 100)
	if sub.Callback == "" {
		sub.SSEStream = make(chan TopicMessage, 100)
	}

	subId := SubscriptionId(uuid.NewV4().String())
	i.Subscriptions = append(i.Subscriptions, sub)
	go sub.subscriptionWorker()

	return subId
}

type HandlerContext struct {
	TopicMsgInbox map[Topic]*TopicMessageInbox
	Subscriptions map[SubscriptionId]*Subscription
}

func (h *HandlerContext) createTopicMessageInbox(t Topic) *TopicMessageInbox {
	inbox := TopicMessageInbox{
		make(chan TopicMessage, 1000),
		make([]*Subscription, 0),
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

func (h *HandlerContext) registerSubscription(sub *Subscription) SubscriptionId {
	inbox := h.getTopicMessageInbox(sub.Topic)
	if inbox == nil {
		inbox = h.createTopicMessageInbox(sub.Topic)
	}
	subId := inbox.addSubscription(sub)
	h.Subscriptions[subId] = sub
	return subId
}

func (h *HandlerContext) validateTopicMsg(message *TopicMessage) bool {
	return message.Topic != "" &&
		message.Payload != nil &&
		h.isSubscribedTopic(message.Topic)
}

func (h *HandlerContext) validateSubscription(sub *Subscription) bool {
	return true
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
		inbox.Queue <- message
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

	c.JSON(201, gin.H{"subscriptionId": subscriptionId})
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
	subscription, ok := h.Subscriptions[subscriptionId]
	if !ok {
		c.JSON(400, gin.H{
			"error": "no such subscription",
		})
		return
	}

	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-subscription.SSEStream; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})
}

func setupRouter(handlerCtx *HandlerContext) *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())
	r.POST("/publish", handlerCtx.handlePublish)
	r.POST("/subscribe", handlerCtx.handleSubscribe)
	r.GET("/subscribe/:subId/sse", handlerCtx.handleSubscribeSSE)
	return r
}
