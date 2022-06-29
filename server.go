package main

import (
	"fmt"
	"log"
	"net"
)

type client chan<- string

type userData struct {
	name       string
	client     chan string
	connection net.Conn
}

var (
	entering = make(chan userData)
	leaving  = make(chan userData)
	messages = make(chan string)
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
	clients := make(map[client]userData)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli.client] = cli
			cli.client <- getClientsList(clients)
		case cli := <-leaving:
			delete(clients, cli.client)
			close(cli.client)
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func getClientsList(clients map[client]userData) string {
	result := ""
	for _, cli := range clients {
		result += fmt.Sprintf("%s, ", cli.name)
	}
	return result[:len(result)-2] + " online"
}
