package engine

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, //  TODO need to develop a check that connection is from applicable client
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
)

func UserRegistrator(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	name, ok := r.URL.Query()["name"]
	if !ok || len(name[0]) == 0 {
		log.Println("Name was not send")
		return
	}
	chatID, ok := r.URL.Query()["chatID"]
	if !ok || len(name[0]) == 0 {
		log.Println("chatID was not send")
		return
	}
	broadcastRequest <- chatID[0]
	br, ok := <-broadcastAnswer
	if !ok {
		return
	}
	cli := NewClient(conn, br, name[0], chatID[0])
	go cli.clientReader()
	go cli.clientWriter()
}
