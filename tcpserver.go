package zed

import (
	"fmt"
	"net"
	"sync"
)

type HandlerCB func(client *Client, msg *NetMsg) bool

type TcpServer struct {
	sync.RWMutex
	running         bool
	ClientNum       int
	listener        *net.TCPListener
	handlerMap      map[CmdType]HandlerCB
	msgSendCorNum   int
	msgHandleCorNum int

	clientIdMap map[*TcpClient]ClientIDType
	idClientMap map[ClientIDType]*TcpClient
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

		defer server.listener.Close()

		server.running = true

		LogInfo(LOG_IDX, LOG_IDX, fmt.Sprintf("TcpServer Running on: %s", tcpAddr.String()))

		for {
			_, err := server.listener.AcceptTCP()

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

	/*for {
		if len(ClientIdMap) == 0 {
			break
		}

		for _, client := range IdClientMap {
			client.ClearAllCloseCB()
			client.Stop()
		}
		time.Sleep(time.Second / 20)
	}*/

	LogInfo(LOG_IDX, LOG_IDX, "[ShutDown] TcpServer Stop!")
}

func (server *TcpServer) AddMsgHandler(cmd CmdType, cb HandlerCB) {
	Logger.Println(LogConf.NetCoreClient, LogConf.SERVER, "AddMsgHandler", cmd, cb)

	server.handlerMap[cmd] = cb
}

func (server *TcpServer) RemoveMsgHandler(cmd CmdType, cb HandlerCB) {
	delete(server.handlerMap, cmd)
}

func (server *TcpServer) GetClientById(id ClientIDType) *TcpClient {
	server.RLock()
	defer server.RUnlock()

	if c, ok := server.idClientMap[client.Id]; ok {
		return c
	}

	return nil
}

func (server *TcpServer) AddClient(client *TcpClient) {
	if client.Id != NullId {
		server.Lock()
		defer server.Unlock()

		server.idClientMap[client.Id] = client
		server.clientIdMap[client] = client.Id
	}
}

func (server *TcpServer) RemoveClient(client *TcpClient) {
	if client.Id != NullId {
		server.Lock()
		defer server.Unlock()

		delete(server.idClientMap, client.Id)
		delete(server.clientIdMap, client)
	}
}

func (server *TcpServer) GetClientNum(client *TcpClient) (int, int) {
	return len(server.clientIdMap), server.ClientNum
}

func NewTcpServer(msgSendCorNum int, msgHandleCorNum int) *TcpServer {
	return &TcpServer{
		running:         false,
		ClientNum:       0,
		listener:        nil,
		handlerMap:      make(map[CmdType]HandlerCB),
		msgSendCorNum:   msgSendCorNum,
		msgHandleCorNum: msgHandleCorNum,
		clientIdMap:     make(map[*TcpClient]ClientIDType),
		idClientMap:     make(map[ClientIDType]*TcpClient),
	}
}
