package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
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

		// split buf by newline and send a response for each line
		requests := bytes.Split(buf[:n], []byte{'\n'})

		for _, request := range requests {
			req := new(Request)
			err = json.Unmarshal(request, req)

			fmt.Printf("received request: %s\n", string(request))

			if err != nil || isMalformed(req) {
				fmt.Fprintf(os.Stderr, "malformed request!\n")
				// the value we Write here doesn't matter, a malformed response is any response that
				// 1) isn't a valid json or 2) does not conform to the provided standards
				conn.Write(nil)
				break
			}

			// by this point, parsing to float cannot possibly return a non-nil err.  we can safely ignore it.
			floatNum, _ := req.Number.Float64()
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
}

func isMalformed(req *Request) bool {
	if req.Method == nil || req.Number == nil {
		fmt.Fprintf(os.Stderr, "either method or number were not present.\n")
		return true
	}
	_, err := req.Number.Float64()
	if err != nil || *req.Method != primeMethodName {
		fmt.Fprintf(os.Stderr, "either method was not 'isPrime' or number wasn't a proper number.\n")
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
