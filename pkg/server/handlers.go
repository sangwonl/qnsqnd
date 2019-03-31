package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sangwonl/qnsqnd/pkg/core"
	"io"
)

type HandlerContext struct {
	pubSubManager *core.PubSubManager
}

func (h *HandlerContext) handlePublish(c *gin.Context) {
	var m core.TopicMessage
	err := c.BindJSON(&m)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid topic message"})
		return
	}

	err = h.pubSubManager.Publish(&m)
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to publish topic message"})
		return
	}

	c.JSON(202, gin.H{})
}

func (h *HandlerContext) handleSubscribe(c *gin.Context) {
	var s core.Subscription
	err := c.BindJSON(&s)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid subscription"})
		return
	}

	subId, err := h.pubSubManager.Subscribe(&s)
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to subscribe"})
		return
	}

	c.JSON(202, gin.H{"subscriptionId": subId})
}

func (h *HandlerContext) handleSubscribeSSE(c *gin.Context) {
	p := c.Params[0]
	if p.Key != "subId" || p.Value == "" {
		c.JSON(400, gin.H{"error": "invalid subscription"})
		return
	}

	subId := core.SubscriptionId(p.Value)
	stream, err := h.pubSubManager.Stream(subId)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid subscription"})
		return
	}

	c.Stream(func(w io.Writer) bool {
		if m, more := <-stream; more {
			c.SSEvent("message", m)
			return true
		}
		return false
	})
}

func (h *HandlerContext) handleSubscribeCancel(c *gin.Context) {
	p := c.Params[0]
	if p.Key != "subId" || p.Value == "" {
		c.JSON(400, gin.H{"error": "invalid subscription"})
		return
	}

	subId := core.SubscriptionId(p.Value)
	err := h.pubSubManager.Unsubscribe(subId)
	if err != nil {
		c.JSON(400, gin.H{"error": "no such subscription"})
		return
	}

	c.JSON(202, gin.H{})
}
