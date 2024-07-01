package main

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
)

const (
	BudgetChatIP     = "chat.protohackers.com"
	BudgetChatPort   = "16963"
	BoguscoinAddress = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

type Client struct {
	clientConn       net.Conn    // client's connection to our server
	upstreamConn     net.Conn    // the connection started by our mitm server for this client
	upstreamMessages chan string // messages from upstream server
	clientMessages   chan string // messages from client
}

func (c Client) Close() {
	c.clientConn.Close()
	c.upstreamConn.Close()
}

func (c Client) ClientListener() {
	scanner := bufio.NewScanner(c.clientConn)
	for scanner.Scan() {
		msg := rewriteCoinAddress(scanner.Text())
		c.clientMessages <- msg
	}
	close(c.clientMessages)
	c.Close()
}

func (c Client) ClientWriter() {
	for {
		msg, ok := <-c.upstreamMessages
		if !ok { // upstream has probably disconnected
			break
		}
		fmt.Fprintln(c.clientConn, rewriteCoinAddress(msg))
	}
}

func (c Client) UpstreamListener() {
	scanner := bufio.NewScanner(c.upstreamConn)
	for scanner.Scan() {
		msg := rewriteCoinAddress(scanner.Text())
		c.upstreamMessages <- msg
	}
	close(c.upstreamMessages)
	c.Close()
}

func (c Client) UpstreamWriter() {
	for {
		msg, ok := <-c.clientMessages
		if !ok { // client has probably disconnected
			break
		}
		fmt.Fprintln(c.upstreamConn, rewriteCoinAddress(msg))
	}
}

func handleConnection(conn net.Conn) {
	uconn, err := net.Dial("tcp", BudgetChatIP+":"+BudgetChatPort)
	if err != nil {
		panic(fmt.Sprintln("failed to dial: ", err))
	}

	client := Client{
		upstreamConn:     uconn,
		clientConn:       conn,
		clientMessages:   make(chan string),
		upstreamMessages: make(chan string),
	}

	go client.ClientListener()
	go client.ClientWriter()
	go client.UpstreamListener()
	go client.UpstreamWriter()
}

func rewriteCoinAddress(msg string) string {
	msg = strings.TrimSpace(msg)
	msgSlice := strings.Fields(msg)

	addrPattern := "^7[a-zA-Z0-9]{25,34}$"

	for i, word := range msgSlice {
		matched, _ := regexp.MatchString(addrPattern, word)
		if matched {
			msgSlice[i] = BoguscoinAddress
		}
	}

	return strings.Join(msgSlice, " ")
}

func main() {
	ln, err := net.Listen("tcp", ":8777")
	if err != nil {
		panic(fmt.Sprintln("failed to listen: ", err))
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(fmt.Sprintln("failed to accept new connection: ", err))
		}

		go handleConnection(conn)
	}
}
