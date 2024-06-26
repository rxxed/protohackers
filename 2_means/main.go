package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sort"
)

type QueryMessage struct {
	minTime int32
	maxTime int32
}

type Response struct {
	price int32
}

// insert ('I') message
type TimestampPrice struct {
	timestamp int32
	price     int32
}

type TimestampPriceStore []TimestampPrice

// implement the sort.Interface interface
func (tps TimestampPriceStore) Len() int {
	return len(tps)
}
func (tps TimestampPriceStore) Less(i, j int) bool {
	return tps[i].timestamp < tps[j].timestamp
}
func (tps TimestampPriceStore) Swap(i, j int) {
	tps[i], tps[j] = tps[j], tps[i]
}

func (tps *TimestampPriceStore) Insert(tp TimestampPrice) {
	fmt.Println("just inserted")
	*tps = append(*tps, tp)
	sort.Sort(tps)
}
func (tps TimestampPriceStore) Average(mintime, maxtime int32) int32 {
	var sum int32
	var length int32
	for _, tp := range tps {
		if tp.timestamp >= mintime && tp.timestamp <= maxtime {
			length++
			sum += tp.price
		}
	}

	if length == 0 {
		return 0
	}
	return sum / length
}

func main() {
	ln, err := net.Listen("tcp", ":8777")
	if err != nil {
		panic("failed to listen at port")
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic("failed to accept connection from client")
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	tps := TimestampPriceStore{}
	buf := make([]byte, 9)
	for {
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			fmt.Println("read fewer than 9 bytes / eof")
			return
		}

		switch buf[0] {
		case 'I':
			reqMsg := TimestampPrice{}
			reqMsg.timestamp = int32(binary.BigEndian.Uint32(buf[1:5]))
			reqMsg.price = int32(binary.BigEndian.Uint32(buf[5:9]))
			tps.Insert(reqMsg)
		case 'Q':
			reqMsg := QueryMessage{}
			reqMsg.minTime = int32(binary.BigEndian.Uint32(buf[1:5]))
			reqMsg.maxTime = int32(binary.BigEndian.Uint32(buf[5:9]))
			avg := tps.Average(reqMsg.minTime, reqMsg.maxTime)
			ret := make([]byte, 4)
			binary.BigEndian.PutUint32(ret, uint32(avg))
			conn.Write(ret)
		default:
			fmt.Println("unexpected message type: ", buf[0])
			return
		}
	}
}
