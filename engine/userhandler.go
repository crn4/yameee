package engine

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	envelopeTypeService = "SRV"
	envelopeTypeMessage = "MSG"
	envelopeTypeKey     = "KEY"
)

type UserData struct {
	userID      int32
	name        string
	client      chan Message
	connection  *websocket.Conn
	chatID      string
	broadcaster *broadcast
	publicKey   string
}

type Peers struct {
	sync.RWMutex
	peers map[int32]*UserData
}

type Message struct {
	MsgType string
	Name    string
	ChatID  string
	Message string
}

func NewClient(conn *websocket.Conn, br *broadcast, name, chatID string) *UserData {
	return &UserData{
		userID:      rand.Int31(),
		client:      make(chan Message),
		connection:  conn,
		name:        name,
		chatID:      chatID,
		broadcaster: br}
}

func (cli *UserData) SetPublicKey(key string) {
	cli.publicKey = key
}

func (cli *UserData) clientReader() {
	conn := cli.connection
	br := cli.broadcaster
	cli.client <- cli.composeMessage(envelopeTypeService, "Welcome, "+cli.name)
	br.entering <- cli
	br.messages <- cli.composeMessage(envelopeTypeService, cli.name+" joined the room")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("%s - %s\n", err, "message JSON was not unmarshalled")
		}
		switch msg.MsgType {
		case envelopeTypeMessage:
			br.messages <- msg
		case envelopeTypeKey:
			cli.SetPublicKey(msg.Message)
			br.handshake <- cli
		}

	}

	br.leaving <- cli
	br.messages <- cli.composeMessage(envelopeTypeService, cli.name+" has left")
	conn.Close()
}

func (cli *UserData) clientWriter() {
	conn := cli.connection
	ch := cli.client
	for msg := range ch {
		conn.SetWriteDeadline(time.Now().Add(writeWait))

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		if jsonMsg, err := json.Marshal(msg); err != nil {
			log.Printf("%s - %s\n", err, "Message struct was not marshalled")
		} else {
			w.Write(jsonMsg)
		}

		for i := 0; i < len(ch); i++ {
			w.Write([]byte{'\n'})
			jsonMsg, _ := json.Marshal(<-ch)
			w.Write(jsonMsg)
		}

		if err := w.Close(); err != nil {
			return
		}
	}
}

func (cli *UserData) composeMessage(msgType string, message string) Message {
	return Message{msgType, cli.name, cli.chatID, message}
}
