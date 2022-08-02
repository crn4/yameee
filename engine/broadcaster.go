package engine

import (
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

func (s *server) Start(port string) error {
	http.HandleFunc("/ws", s.userRegistrator)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) broadcastController(chatID string) *broadcast {
	s.rwMutex.RLock()
	wlock := false
	defer func() {
		if wlock {
			s.rwMutex.Unlock()
		} else {
			s.rwMutex.RUnlock()
		}
	}()
	broadcast, found := s.broadcasts[chatID]
	if !found {
		s.rwMutex.RUnlock()
		s.rwMutex.Lock()
		wlock = true
		broadcast, found = s.broadcasts[chatID]
		if !found {
			broadcast = newBroadcast()
			s.broadcasts[chatID] = broadcast
			go broadcast.startBroadcast()
		}
	}
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
	name := r.URL.Query().Get("name")
	if len(name) == 0 {
		log.Println("Name was not send")
		conn.Close()
		return
	}
	chatID := r.URL.Query().Get("chatID")
	if len(chatID) == 0 {
		log.Println("chatID was not send")
		conn.Close()
		return
	}
	br := s.broadcastController(chatID)
	cli := NewClient(conn, br, name, chatID)
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

// мапа activeConnections - рудимент. когда я еще не дошел до бродкастера, она содержала в себе Peers по ключу chatID
// теперь это не актуально, так как сам startBroadcaster запускается для каждого chatID
func (br *broadcast) startBroadcast() {
	peersList := &Peers{peers: make(map[*websocket.Conn]*UserData)}
	for {
		select {
		case msg := <-br.messages:
			peersList.RWMutex.RLock()
			for _, peer := range peersList.peers {
				peer.client <- msg
			}
			peersList.RWMutex.RUnlock()
		case cli := <-br.entering:
			peersList.RWMutex.Lock()
			peersList.peers[cli.connection] = cli
			peersList.RWMutex.Unlock()
			cli.client <- cli.composeMessage(envelopeTypeService, peersList.getNamesByConnection())
		case <-br.handshake:
			peersList.RWMutex.RLock()
			if len(peersList.peers) == 2 {
				peersList.exchangeKeysBetweenPeers()
			}
			peersList.RWMutex.RUnlock()
		case cli := <-br.leaving:
			peersList.RWMutex.Lock()
			delete(peersList.peers, cli.connection)
			close(cli.client)
			peersList.RWMutex.Unlock()
		}
	}
}
