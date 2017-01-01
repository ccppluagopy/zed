package zed

import (
	//"encoding/binary"
	//"fmt"
	"net"
	"sync"
	"time"
)

var (
	servers      = make(map[string]*TcpServer)
	serversMutex = sync.Mutex{}
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
			client = newTcpClient(server.delegate, conn, server.ClientNum)
			if client.start() {
				server.ClientNum = server.ClientNum + 1

				//server.onNewClient(client)
				if onnew := server.delegate.NewConnCB(); onnew != nil {
					onnew(client)
				}

				/*for _, cb := range server.newConnCBMap {
					cb(client)
				}*/
			}
		}
	}

}

func (server *TcpServer) ShowClientData() bool {
	return server.showClientData
}

func (server *TcpServer) MaxPackLen() int {
	return server.maxPackLen
}

func (server *TcpServer) RecvBufLen() int {
	return server.recvBufLen
}

func (server *TcpServer) SendBufLen() int {
	return server.sendBufLen
}

func (server *TcpServer) RecvBlockTime() time.Duration {
	return server.recvBlockTime
}

func (server *TcpServer) SendBlockTime() time.Duration {
	return server.sendBlockTime
}

func (server *TcpServer) AliveTime() time.Duration {
	return server.aliveTime
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
	server.listener.Close()
	server.running = false

	for k, _ := range server.handlerMap {
		delete(server.handlerMap, k)
	}

	if server.onStopCB != nil {
		server.onStopCB()
	}

	/*for k, _ := range server.clientIdMap {
		delete(server.clientIdMap, k)
	}
	for k, _ := range server.idClientMap {
		delete(server.idClientMap, k)
	}*/

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

	//server.handlerMap[cmd] = cb
	if server.delegate != nil {
		server.delegate.AddMsgHandler(cmd, cb)
	}
}

func (server *TcpServer) RemoveMsgHandler(cmd CmdType) {
	server.Lock()
	defer server.Unlock()

	ZLog("TcpServer RemoveMsgHandler, Cmd: %d", cmd)

	//delete(server.handlerMap, cmd)
	if server.delegate != nil {
		server.delegate.RemoveMsgHandler(cmd)
	}
}

func (server *TcpServer) SetMsgFilter(filter func(*NetMsg) bool) {
	server.msgFilter = filter
}

func (server *TcpServer) SetNewConnCB(cb func(*TcpClient)) {
	if server.delegate != nil {
		server.delegate.SetNewConnCB(cb)
	}
}

/*func (server *TcpServer) SetConnCloseCB(cb func(*TcpClient)) {
	server.onConnCloseCB = cb
}
*/
func (server *TcpServer) onNewClient(client *TcpClient) {
	/*server.Lock()
	defer server.Unlock()

	server.clients[client] = client*/
}

func (server *TcpServer) SetServerStopCB(cb func()) {
	server.Lock()
	defer server.Unlock()
	server.onStopCB = cb
}

func (server *TcpServer) OnClientMsgError(msg *NetMsg) {

}

func (server *TcpServer) HandleMsg(msg *NetMsg) {
	if server.msgFilter != nil {
		defer PanicHandle(true, func() {
			ZLog("HandleMsg %s panic err!", msg.Client.Info())
			msg.Client.Stop()
		})

		if server.msgFilter(msg) {
			if server.delegate != nil {
				server.delegate.HandleMsg(msg)
			}
		} else {
			msg.Client.Stop()
		}
	}
	/*
		cb, ok := server.handlerMap[msg.Cmd]
		if ok && ((server.msgFilter == nil) || server.msgFilter(msg)) {
			defer PanicHandle(true, func() {
				ZLog("HandleMsg %s panic err!", msg.Client.Info())
				msg.Client.Stop()
			})
			if cb(msg) {
				return
			} else {
				ZLog("HandleMsg Error, %s Msg Cmd: %d, Data: %v.", msg.Client.Info(), msg.Cmd, msg.Data)
			}

		} else {
			ZLog("No Handler For Cmd %d, %s", msg.Cmd, msg.Client.Info())
		}

		msg.Client.Stop()
	*/
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

func (server *TcpServer) SetIOBlockTime(recvBT time.Duration, sendBT time.Duration) {
	if server.delegate != nil {
		server.delegate.SetIOBlockTime(recvBT, sendBT)
	}
}

func (server *TcpServer) SetIOBufLen(recvBL int, sendBL int) {
	if server.delegate != nil {
		server.delegate.SetIOBufLen(recvBL, sendBL)
	}
}

func (server *TcpServer) SetCientAliveTime(aliveT time.Duration) {
	if server.delegate != nil {
		server.delegate.SetCientAliveTime(aliveT)
	}
}

func (server *TcpServer) SetMaxPackLen(maxPL int) {
	if server.delegate != nil {
		server.delegate.SetMaxPackLen(maxPL)
	}
}

func (server *TcpServer) SetDataInSupervisor(dataInSupervisor func(msg *NetMsg)) {
	if server.delegate != nil {
		server.delegate.SetDataInSupervisor(dataInSupervisor)
	}
}

func (server *TcpServer) SetDataOutSupervisor(dataOutSupervisor func(msg *NetMsg)) {
	if server.delegate != nil {
		server.delegate.SetDataOutSupervisor(dataOutSupervisor)
	}
}

func (server *TcpServer) SetShowClientData(show bool) {
	if server.delegate != nil {
		server.delegate.SetShowClientData(show)
	}
}

func (server *TcpServer) SetDelegate(delegate ZTcpClientDelegate) {
	server.Lock()
	defer server.Unlock()
	delegate.Init()
	/*if delegate.AliveTime() == 0 {
		delegate.SetCientAliveTime(DEFAULT_KEEP_ALIVE_TIME)
	}

	if delegate.RecvBlockTime() == 0 {
		delegate.SetRecvBlockTime(DEFAULT_RECV_BLOCK_TIME)
	}

	if delegate.SendBlockTime() == 0 {
		delegate.SetSendBlockTime(DEFAULT_SEND_BLOCK_TIME)
	}

	if delegate.MaxPackLen() == 0 {
		delegate.SetMaxPackLen(DEFAULT_MAX_PACK_LEN)
	}

	if delegate.RecvBufLen() == 0 {
		delegate.SetRecvBufLen(DEFAULT_RECV_BUF_LEN)
	}
	if delegate.SendBufLen() == 0 {
		delegate.SetSendBufLen(DEFAULT_SEND_BUF_LEN)
	}
	*/
	server.delegate = delegate
	delegate.SetServer(server)
}

func (server *TcpServer) RecvMsg(client *TcpClient) *NetMsg {
	if server.delegate != nil {
		msg := server.delegate.RecvMsg(client)
		if msg != nil {
			if server.dataInSupervisor != nil {
				server.dataInSupervisor(msg)
			}
		}
		return msg
	}
	return nil
}

type SendMsgError struct {
}

func (err *SendMsgError) Error() string {
	return "SendMsg Failed Error"
}

func (server *TcpServer) SendMsg(client *TcpClient, msg *NetMsg) {
	if server.delegate != nil {
		/*client.Lock()
		defer client.Unlock()
		if client.running {*/
		//msg.Client = client
		if server.delegate.SendMsg(client, msg) {
			if server.dataOutSupervisor != nil {
				server.dataOutSupervisor(msg)
			}
		} else {
			ZLog("SendMsg Failed, Client: %s %d %d %s", client.Info(), msg.Cmd, msg.Len, string(msg.Data))
			panic(&SendMsgError{})
		}
		//}
	}
}

func NewTcpServer(name string) *TcpServer {
	serversMutex.Lock()
	defer serversMutex.Unlock()

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
		clients:    make(map[*TcpClient]*TcpClient),
		//clientIdMap: make(map[*TcpClient]uint32),
		//idClientMap: make(map[uint32]*TcpClient),
		msgFilter:     nil,
		onNewConnCB:   nil,
		onStopCB:      nil,
		recvBlockTime: DEFAULT_RECV_BLOCK_TIME,
		sendBlockTime: DEFAULT_SEND_BLOCK_TIME,

		aliveTime: DEFAULT_KEEP_ALIVE_TIME,

		recvBufLen: DEFAULT_RECV_BUF_LEN,
		sendBufLen: DEFAULT_SEND_BUF_LEN,

		maxPackLen: DEFAULT_MAX_PACK_LEN,

		dataInSupervisor:  nil,
		dataOutSupervisor: nil,
		showClientData:    false,
	}

	server.SetDelegate(&DefaultTCDelegate{
		Server: server,
	})

	servers[name] = server

	return server
}

func GetTcpServerByName(name string) (*TcpServer, bool) {
	serversMutex.Lock()
	defer serversMutex.Unlock()

	server, ok := servers[name]
	if !ok {
		ZLog("GetTcpServerByName Error: TcpServer %s doesn't exists.", name)
	}
	return server, ok
}
