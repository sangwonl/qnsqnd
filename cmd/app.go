package main

import (
	"fmt"
	"github.com/sangwonl/qnsqnd/pkg/server"
)

func main() {
	s := server.InitServer()
	if err := s.Run(8080); err != nil {
		fmt.Printf("failed to run server... %s", err)
	}
}
