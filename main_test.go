package main

import (
	"fmt"
	"io"
	"net"
)

type echo struct {
	Name string
}

// echoService send whatever it receives
type echoMessage struct {
	Msg string
}

// Start starts a tcp server on random port and accepts only one client
// The tcp server responses any data received from client
// It returns the port to which the server is binding
func (e *echo) Start() int {
	listener, err := net.Listen("tcp", ":0")
	checkErr(err)
	go func() {
		conn, err := listener.Accept()
		checkErr(err)
		_, err = io.Copy(conn, conn)
		if err != nil {
			if err == io.EOF {
				return
			}

			fmt.Println("echo: error while reading or writing, ", err.Error())
		}
	}()
	return listener.Addr().(*net.TCPAddr).Port
}

// checkErr panics if err is not null
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
