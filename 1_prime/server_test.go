package main

import (
	"encoding/json"
	"testing"
)

func strPtr(s string) *string {
	return &s
}

func jsonNumberPtr(n string) *json.Number {
	num := json.Number(n)
	return &num
}

func TestIsMalformed(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		want    bool
	}{
		{
			name: "valid request",
			request: Request{
				Method: strPtr("isPrime"),
				Number: jsonNumberPtr("123"),
			},
			want: false,
		},
		{
			name: "nil method",
			request: Request{
				Method: nil,
				Number: jsonNumberPtr("123"),
			},
			want: true,
		},
		{
			name: "nil number",
			request: Request{
				Method: strPtr("isPrime"),
				Number: nil,
			},
			want: true,
		},
		{
			name: "empty method",
			request: Request{
				Method: strPtr(""),
				Number: jsonNumberPtr("123"),
			},
			want: true,
		},
		{
			name: "non-numeric number",
			request: Request{
				Method: strPtr("isPrime"),
				Number: jsonNumberPtr("abc"),
			},
			want: true,
		},
		{
			name: "incorrect method name",
			request: Request{
				Method: strPtr("getPrime"),
				Number: jsonNumberPtr("123"),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMalformed(&tt.request); got != tt.want {
				t.Errorf("isMalformed() = %v, want %v", got, tt.want)
			}
		})
	}
}
