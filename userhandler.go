package main

import (
	"bufio"
	"fmt"
	"net"
)

func handleConn(conn net.Conn) {
	cli := &userData{name: conn.RemoteAddr().String(), client: make(chan string), connection: conn}
	go clientWriter(conn, cli.client)

	getUsername(conn, cli)

	cli.client <- "Welcome, " + cli.name
	messages <- cli.name + " has arrived"
	entering <- *cli

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- cli.name + ">> " + input.Text()
	}
	leaving <- *cli
	messages <- cli.name + " has left"
	conn.Close()
}

func getUsername(conn net.Conn, cli *userData) {
	fmt.Fprint(conn, "pls, enter your name: ")
	input := bufio.NewScanner(conn)
	if input.Scan() {
		cli.name = input.Text()
	}
}
