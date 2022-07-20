package engine

import "fmt"

type broadcast struct {
	entering  chan *UserData
	leaving   chan *UserData
	handshake chan *UserData
	messages  chan Message
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
			go newbroadcast.startBroadcast()
		}
	}
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
