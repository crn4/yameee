package engine

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

type UserData struct {
	userID     int32
	name       string
	client     chan string
	connection *websocket.Conn
	chatID     string
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

func NewClient(conn *websocket.Conn, name, chatID string) *UserData {
	return &UserData{
		userID:     rand.Int31(),
		client:     make(chan string),
		connection: conn,
		name:       name,
		chatID:     chatID}
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
