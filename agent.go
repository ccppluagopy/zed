package zed

import (
	"fmt"
	"net"
)

func newAgentConn(clientConn *net.TCPConn, serverAddr string) {
	serverConn, dailErr := net.Dial("tcp", serverAddr)

	if dailErr == nil {
		c2sCor := func() {
			defer func() {
				_ = recover()
			}()

			var nread int
			var nwrite int
			var err error
			var buf = make([]byte, 1024)
			for {
				nread, err = clientConn.Read(buf)
				if err != nil {
					clientConn.Close()
					serverConn.Close()
					break
				}

				nwrite, err = serverConn.Write(buf[:nread])
				if nwrite != nread || err != nil {
					clientConn.Close()
					serverConn.Close()
					break
				}
			}
		}

		s2cCor := func() {
			defer func() {
				_ = recover()
			}()

			var nread int
			var nwrite int
			var err error
			var buf = make([]byte, 1024)
			for {
				nread, err = serverConn.Read(buf)
				if err != nil {
					clientConn.Close()
					serverConn.Close()
					break
				}

				nwrite, err = clientConn.Write(buf[:nread])
				if nwrite != nread || err != nil {
					clientConn.Close()
					serverConn.Close()
					break
				}
			}
		}

		go c2sCor()
		go s2cCor()
	}
}

func RunAgent(agentAddr string, serverAddr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", agentAddr)
	if err != nil {
		fmt.Println("ResolveTCPAddr Error: ", err)
		return
	}

	listener, err2 := net.ListenTCP("tcp", tcpAddr)
	if err2 != nil {
		fmt.Println("ListenTCP Error: ", err2)
		return
	}

	defer listener.Close()

	fmt.Println(fmt.Sprintf("Agent Start Running on: Agent(%s) -> Server(%s)!", agentAddr, serverAddr))
	for {
		conn, err := listener.AcceptTCP()

		if err != nil {
			fmt.Println("AcceptTCP Error: ", err2)
		} else {
			go newAgentConn(conn, serverAddr)
		}
	}
}

func TestAgent() {
	Run("127.0.0.1:8888", "127.0.0.1:9999")
}
