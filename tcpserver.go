package zed

import (
	//"encoding/binary"
	//"fmt"
	"net"
	"sync"
	//"time"
)

type TcpServer struct {
	sync.RWMutex
	running     bool
	ClientNum   int
	listener    *net.TCPListener
	handlerMap  map[CmdType]MsgHandler
	clients     map[int]*TcpClient
	clientIdMap map[*TcpClient]ClientIDType
	idClientMap map[ClientIDType]*TcpClient
}

func (server *TcpServer) startListener(addr string) {
	defer Println("TcpServer Stopped.")
	var (
		tcpAddr *net.TCPAddr
		err     error
		client  *TcpClient
	)

	tcpAddr, err = net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		LogInfo(LOG_IDX, LOG_IDX, "ResolveTCPAddr error: %v", err)
		return
	}

	server.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		LogError(LOG_IDX, LOG_IDX, "Listening error: %v", err)
		return
	}

	defer server.listener.Close()

	server.running = true

	ZLog("TcpServer Running on: %s", tcpAddr.String())
	LogInfo(LOG_IDX, LOG_IDX, "TcpServer Running on: %s", tcpAddr.String())
	for {
		conn, err := server.listener.AcceptTCP()

		if !server.running {
			break
		}
		if err != nil {
			LogInfo(LOG_IDX, LOG_IDX, "Accept error: %v\n", err)
		} else {
			client = newTcpClient(server, conn)
			if client.start() {
				server.ClientNum = server.ClientNum + 1
				server.clients[client.Idx] = client
				client.AddCloseCB(0, func(client *TcpClient) {
					delete(server.clients, client.Idx)
				})
			}
		}
	}

}

func (server *TcpServer) Start(addr string) {
	server.Lock()
	running := server.running
	if !server.running {
		server.running = true
	}
	server.Unlock()

	if !running {
		server.startListener(addr)
	} else {

	}
}

func (server *TcpServer) Stop() {

	if server.running {
		defer PanicHandle(true, "TcpServer Stop().")

		for idx, client := range server.clients {
			client.Stop()
			delete(server.clients, idx)
		}
		//server.stopHandlers()
		//server.stopSenders()
		for k, _ := range server.handlerMap {
			delete(server.handlerMap, k)
		}
		for k, _ := range server.clientIdMap {
			delete(server.clientIdMap, k)
		}
		for k, _ := range server.idClientMap {
			delete(server.idClientMap, k)
		}

		server.listener.Close()
		server.running = false
	}
}

func (server *TcpServer) AddMsgHandler(cmd CmdType, cb MsgHandler) {
	server.Lock()
	defer server.Unlock()

	LogInfo(LOG_IDX, LOG_IDX, "TcpServer AddMsgHandler, Cmd: %d", cmd)

	server.handlerMap[cmd] = func(msg *NetMsg) bool {
		defer PanicHandle(true)
		return cb(msg)
	}
}

func (server *TcpServer) RemoveMsgHandler(cmd CmdType, cb MsgHandler) {
	server.Lock()
	defer server.Unlock()

	delete(server.handlerMap, cmd)
}

func (server *TcpServer) OnClientMsgError(msg *NetMsg) {
	msg.Client.SendMsg(msg)
}

func (server *TcpServer) HandleMsg(msg *NetMsg) {
	//defer PanicHandle(true)

	//server.RLock()
	//defer server.RUnlock()

	cb, ok := server.handlerMap[msg.Cmd]
	if ok {
		if cb(msg) {
			return
		} else {
			LogError(LOG_IDX, msg.Client.Idx, "HandleMsg Error, Client(Id: %s, Addr: %s) Msg Cmd: %d, Buf: %v.", msg.Client.Id, msg.Client.Addr, msg.Cmd, msg.Buf)
		}
	} else {
		LogError(LOG_IDX, msg.Client.Idx, "No Handler For Cmd %d From Client(Id: %s, Addr: %s)", msg.Cmd, msg.Client.Id, msg.Client.Addr)
	}

	//Println("HandleMsg ==>>")
	server.OnClientMsgError(msg)
}

func (server *TcpServer) GetClientById(id ClientIDType) *TcpClient {
	server.RLock()
	defer server.RUnlock()

	if c, ok := server.idClientMap[id]; ok {
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

func NewTcpServer() *TcpServer {
	return &TcpServer{
		running:     false,
		ClientNum:   0,
		listener:    nil,
		handlerMap:  make(map[CmdType]MsgHandler),
		clients:     make(map[int]*TcpClient),
		clientIdMap: make(map[*TcpClient]ClientIDType),
		idClientMap: make(map[ClientIDType]*TcpClient),
	}
}
