package NetCore

import (
	"net"
	"os"
	"time"
)

type TcpServer struct {
	running  bool
	listener *net.TCPListener
}

func (server *TcpServer) Start(addr string, chStop chan string) *TcpServer {
	if server.running && server.listener != nil {
		return server
	}

	go func() {
		var (
			tcpAddr *net.TCPAddr
			err     error
		)

		tcpAddr, err = net.ResolveTCPAddr("tcp4", addr)
		if err != nil {
			LogError(LOG_IDX, LOG_IDX, fmt.Sprintf("ResolveTCPAddr error: %v\n", err)+GetStackInfo())
			chStop <- "TcpServer Start Failed!"
		}

		server.listener, err = net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			LogError(LOG_IDX, LOG_IDX, fmt.Sprintf("Listening error: %v\n", err)+GetStackInfo())
			chStop <- "TcpServer Start Failed!"
		}

		defer listener.Close()

		server.running = true

		LogInfo(LOG_IDX, LOG_IDX, fmt.Sprintf("TcpServer Running on: %s", tcpAddr.String()))

		for {
			conn, err := listener.AcceptTCP()

			if !server.running {
				break
			}
			if err != nil {
				LogInfo(LOG_IDX, LOG_IDX, fmt.Sprintf("Accept error: %v\n", err)+GetStackInfo())
			} else {
				//newClient(conn)
			}
		}
	}()

	return server
}

func (server *TcpServer) Stop() {
	server.running = false
	server.listener.Close()

	for {
		if len(ClientIdMap) == 0 {
			break
		}

		for _, client := range IdClientMap {
			client.ClearAllCloseCB()
			client.Stop()
		}
		time.Sleep(time.Second / 20)
	}

	LogInfo(LOG_IDX, LOG_IDX, "[ShutDown] TcpServer Stop!")
}

func NewTcpServer() *TcpServer {
	return &TcpServer{
		running:  false,
		listener: nil,
	}
}
