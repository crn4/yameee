package main

import (
	"log"
	"mytests/yetanothermessenger/server/engine"
	"net/http"
)

func main() {
	engine.BroadcastManager()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		engine.UserRegistrator(w, r)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Listen and serve: ", err)
	}
}
