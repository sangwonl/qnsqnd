package server

import (
	"fmt"
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

func (s *Server) Run(port uint16) error {
	return s.router.Run(fmt.Sprintf(":%d", port))
}
