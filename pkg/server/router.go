package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRouter(handlerCtx *HandlerContext) *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())
	r.POST("/publish", handlerCtx.handlePublish)
	r.POST("/subscribe", handlerCtx.handleSubscribe)
	r.GET("/subscribe/:subId/sse", handlerCtx.handleSubscribeSSE)
	r.POST("/subscribe/:subId/cancel", handlerCtx.handleSubscribeCancel)
	return r
}
