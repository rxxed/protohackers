package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Request struct {
	Method string      `json:"method"`
	Number json.Number `json:"number"`
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

const PrimeMethodName = "isPrime"

func (r *Request) IsMalformed() bool {
	if r.Method == "" || r.Number == "" {
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
				fmt.Println("malformed request (invalid json)")
				// break will disconnect client
				break
			}

			fmt.Println("req.Method is empty str?: ", req.Method)
			fmt.Println("req.Number is empty str?: ", req.Number)

			if req.IsMalformed() {
				// malformed request (missing required fields or wrong method name)
				// nil is sent as malformed response
				fmt.Fprintln(conn, nil)
				fmt.Println("malformed request (missing required fields or wrong method name)")
				// break will disconnect client
				break
			}

			num, err := req.Number.Int64()
			if err != nil {
				// malformed request (req.Number is not a valid numeric value)
				// nil is sent as malformed response
				fmt.Fprintln(conn, nil)
				fmt.Println("malformed request (req.Number is not a valid numeric value)")
				// break will disconnect client
				break
			}

			resp := Response{Method: PrimeMethodName}
			resp.Prime = isPrime(num)
			// marshaling of resp will never fail, so ignore the error
			respBytes, _ := json.Marshal(resp)

			fmt.Fprintln(conn, string(respBytes))
			fmt.Fprintln(os.Stdout, string(respBytes))
		}
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
