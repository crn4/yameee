package main

import (
	"github.com/crn4/yameee/engine"
)

func main() {
	s := engine.NewServer()
	s.Start(":8080")
}
