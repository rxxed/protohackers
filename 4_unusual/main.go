package main

import (
	"fmt"
	"net"
)

const Version = "Ken's Key-Value Store 1.0"

type RequestType int

const (
	Insert RequestType = iota
	Retrieve
)

type Request struct {
	reqType RequestType
	key     []byte
	value   []byte
}

func (r *Request) Parse(buf []byte) {
	valueFlag := false
	for _, byt := range buf {
		if byt == '=' && !valueFlag {
			valueFlag = true
			continue
		}
		if valueFlag {
			r.value = append(r.value, byt)
		} else {
			r.key = append(r.key, byt)
		}
	}
	if !valueFlag { // if an '=' was never found
		r.reqType = Retrieve
	}
}

func main() {
	addr := net.UDPAddr{
		Port: 8777,
		IP:   net.IPv4(0, 0, 0, 0),
	}

	udpconn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer udpconn.Close()

	buf := make([]byte, 1024)
	store := make(map[string]string)
	store["version"] = Version

	for {
		n, clientIP, err := udpconn.ReadFromUDP(buf[:])
		if err != nil {
			fmt.Println("failed to read from udpconn")
			panic(err)
		}

		req := Request{}
		req.Parse(buf[:n])

		if req.reqType == Insert {
			if string(req.key) != "version" { // the "version" value should never be modified by clients
				store[string(req.key)] = string(req.value)
			}
		} else if req.reqType == Retrieve {
			udpconn.WriteToUDP([]byte(fmt.Sprintf("%s=%s", req.key, store[string(req.key)])), clientIP)
		}
	}
}
