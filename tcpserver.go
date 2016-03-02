package zed

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

type HandlerCB func(msg *NetMsg) bool

type msgtask struct {
	msgQ chan *NetMsg
}

func (task *msgtask) start4Sender() {
	var (
		msg      *NetMsg
		buf      []byte
		writeLen int
		err      error
	)

	for {
		for {
			msg = <-task.msgQ

			if msg == nil {
				return
			}

			if err = msg.Client.conn.SetWriteDeadline(time.Now().Add(WRITE_BLOCK_TIME)); err != nil {
				LogInfo(LOG_IDX, msg.Client.Idx, "Write Failed Cmd: %d, Len: %d, Buf: %s", msg.Cmd, msg.BufLen, string(msg.Buf))
				LogError(LOG_IDX, msg.Client.Idx, "Client(Id: %s, Addr: %s) SetWriteDeadline Error: %v!", msg.Client.Id, msg.Client.Addr, err)
				msg.Client.Stop()
			}

			buf = make([]byte, PACK_HEAD_LEN+len(msg.Buf))
			binary.LittleEndian.PutUint32(buf, uint32(len(msg.Buf)))
			binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))
			copy(buf[PACK_HEAD_LEN:], msg.Buf)

			writeLen, err = msg.Client.conn.Write(buf)
			LogInfo(LOG_IDX, msg.Client.Idx, "Write Success Cmd: %d, Len: %d, Buf: %s", msg.Cmd, msg.BufLen, string(msg.Buf))

			if err != nil || writeLen != len(buf) {
				msg.Client.Stop()
			}
		}

	}
}

func (task *msgtask) start4Handler(server *TcpServer) {
	var (
		msg *NetMsg
	)

	for {
		msg = <-task.msgQ

		if msg == nil {
			return
		}

		server.HandleMsg(msg)
	}
}

func (task *msgtask) stop() {
	if task.msgQ != nil {
		close(task.msgQ)
	}
}

type TcpServer struct {
	sync.RWMutex
	running    bool
	ClientNum  int
	listener   *net.TCPListener
	handlerMap map[CmdType]HandlerCB

	msgSendCorNum   int
	msgHandleCorNum int
	senders         []*msgtask
	handlers        []*msgtask

	clients     map[int]*TcpClient
	clientIdMap map[*TcpClient]ClientIDType
	idClientMap map[ClientIDType]*TcpClient
}

func (server *TcpServer) startSenders() *TcpServer {
	if server.msgSendCorNum != len(server.senders) {
		server.senders = make([]*msgtask, server.msgSendCorNum)
		for i := 0; i < server.msgSendCorNum; i++ {
			server.senders[i] = &msgtask{msgQ: make(chan *NetMsg, 5)}
			go server.senders[i].start4Sender()
			LogInfo(LOG_IDX, LOG_IDX, "TcpServer startSenders %d.", i)
		}
	}
	return server
}

func (server *TcpServer) stopSenders() *TcpServer {
	for i := 0; i < server.msgSendCorNum; i++ {
		server.senders[i].stop()
		LogInfo(LOG_IDX, LOG_IDX, "TcpServer stopSenders %d / %d.", i, server.msgSendCorNum)
	}

	return server
}

func (server *TcpServer) startHandlers() *TcpServer {
	if server.msgHandleCorNum != len(server.handlers) {
		server.handlers = make([]*msgtask, server.msgHandleCorNum)
		for i := 0; i < server.msgHandleCorNum; i++ {
			server.handlers[i] = &msgtask{msgQ: make(chan *NetMsg, 5)}
			go server.handlers[i].start4Handler(server)
			LogInfo(LOG_IDX, LOG_IDX, "TcpServer startHandlers %d.", i)
		}
	}
	return server
}

func (server *TcpServer) stopHandlers() *TcpServer {
	for i := 0; i < server.msgHandleCorNum; i++ {
		server.handlers[i].stop()
		LogInfo(LOG_IDX, LOG_IDX, "TcpServer stopHandlers %d / %d.", i, server.msgHandleCorNum)
	}

	return server
}

func (server *TcpServer) startListener(addr string) {
	var (
		tcpAddr *net.TCPAddr
		err     error
		client  *TcpClient
	)

	tcpAddr, err = net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		LogError(LOG_IDX, LOG_IDX, fmt.Sprintf("ResolveTCPAddr error: %v\n", err)+GetStackInfo())
		//chStop <- "TcpServer Start Failed!"
		return
	}

	server.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		LogError(LOG_IDX, LOG_IDX, fmt.Sprintf("Listening error: %v\n", err)+GetStackInfo())
		//chStop <- "TcpServer Start Failed!"
		return
	}

	defer server.listener.Close()

	server.running = true

	LogInfo(LOG_IDX, LOG_IDX, fmt.Sprintf("TcpServer Running on: %s", tcpAddr.String()))

	for {
		conn, err := server.listener.AcceptTCP()

		if !server.running {
			break
		}
		if err != nil {
			LogInfo(LOG_IDX, LOG_IDX, fmt.Sprintf("Accept error: %v\n", err)+GetStackInfo())
		} else {
			client = newTcpClient(server, conn)
			if !client.start() {
				server.ClientNum = server.ClientNum - 1
			} else {
				server.clients[client.Idx] = client
			}
		}
	}

	LogInfo(LOG_IDX, LOG_IDX, "TcpServer startListener Stopped...")
}

func (server *TcpServer) Start(addr string) {
	if !server.running {
		server.startSenders()
		server.startHandlers()
		server.startListener(addr)
	}
}

func (server *TcpServer) Stop() {
	LogInfo(LOG_IDX, LOG_IDX, "......... Stop() 111")
	if server.running {
		defer PanicHandle(true, "TcpServer Stop()xx.")

		for idx, client := range server.clients {
			client.Stop()
			delete(server.clients, idx)
		}
		LogInfo(LOG_IDX, LOG_IDX, "......... Stop() 222")
		server.stopHandlers()
		LogInfo(LOG_IDX, LOG_IDX, "......... Stop() 333")
		server.stopSenders()
		LogInfo(LOG_IDX, LOG_IDX, "......... Stop() 444")
		for k, _ := range server.handlerMap {
			delete(server.handlerMap, k)
		}
		LogInfo(LOG_IDX, LOG_IDX, "......... Stop() 555")
		for k, _ := range server.clientIdMap {
			delete(server.clientIdMap, k)
		}
		LogInfo(LOG_IDX, LOG_IDX, "......... Stop() 666")
		for k, _ := range server.idClientMap {
			delete(server.idClientMap, k)
		}
		LogInfo(LOG_IDX, LOG_IDX, "......... Stop() 777")

		server.listener.Close()
		server.running = false
		LogInfo(LOG_IDX, LOG_IDX, "[TcpServer Stop] 888")
	}
}

func (server *TcpServer) AddMsgHandler(cmd CmdType, cb HandlerCB) {
	LogInfo(LOG_IDX, LOG_IDX, "TcpServer AddMsgHandler", cmd, cb)

	server.handlerMap[cmd] = cb
}

func (server *TcpServer) RemoveMsgHandler(cmd CmdType, cb HandlerCB) {
	delete(server.handlerMap, cmd)
}

func (server *TcpServer) RelayMsg(msg *NetMsg) {
	if server.msgHandleCorNum == 0 {
		LogError(LOG_IDX, msg.Client.Idx, "TcpServer RelayMsg Error, msgHandleCorNum is 0.")
		return
	}
	server.handlers[msg.Client.Idx%server.msgHandleCorNum].msgQ <- msg
}

func (server *TcpServer) OnClientMsgError(msg *NetMsg) {
	msg.Client.SendMsg(msg)
}

func (server *TcpServer) HandleMsg(msg *NetMsg) {
	defer PanicHandle(true)

	cb, ok := server.handlerMap[msg.Cmd]
	if ok {
		if cb(msg) {
			return
		} else {
			LogInfo(LOG_IDX, msg.Client.Idx, fmt.Sprintf("HandleMsg Error, Client(Id: %s, Addr: %s) Msg Cmd: %d, Buf: %v.", msg.Client.Id, msg.Client.Addr, msg.Cmd, msg.Buf))
		}
	} else {
		LogInfo(LOG_IDX, msg.Client.Idx, "No Handler For Cmd %d From Client(Id: %s, Addr: %s.", msg.Cmd, msg.Client.Id, msg.Client.Addr)
	}

	server.OnClientMsgError(msg)
}

func (server *TcpServer) SendMsg(msg *NetMsg) {
	if server.msgSendCorNum == 0 {
		LogError(LOG_IDX, msg.Client.Idx, "TcpServer SendMsg Error, msgSendCorNum is 0.")
		return
	}
	server.senders[msg.Client.Idx%server.msgSendCorNum].msgQ <- msg
	LogInfo(LOG_IDX, msg.Client.Idx, "TcpServer SendMsg For Client(Id: %s, Addr: %s.", msg.Cmd, msg.Client.Id, msg.Client.Addr)
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
