package main

import (
	"bufio"
	"fmt"
	"net"
)

func handleConn(conn net.Conn) {
	cli := NewClient(conn)
	go clientWriter(conn, cli.client)

	if !getUsername(conn, cli) && cli.name == "" {
		fmt.Fprintf(conn, "problems joining. you are disconnected")
		conn.Close()
		return
	}

	if !getChatID(conn, cli) && cli.chatID == "" {
		fmt.Fprintf(conn, "invalid chatID. you are disconnecting")
		conn.Close()
		return
	}

	cli.client <- "Welcome, " + cli.name
	entering <- *cli
	messages <- *composeMessage(cli.chatID, cli.name+" has arrived")

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- *composeMessage(cli.chatID, cli.name+">> "+input.Text())
	}
	leaving <- *cli
	messages <- *composeMessage(cli.chatID, cli.name+" has left")
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func getUsername(conn net.Conn, cli *UserData) bool {
	fmt.Fprint(conn, "pls, enter your name: ")
	input := bufio.NewScanner(conn)
	if input.Scan() {
		if err := cli.SetName(input.Text()); err != nil {
			return false
		}
	}
	return true
}

func getChatID(conn net.Conn, cli *UserData) bool {
	fmt.Fprint(conn, "pls, enter your chatID: ")
	input := bufio.NewScanner(conn)
	if input.Scan() {
		if err := cli.SetChatID(input.Text()); err != nil {
			return false
		}
	}
	return true
}

func composeMessage(chatid string, message string) *Message {
	return &Message{chatid, message}
}
