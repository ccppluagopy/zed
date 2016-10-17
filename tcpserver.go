package zed

import (
	//"encoding/binary"
	//"fmt"
	"net"
	//"sync"
	//"time"
)

var (
	servers = make(map[string]*TcpServer)
)

func (server *TcpServer) startListener(addr string) {
	defer Println("TcpServer Stopped.")
	var (
		tcpAddr *net.TCPAddr
		err     error
		client  *TcpClient
	)

	tcpAddr, err = net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		//LogInfo(LOG_IDX, LOG_IDX, "ResolveTCPAddr error: %v", err)
		ZLog("ResolveTCPAddr error: %v", err)
		return
	}

	server.listener, err = net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		//LogError(LOG_IDX, LOG_IDX, "Listening error: %v", err)
		ZLog("Listening error: %v", err)
		return
	}

	defer server.listener.Close()

	server.running = true

	ZLog("TcpServer Running on: %s", tcpAddr.String())
	//LogInfo(LOG_IDX, LOG_IDX, "TcpServer Running on: %s", tcpAddr.String())
	for {
		conn, err := server.listener.AcceptTCP()

		if !server.running {
			break
		}
		if err != nil {
			//LogInfo(LOG_IDX, LOG_IDX, "TcpServer Accept error: %v\n", err)
			ZLog("TcpServer Accept error: %v\n", err)
		} else {
			client = newTcpClient(server, conn)
			if client.start() {
				server.ClientNum = server.ClientNum + 1

				server.onNewClient(client)
				if server.onNewConnCB != nil {
					server.onNewConnCB(client)
				}

				/*for _, cb := range server.newConnCBMap {
					cb(client)
				}*/
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
	server.Lock()
	defer server.Unlock()
	defer PanicHandle(true, "TcpServer Stop().")

	if !server.running {
		return
	}

	for idx, client := range server.clients {
		client.Stop()
		delete(server.clients, idx)
	}
	//server.stopHandlers()
	//server.stopSenders()
	for k, _ := range server.handlerMap {
		delete(server.handlerMap, k)
	}
	/*for k, _ := range server.clientIdMap {
		delete(server.clientIdMap, k)
	}
	for k, _ := range server.idClientMap {
		delete(server.idClientMap, k)
	}*/

	server.running = false
	server.listener.Close()
}

/*func (server *TcpServer) AddNewConnCB(name string, cb func(client *TcpClient)) {
	server.Lock()
	defer server.Unlock()

	ZLog("TcpServer AddNewConnCB, name: %s", name)

	server.newConnCBMap[name] = cb
}

func (server *TcpServer) RemoveNewConnCB(name string) {
	server.Lock()
	defer server.Unlock()

	ZLog("TcpServer RemoveNewConnCB, name: %s", name)

	delete(server.newConnCBMap, name)
}
*/
func (server *TcpServer) AddMsgHandler(cmd CmdType, cb MsgHandler) {
	server.Lock()
	defer server.Unlock()

	ZLog("TcpServer AddMsgHandler, Cmd: %d", cmd)

	/*server.handlerMap[cmd] = func(msg *NetMsg) bool {
		defer PanicHandle(true)
		return cb(msg)
	}*/
	server.handlerMap[cmd] = cb
}

func (server *TcpServer) RemoveMsgHandler(cmd CmdType, cb MsgHandler) {
	server.Lock()
	defer server.Unlock()

	ZLog("TcpServer RemoveMsgHandler, Cmd: %d", cmd)

	delete(server.handlerMap, cmd)
}

func (server *TcpServer) SetMsgFilter(filter func(*NetMsg) bool) {
	server.msgFilter = filter
}

func (server *TcpServer) SetNewConnCB(cb func(*TcpClient)) {
	server.onNewConnCB = cb
}

func (server *TcpServer) SetConnCloseCB(cb func(*TcpClient)) {
	server.onConnCloseCB = cb
}

func (server *TcpServer) onNewClient(client *TcpClient) {
	server.Lock()
	defer server.Unlock()

	server.clients[client.Idx] = client
}

func (server *TcpServer) onClientStop(client *TcpClient) {
	server.Lock()
	defer server.Unlock()

	delete(server.clients, client.Idx)
	if server.onNewConnCB != nil {
		server.onNewConnCB(client)
	}
}

func (server *TcpServer) OnClientMsgError(msg *NetMsg) {

}

func (server *TcpServer) HandleMsg(msg *NetMsg) {
	//defer PanicHandle(true)

	//server.RLock()
	//defer server.RUnlock()

	cb, ok := server.handlerMap[msg.Cmd]
	if ok && ((server.msgFilter == nil) || server.msgFilter(msg)) {
		defer PanicHandle(true, func() {
			LogError(LOG_IDX, LOG_IDX, "HandleMsg %s panic err!", msg.Client.Info())
			msg.Client.Stop()
		})
		if cb(msg) {
			return
		} else {
			LogError(LOG_IDX, msg.Client.Idx, "HandleMsg Error, %s Msg Cmd: %d, Data: %v.", msg.Client.Info(), msg.Cmd, msg.Data)
		}

	} /*else {
		LogError(LOG_IDX, msg.Client.Idx, "No Handler For Cmd %d From %s", msg.Cmd, msg.Client.Info())
	}*/

	//server.OnClientMsgError(msg)
	msg.Client.Stop()
}

/*
func (server *TcpServer) GetClientById(id ClientIDType) *TcpClient {
	server.RLock()
	defer server.RUnlock()

	if c, ok := server.idClientMap[id]; ok {
		return c
	}

	return nil
}

func (server *TcpServer) AddClient(client *TcpClient) {
	if client.ID != NullID {
		server.Lock()
		defer server.Unlock()

		server.idClientMap[client.ID] = client
		server.clientIdMap[client] = client.ID
	}
}

func (server *TcpServer) RemoveClient(client *TcpClient) {
	//if client.ID != NullID {
	server.Lock()
	defer server.Unlock()

	delete(server.idClientMap, client.ID)
	delete(server.clientIdMap, client)
	//}
}

func (server *TcpServer) GetClientNum(client *TcpClient) (int, int) {
	server.RLock()
	defer server.RUnlock()

	return len(server.clientIdMap), server.ClientNum
}
*/

func NewTcpServer(name string) *TcpServer {
	if _, ok := servers[name]; ok {
		ZLog("NewTcpServer Error: TcpServer %s already exists.", name)
		return nil
	}

	server := &TcpServer{
		running:   false,
		ClientNum: 0,
		listener:  nil,
		//newConnCBMap: make(map[string]func(client *TcpClient)),
		handlerMap: make(map[CmdType]MsgHandler),
		clients:    make(map[int]*TcpClient),
		//clientIdMap: make(map[*TcpClient]uint32),
		//idClientMap: make(map[uint32]*TcpClient),
		msgFilter:     nil,
		onNewConnCB:   nil,
		onConnCloseCB: nil,
	}

	servers[name] = server

	return server
}

func GetTcpServerByName(name string) (*TcpServer, bool) {
	server, ok := servers[name]
	if !ok {
		ZLog("GetTcpServerByName Error: TcpServer %s doesn't exists.", name)
	}
	return server, ok
}
