package main

import (
	"fmt"
	"github.com/sangwonl/qnsqnd/pkg/core"
)

func main() {
	s := core.InitServer()
	if err := s.Run(); err != nil {
		fmt.Printf("failed to run server... %s", err)
	}
}
