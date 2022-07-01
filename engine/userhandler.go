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
	go handleConn(conn, br, name[0], chatID[0])
}

func handleConn(conn *websocket.Conn, br *broadcast, name, chatID string) {
	cli := NewClient(conn, name, chatID)
	go clientWriter(conn, cli.client)

	cli.client <- "Welcome, " + cli.name
	br.entering <- *cli
	br.messages <- *composeMessage(cli.chatID, cli.name+" has arrived")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		br.messages <- *composeMessage(cli.chatID, cli.name+">> "+string(message))
	}

	br.leaving <- *cli
	br.messages <- *composeMessage(cli.chatID, cli.name+" has left")
	conn.Close()
}

func clientWriter(conn *websocket.Conn, ch <-chan string) {
	for msg := range ch {
		conn.SetWriteDeadline(time.Now().Add(writeWait))

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write([]byte(msg))

		for i := 0; i < len(ch); i++ {
			w.Write([]byte{'\n'})
			w.Write([]byte(<-ch))
		}

		if err := w.Close(); err != nil {
			return
		}
	}
}

func composeMessage(chatid string, message string) *Message {
	return &Message{chatid, message}
}
