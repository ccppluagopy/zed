package main

import (
	"fmt"
	"net"
	"os"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

func main() {
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go HandleConn(conn)
	}
}

func HandleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		reqLen, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Client %s Stop: %s\n", conn.RemoteAddr().String(), err.Error())
			break
		}
		fmt.Println("[Recv]", string(buf[8:reqLen]))

		_, err = conn.Write(buf[:reqLen])
		if err != nil {
			fmt.Printf("Client %s Stop: %s\n", conn.RemoteAddr().String(), err.Error())
			break
		}
		fmt.Println("[Send]", string(buf[8:reqLen]))
	}
}
