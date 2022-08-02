package main

import (
	"log"

	"github.com/crn4/yameee/engine"
)

func main() {
	s := engine.NewServer()
	err := s.Start(":8080")
	if err != nil {
		log.Fatal("Error starting the server: ", err)
	}
}
