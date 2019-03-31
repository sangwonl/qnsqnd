package main

import (
	"fmt"
	"github.com/sangwonl/qnsqnd/pkg/server"
)

func main() {
	s := server.InitServer()
	if err := s.Run(); err != nil {
		fmt.Printf("failed to run server... %s", err)
	}
}
