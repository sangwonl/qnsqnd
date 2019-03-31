package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sangwonl/qnsqnd/pkg/core"
)

type Server struct {
	router     *gin.Engine
	handlerCtx *HandlerContext
}

func InitServer() *Server {
	handlerCtx := HandlerContext{
		core.GetPubSubManager(),
	}

	return &Server{
		setupRouter(&handlerCtx),
		&handlerCtx,
	}
}

func (s *Server) Run() error {
	return s.router.Run(":8080")
}
