package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	fmt.Println("hello, world!")
	tcpListener, err := net.Listen("tcp", ":4242")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create tcp connection: %s", err)
		os.Exit(1)
	}
	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection: %s", err)
			continue
		}

		go func(c net.Conn) {
			io.Copy(c, c)
			c.Close()
		}(conn)
	}
}
