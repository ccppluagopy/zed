package zed

import (
	"encoding/binary"
	"io"
	"sync"
	"time"
)

type ZTcpClientDelegate interface {
	Init()

	RecvMsg(*TcpClient) *NetMsg
	SendMsg(*TcpClient, *NetMsg) bool
	HandleMsg(*NetMsg)

	SetServer(*TcpServer)
	AddMsgHandler(cmd CmdType, cb MsgHandler)
	RemoveMsgHandler(cmd CmdType)
	SetShowClientData(bool)
	SetDataInSupervisor(func(msg *NetMsg))
	SetDataOutSupervisor(func(msg *NetMsg))
	SetIOBlockTime(time.Duration, time.Duration)
	SetRecvBlockTime(time.Duration)
	SetSendBlockTime(time.Duration)
	SetIOBufLen(int, int)
	SetRecvBufLen(int)
	SetSendBufLen(int)
	NoDelay() bool
	SetNoDelay(bool)
	KeepAlive() bool
	SetKeepAlive(bool)
	SetCientAliveTime(time.Duration)
	SetMaxPackLen(int)
	OnNewConn(*TcpClient)
	//SetNewConnCB(cb func(*TcpClient))

	ShowClientData() bool
	MaxPackLen() int
	RecvBufLen() int
	SendBufLen() int
	RecvBlockTime() time.Duration
	SendBlockTime() time.Duration
	AliveTime() time.Duration

	OnServerStop()
}

type DefaultTCDelegate struct {
	sync.Mutex
	//Mutex
	inited     bool
	Server     *TcpServer
	HandlerMap map[CmdType]MsgHandler

	showClientData    bool
	maxPackLen        int
	recvBlockTime     time.Duration
	recvBufLen        int
	sendBlockTime     time.Duration
	sendBufLen        int
	aliveTime         time.Duration
	noDelay           bool
	keepAlive         bool
	delegate          ZTcpClientDelegate
	dataInSupervisor  func(*NetMsg)
	dataOutSupervisor func(*NetMsg)
	newConnCB         func(*TcpClient)

	tag string
}

func (dele *DefaultTCDelegate) RecvMsg(client *TcpClient) *NetMsg {
	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
		msg     *NetMsg
		//server  = dele.Server
	)

	if err = (*client.conn).SetReadDeadline(time.Now().Add(dele.recvBlockTime)); err != nil {
		if dele.showClientData {
			ZLog("RecvMsg %s SetReadDeadline Err: %v.", client.Info(), err)
		}
		goto Exit
	}

	readLen, err = io.ReadFull(client.conn, head)
	if err != nil || readLen < PACK_HEAD_LEN {
		if dele.showClientData {
			ZLog("RecvMsg %s Read Head Err: %v %d.", client.Info(), err, readLen)
		}
		goto Exit
	}

	if err = (*client.conn).SetReadDeadline(time.Now().Add(dele.recvBlockTime)); err != nil {
		if dele.showClientData {
			ZLog("RecvMsg %s SetReadDeadline Err: %v.", client.Info(), err)
		}
		goto Exit
	}

	msg = &NetMsg{
		Cmd:    CmdType(binary.LittleEndian.Uint32(head[4:8])),
		Len:    int(binary.LittleEndian.Uint32(head[0:4])),
		Client: client,
	}
	if msg.Len > dele.maxPackLen {
		ZLog("RecvMsg Read Body Err: Body Len(%d) > MAXPACK_LEN(%d)", msg.Len, dele.maxPackLen)
		goto Exit
	}
	if msg.Len > 0 {
		msg.Data = make([]byte, msg.Len)
		readLen, err := io.ReadFull(client.conn, msg.Data)
		if err != nil || readLen != int(msg.Len) {
			if dele.showClientData {
				ZLog("RecvMsg %s Read Body Err: %v.", client.Info(), err)
			}
			goto Exit
		}
	}

	return msg

Exit:
	return nil
}

func (dele *DefaultTCDelegate) SendMsg(client *TcpClient, msg *NetMsg) bool {
	var (
		writeLen = 0
		buf      []byte
		err      error
	)

	msg.Len = len(msg.Data)

	if msg.Len > dele.maxPackLen {
		if dele.showClientData {
			ZLog("SendMsg Error: Body Len(%d) > MAXPACK_LEN(%d)", msg.Len, dele.maxPackLen)
		}
		goto Exit
	}

	if err := (*client.conn).SetWriteDeadline(time.Now().Add(dele.sendBlockTime)); err != nil {
		if dele.showClientData {
			ZLog("%s SetWriteDeadline Error: %v.", client.Info(), err)
		}
		goto Exit
	}

	buf = make([]byte, PACK_HEAD_LEN+msg.Len)
	binary.LittleEndian.PutUint32(buf, uint32(msg.Len))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))
	if msg.Len > 0 {
		copy(buf[PACK_HEAD_LEN:], msg.Data)
	}

	writeLen, err = client.conn.Write(buf)

	if err == nil && writeLen == len(buf) {
		if dele.showClientData {
			ZLog("[Send] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
		}
		return true
	}

Exit:
	if dele.showClientData {
		ZLog("[Send] %s Cmd: %d Len: %d, Error: %s", client.Info(), msg.Cmd, msg.Len, err.Error())
	}
	client.Stop()
	return false
}

/*
func (dele *DefaultTCDelegate) MsgFilter(msg *NetMsg) bool {
	Println("DefaultTCDelegate MsgFilter 00000000000000000000000")
	server := dele.Server
	if server != nil && server.msgFilter != nil {
		return server.msgFilter(msg)
	}
	return true
}
*/
func (dele *DefaultTCDelegate) Init() {
	dele.Lock()
	inited := dele.inited
	dele.Unlock()
	if inited {
		return
	}
	dele.Lock()
	dele.inited = true
	dele.Unlock()

	if dele.AliveTime() == 0 {
		dele.SetCientAliveTime(DEFAULT_KEEP_ALIVE_TIME)
	}

	if dele.RecvBlockTime() == 0 {
		dele.SetRecvBlockTime(DEFAULT_RECV_BLOCK_TIME)
	}

	if dele.SendBlockTime() == 0 {
		dele.SetSendBlockTime(DEFAULT_SEND_BLOCK_TIME)
	}

	if dele.MaxPackLen() == 0 {
		dele.SetMaxPackLen(DEFAULT_MAX_PACK_LEN)
	}

	if dele.RecvBufLen() == 0 {
		dele.SetRecvBufLen(DEFAULT_RECV_BUF_LEN)
	}
	if dele.SendBufLen() == 0 {
		dele.SetSendBufLen(DEFAULT_SEND_BUF_LEN)
	}

	dele.tag = "DefaultTCDelegate"

	dele.SetShowClientData(false)
}

func (dele *DefaultTCDelegate) HandleMsg(msg *NetMsg) {
	/*defer HandlePanic(true, func() {
		ZLog("HandleMsg %s panic err!", msg.Client.Info())
		msg.Client.Stop()
	})*/

	cb, ok := dele.HandlerMap[msg.Cmd]
	if ok {
		if cb(msg) {
			return
		} else {
			if dele.showClientData {
				ZLog("%s HandleMsg Error, %s Msg Cmd: %d, Data: %v.", dele.tag, msg.Client.Info(), msg.Cmd, msg.Data)
			}
			msg.Client.Stop()
		}
	} else {
		if dele.showClientData {
			ZLog("%s HandleMsg Error: No Handler For Cmd %d, %s", dele.tag, msg.Cmd, msg.Client.Info())
		}
	}
}

func (dele *DefaultTCDelegate) AddMsgHandler(cmd CmdType, cb MsgHandler) {
	dele.Lock()
	defer dele.Unlock()

	ZLog("%s AddMsgHandler, Cmd: %d", dele.tag, cmd)
	if dele.HandlerMap == nil {
		dele.HandlerMap = make(map[CmdType]MsgHandler)
	}
	dele.HandlerMap[cmd] = cb
}

func (dele *DefaultTCDelegate) RemoveMsgHandler(cmd CmdType) {
	dele.Lock()
	defer dele.Unlock()

	ZLog("%s RemoveMsgHandler, Cmd: %d", dele.tag, cmd)

	delete(dele.HandlerMap, cmd)
}

func (dele *DefaultTCDelegate) SetServer(server *TcpServer) {
	dele.Server = server
}

func (dele *DefaultTCDelegate) SetIOBlockTime(recvBT time.Duration, sendBT time.Duration) {
	dele.Lock()
	defer dele.Unlock()
	dele.recvBlockTime = recvBT
	dele.sendBlockTime = sendBT
}

func (dele *DefaultTCDelegate) SetRecvBlockTime(recvBT time.Duration) {
	dele.Lock()
	defer dele.Unlock()
	dele.recvBlockTime = recvBT
}

func (dele *DefaultTCDelegate) SetSendBlockTime(sendBT time.Duration) {
	dele.Lock()
	defer dele.Unlock()
	dele.sendBlockTime = sendBT
}

func (dele *DefaultTCDelegate) SetIOBufLen(recvBL int, sendBL int) {
	dele.Lock()
	defer dele.Unlock()
	dele.recvBufLen = recvBL
	dele.sendBufLen = sendBL
}

func (dele *DefaultTCDelegate) SetRecvBufLen(recvBL int) {
	dele.Lock()
	defer dele.Unlock()
	dele.recvBufLen = recvBL
}

func (dele *DefaultTCDelegate) SetSendBufLen(sendBL int) {
	dele.Lock()
	defer dele.Unlock()
	dele.sendBufLen = sendBL
}

func (dele *DefaultTCDelegate) SetNoDelay(nodelay bool) {
	dele.Lock()
	defer dele.Unlock()
	dele.noDelay = nodelay
}

func (dele *DefaultTCDelegate) SetKeepAlive(keppalive bool) {
	dele.Lock()
	defer dele.Unlock()
	dele.keepAlive = keppalive
}

func (dele *DefaultTCDelegate) SetCientAliveTime(aliveT time.Duration) {
	dele.Lock()
	defer dele.Unlock()
	dele.aliveTime = aliveT
}

func (dele *DefaultTCDelegate) SetMaxPackLen(maxPL int) {
	dele.Lock()
	defer dele.Unlock()
	dele.maxPackLen = maxPL
}

func (dele *DefaultTCDelegate) SetDataInSupervisor(dataInSupervisor func(msg *NetMsg)) {
	dele.Lock()
	defer dele.Unlock()
	dele.dataInSupervisor = dataInSupervisor
}

func (dele *DefaultTCDelegate) SetDataOutSupervisor(dataOutSupervisor func(msg *NetMsg)) {
	dele.Lock()
	defer dele.Unlock()
	dele.dataOutSupervisor = dataOutSupervisor
}

func (dele *DefaultTCDelegate) SetShowClientData(show bool) {
	dele.Lock()
	defer dele.Unlock()
	dele.showClientData = show
}

func (dele *DefaultTCDelegate) ShowClientData() bool {
	return dele.showClientData
}

func (dele *DefaultTCDelegate) MaxPackLen() int {
	return dele.maxPackLen
}

func (dele *DefaultTCDelegate) RecvBufLen() int {
	return dele.recvBufLen
}

func (dele *DefaultTCDelegate) SendBufLen() int {
	return dele.sendBufLen
}

func (dele *DefaultTCDelegate) NoDelay() bool {
	return dele.noDelay
}

func (dele *DefaultTCDelegate) KeepAlive() bool {
	return dele.keepAlive
}

func (dele *DefaultTCDelegate) RecvBlockTime() time.Duration {
	return dele.recvBlockTime
}

func (dele *DefaultTCDelegate) SendBlockTime() time.Duration {
	return dele.sendBlockTime
}

func (dele *DefaultTCDelegate) AliveTime() time.Duration {
	return dele.aliveTime
}

func (dele *DefaultTCDelegate) OnNewConn(client *TcpClient) {
	//return dele.newConnCB
}

func (dele *DefaultTCDelegate) SetTag(tag string) {
	dele.tag = tag
}

/*func (dele *DefaultTCDelegate) SetNewConnCB(cb func(*TcpClient)) {
	dele.newConnCB = cb
}*/

func (dele *DefaultTCDelegate) OnServerStop() {

}
