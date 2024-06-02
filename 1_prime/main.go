package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

type Request struct {
	Method *string      `json:"method"`
	Number *json.Number `json:"number"`
}

type Response struct {
	Method  string `json:"method"`
	IsPrime bool   `json:"prime"`
}

const primeMethodName = "isPrime"

func main() {
	tcpListener, err := net.Listen("tcp", ":4242")
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to listen\n")
			continue
		}

		go connHandler(conn)
	}
}

func connHandler(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 65536)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connection closed: %v\n", err)
			break
		}

		req := new(Request)
		err = json.Unmarshal(buf[:n], req)

		if err != nil || isMalformed(req) {
			fmt.Fprintf(os.Stderr, "malformed request!\n")
			// the value we Write here doesn't matter, a malformed response is any response that
			// 1) isn't a valid json or 2) does not conform to the provided standards
			conn.Write(nil)
			break
		}

		// at this point, ParseFloat cannot possibly return a non-nil err.  we can safely ignore it.
		floatNum, _ := strconv.ParseFloat(string(*req.Number), 32)
		num := int(floatNum)

		resp := new(Response)
		resp.Method = primeMethodName
		resp.IsPrime = isPrime(num)

		jsonBytes, err := json.Marshal(resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal response: %v\n", resp)
		}

		_, err = conn.Write(append(jsonBytes, '\n'))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", string(jsonBytes), err)
		}
	}
}

func isMalformed(req *Request) bool {
	if req.Method == nil || req.Number == nil {
		return true
	}
	_, err := strconv.Atoi(string(*req.Number))
	if err != nil || *req.Method != primeMethodName {
		return true
	}
	return false
}

func isPrime(num int) bool {
	if num == 2 {
		return true
	}
	if num < 2 || num%2 == 0 {
		return false
	}
	for i := 3; i*i <= num; i += 2 {
		if num%i == 0 {
			return false
		}
	}
	return true
}
