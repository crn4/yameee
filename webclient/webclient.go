package main

import (
	"log"
	"net/http"
)

func main() {
	fileServ := http.FileServer(http.Dir("."))
	if err := http.ListenAndServe(":80", fileServ); err != nil {
		log.Fatal(err)
	}
}
