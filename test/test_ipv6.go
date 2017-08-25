package main 

import(
	"fmt"
	"net"
	"time"
)

const(
	ipv6addr = "[::]:3333"
)

func client() {
	
	tcpAddr, err := net.ResolveTCPAddr("tcp6", ipv6addr)
	if err != nil {
		fmt.Printf("TcpClient Connect ResolveTCPAddr Failed, err: %s, Addr: %s\n", err, tcpAddr)
		return
	}
	conn, err := net.DialTCP("tcp6", nil, tcpAddr)
	if err != nil {
		fmt.Printf("TcpClient Connect ResolveTCPAddr Failed, err: %s, Addr: %s\n", err, tcpAddr)
		return
	}
	buf := []byte("hehehehe")
	for {
		time.Sleep(time.Second)
		_, err = conn.Write(buf)
		if err != nil {
			fmt.Printf("Client %s Stop: %s\n", conn.RemoteAddr().String(), err.Error())
			break
		}
		fmt.Println("[Client Send]", string(buf))

		reqLen, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Client %s Stop: %s\n", conn.RemoteAddr().String(), err.Error())
			break
		}
		fmt.Println("[Client Recv]", string(buf[:reqLen]))
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		reqLen, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Client %s Stop: %s\n", conn.RemoteAddr().String(), err.Error())
			break
		}
		fmt.Println("[Server Recv]", string(buf[:reqLen]))

		_, err = conn.Write(buf[:reqLen])
		if err != nil {
			fmt.Printf("Client %s Stop: %s\n", conn.RemoteAddr().String(), err.Error())
			break
		}
		fmt.Println("[Server Send]", string(buf[:reqLen]))
	}
}

func main() {
	ln, err := net.Listen("tcp6", ipv6addr)
	if err != nil {
		fmt.Println("Listen err: ", err)
	}
	defer ln.Close()
	
	go client()

	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept err: ", err)
		} else {
			go handleConn(c) 
		}
	}
}
