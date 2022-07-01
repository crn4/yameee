package engine

import "fmt"

type broadcast struct {
	entering chan UserData
	leaving  chan UserData
	messages chan Message
}

var (
	broadcastRequest = make(chan string)
	broadcastAnswer  = make(chan *broadcast)
)

func BroadcastManager() {
	go broadcastController(broadcastAnswer, broadcastRequest)
}

func broadcastController(broadcastAnswer chan<- *broadcast, broadcastRequest <-chan string) {
	activeBroadcasts := make(map[string]*broadcast)
	for {
		request, ok := <-broadcastRequest
		if !ok {
			close(broadcastAnswer)
		}
		if data, found := activeBroadcasts[request]; found {
			broadcastAnswer <- data
		} else {
			newbroadcast := newBroadcast()
			activeBroadcasts[request] = newbroadcast
			broadcastAnswer <- newbroadcast
			go startBroadcast(newbroadcast)
		}
	}
}

func newBroadcast() *broadcast {
	entering := make(chan UserData)
	leaving := make(chan UserData)
	messages := make(chan Message)
	return &broadcast{entering, leaving, messages}
}

func startBroadcast(br *broadcast) {
	activeConnections := make(map[string]*Peers)
	for {
		select {
		case msg := <-br.messages:
			activeConnections[msg.chatID].RWMutex.RLock()
			for _, peer := range activeConnections[msg.chatID].peers {
				peer.clientChan <- msg.message
			}
			activeConnections[msg.chatID].RWMutex.RUnlock()
		case cli := <-br.entering:
			peerCurr := &Peer{connection: cli.connection, clientChan: cli.client, name: cli.name, peerID: cli.userID}
			if value, found := activeConnections[cli.chatID]; !found {
				activeConnections[cli.chatID] = &Peers{peers: map[int32]*Peer{cli.userID: peerCurr}}
			} else {
				value.RWMutex.Lock()
				value.peers[cli.userID] = peerCurr
				value.RWMutex.Unlock()
			}
			cli.client <- getNamesByConnection(activeConnections[cli.chatID])
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
