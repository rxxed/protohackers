package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

const PrimeMethodName = "isPrime"

// using this type instead of json.Number because json.Number will allow
// numbers with quotes like "66" which should actually be considered a malformed request
type ProtohackersInt int64

func (i *ProtohackersInt) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || b[0] == '"' {
		return fmt.Errorf("rejecting string")
	}
	var f float64
	err := json.Unmarshal(b, &f)
	if err != nil {
		return fmt.Errorf("invalid number")
	}
	*i = ProtohackersInt(f)
	return nil
}

type Request struct {
	Method string `json:"method"`

	// this is a pointer and not a value so we can have a way to determine if there was
	// no number in the request object as this pointer will be `nil` by default.
	Number *ProtohackersInt `json:"number"`
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func (r *Request) IsMalformed() bool {
	if r.Method == "" || r.Number == nil {
		return true
	}
	if r.Method != PrimeMethodName {
		return true
	}
	return false
}

func main() {
	ln, err := net.Listen("tcp", ":8777")
	if err != nil {
		panic("failed to listen on port")
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic("failed to accept connection")
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		req := Request{}
		err := json.Unmarshal([]byte(scanner.Text()), &req)

		if err != nil {
			// malformed request (invalid json)
			// nil is sent as malformed response
			fmt.Fprintln(conn, nil)
			// break will disconnect client
			break
		}

		if req.IsMalformed() {
			// malformed request (missing required fields / wrong method name)
			fmt.Fprintln(conn, nil)
			break
		}

		resp := Response{Method: PrimeMethodName}
		// the unmarshaling of req.Number itself guarantees its a number
		resp.Prime = isPrime(int64(*req.Number))
		// marshaling of resp will never fail, so ignore the error
		respBytes, _ := json.Marshal(resp)

		fmt.Fprintln(conn, string(respBytes))
	}
}

func isPrime(num int64) bool {
	if num <= 1 {
		return false
	}
	for i := int64(2); i*i <= num; i++ {
		if num%i == 0 {
			return false
		}
	}

	return true
}
