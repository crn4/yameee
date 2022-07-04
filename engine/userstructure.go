package engine

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type UserData struct {
	userID      int32
	name        string
	client      chan string
	connection  *websocket.Conn
	chatID      string
	broadcaster *broadcast
}

type Peer struct {
	peerID     int32
	connection *websocket.Conn
	clientChan chan string
	name       string
}

type Peers struct {
	sync.RWMutex
	peers map[int32]*Peer
}

type Message struct {
	chatID  string
	message string
}

func NewClient(conn *websocket.Conn, br *broadcast, name, chatID string) *UserData {
	return &UserData{
		userID:      rand.Int31(),
		client:      make(chan string),
		connection:  conn,
		name:        name,
		chatID:      chatID,
		broadcaster: br}
}

func (ud *UserData) SetName(s string) error {
	if 0 < len([]byte(s)) && len([]byte(s)) < 20 {
		ud.name = s
		return nil
	}
	return fmt.Errorf("%s was not proper to be set as acc name", s)
}

// this func should be redesigned for secure chat id
func (ud *UserData) SetChatID(s string) error {
	if 0 < len([]byte(s)) && len([]byte(s)) < 64 {
		ud.chatID = s
		return nil
	}
	return fmt.Errorf("%s was not proper to be set as Chat ID", s)
}

func (ud *UserData) GetCurrentChatID() string {
	return ud.chatID
}

func (cli *UserData) clientReader() {
	conn := cli.connection
	br := cli.broadcaster
	cli.client <- "Welcome, " + cli.name
	br.entering <- *cli
	br.messages <- *cli.composeMessage(cli.name + " joined the room")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		br.messages <- *cli.composeMessage(cli.name + ">> " + string(message))
	}

	br.leaving <- *cli
	br.messages <- *cli.composeMessage(cli.name + " has left")
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

func (cli *UserData) composeMessage(message string) *Message {
	return &Message{cli.chatID, message}
}
