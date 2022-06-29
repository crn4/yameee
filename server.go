package main

import (
	"fmt"
	"log"
	"net"
)

var (
	entering = make(chan UserData)
	leaving  = make(chan UserData)
	messages = make(chan Message)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}

func broadcaster() {
	activeConnections := make(map[string]*Peers)
	for {
		select {
		case msg := <-messages:
			activeConnections[msg.chatID].RWMutex.RLock()
			for _, peer := range activeConnections[msg.chatID].peers {
				peer.clientChan <- msg.message
			}
			activeConnections[msg.chatID].RWMutex.RUnlock()
		case cli := <-entering:
			peerCurr := &Peer{connection: &cli.connection, clientChan: cli.client, name: cli.name, peerID: cli.userID}
			if value, found := activeConnections[cli.chatID]; !found {
				activeConnections[cli.chatID] = &Peers{peers: map[int32]*Peer{cli.userID: peerCurr}}
			} else {
				value.RWMutex.Lock()
				value.peers[cli.userID] = peerCurr
				value.RWMutex.Unlock()
			}
			cli.client <- getNamesByConnection(activeConnections[cli.chatID])
		case cli := <-leaving:
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
