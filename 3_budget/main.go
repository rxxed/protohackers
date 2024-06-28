package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

type User struct {
	username string
	id       int64
	conn     net.Conn
}

type Room struct {
	users []User
}

func (room *Room) AddUser(user User) {
	room.users = append(room.users, user)
}

func (room Room) GetOtherUsers(user User) (usernames []string) {
	for _, u := range room.users {
		if u.id != user.id {
			usernames = append(usernames, u.username)
		}
	}

	return
}

func (room Room) SendMessageToOthers(msg string, user User) {
	for _, u := range room.users {
		if user.id != u.id {
			fmt.Fprintf(u.conn, msg)
		}
	}
}

func (room *Room) RemoveUser(user User) {
	var deleteIdx int
	for i, u := range room.users {
		if u.id == user.id {
			deleteIdx = i
			break
		}
	}
	newUsers := make([]User, 0)
	newUsers = append(newUsers, room.users[:deleteIdx]...)
	newUsers = append(newUsers, room.users[deleteIdx+1:]...)
	room.users = newUsers
}

func main() {
	ln, err := net.Listen("tcp", ":8777")
	if err != nil {
		panic("Failed to listen on port")
	}

	room := new(Room)
	var uniqueId int64 = 0

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic("Failed to accept connection.")
		}

		uniqueId++
		go handleNewConnection(conn, room, uniqueId)
	}
}

func handleNewConnection(conn net.Conn, room *Room, uniqueId int64) {
	defer conn.Close()

	fmt.Fprintln(conn, "Welcome to budgetchat! What shall I call you?")
	name := make([]byte, 24)
	n, err := conn.Read(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read from client")
		return
	}

	name = name[:n]
	name = bytes.TrimSpace(name)
	if !isAlphaNumeric(name) || len(name) < 1 {
		fmt.Fprintln(conn, "Username must only contain alphanumeric values.")
		return
	}

	user := User{username: string(name), id: uniqueId, conn: conn}
	room.AddUser(user)

	room.SendMessageToOthers(fmt.Sprintf("* %s has entered the room\n", name), user)
	fmt.Fprintf(conn, "* The room contains: %s\n", strings.Join(room.GetOtherUsers(user), ", "))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		room.SendMessageToOthers(fmt.Sprintf("[%s] %s\n", user.username, scanner.Text()), user)
	}

	room.SendMessageToOthers(fmt.Sprintf("* %s has left the room\n", user.username), user)
	room.RemoveUser(user)
}

func isAlphaNumeric(bytes []byte) bool {
	for _, byt := range bytes {
		if !((byt >= '0' && byt <= '9') || (byt >= 'a' && byt <= 'z') || (byt >= 'A' && byt <= 'Z')) {
			return false
		}
	}
	return true
}
