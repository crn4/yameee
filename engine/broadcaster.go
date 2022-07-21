package engine

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type server struct {
	broadcasts map[string]*broadcast
	rwMutex    sync.RWMutex
}

func NewServer() *server {
	br := make(map[string]*broadcast)
	return &server{broadcasts: br}
}

func (s *server) Start(port string) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.userRegistrator(w, r)
	})
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Listen and serve: ", err)
	}
}

func (s *server) broadcastController(chatID string) *broadcast {
	s.rwMutex.Lock()
	broadcast, found := s.broadcasts[chatID]
	if !found {
		broadcast = newBroadcast()
		s.broadcasts[chatID] = broadcast
		go broadcast.startBroadcast()
	}
	s.rwMutex.Unlock()
	return broadcast
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, //  TODO need to develop a check that connection is from applicable client
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
)

func (s *server) userRegistrator(w http.ResponseWriter, r *http.Request) {
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
	br := s.broadcastController(chatID[0])
	cli := NewClient(conn, br, name[0], chatID[0])
	go cli.clientReader()
	go cli.clientWriter()
}

type broadcast struct {
	entering  chan *UserData
	leaving   chan *UserData
	handshake chan *UserData
	messages  chan Message
}

func newBroadcast() *broadcast {
	entering := make(chan *UserData)
	leaving := make(chan *UserData)
	handshake := make(chan *UserData)
	messages := make(chan Message)
	return &broadcast{entering, leaving, handshake, messages}
}

func (br *broadcast) startBroadcast() {
	activeConnections := make(map[string]*Peers)
	for {
		select {
		case msg := <-br.messages:
			activeConnections[msg.ChatID].RWMutex.RLock()
			for _, peer := range activeConnections[msg.ChatID].peers {
				peer.client <- msg
			}
			activeConnections[msg.ChatID].RWMutex.RUnlock()
		case cli := <-br.entering:
			if value, found := activeConnections[cli.chatID]; !found {
				activeConnections[cli.chatID] = &Peers{peers: map[int32]*UserData{cli.userID: cli}}
			} else {
				value.RWMutex.Lock()
				value.peers[cli.userID] = cli
				value.RWMutex.Unlock()
			}
			cli.client <- *cli.composeMessage("SRV", getNamesByConnection(activeConnections[cli.chatID]))
		case cli := <-br.handshake:
			if value := activeConnections[cli.chatID]; len(value.peers) == 2 {
				exchangeKeysBetweenPeers(value.peers)
			}
		case cli := <-br.leaving:
			if value, found := activeConnections[cli.chatID]; found {
				value.RWMutex.Lock()
				delete(value.peers, cli.userID)
				close(cli.client)
				value.RWMutex.Unlock()
			}
		}
	}
}

func getNamesByConnection(ac *Peers) string {
	result := ""
	for _, cli := range ac.peers {
		result += fmt.Sprintf("%s, ", cli.name)
	}
	return result[:len(result)-2] + " online"
}

func exchangeKeysBetweenPeers(peers map[int32]*UserData) {
	peersSlice := make([]*UserData, 0, 2)
	if len(peers) == 2 {
		for _, peer := range peers {
			peersSlice = append(peersSlice, peer)
		}
		peersSlice[0].client <- *peersSlice[1].composeMessage("KEY", peersSlice[1].publicKey)
		peersSlice[1].client <- *peersSlice[0].composeMessage("KEY", peersSlice[0].publicKey)
	}
}
