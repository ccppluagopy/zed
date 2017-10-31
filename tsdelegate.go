package zed

import (
	"encoding/binary"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type ITcpClientDelegate interface {
	Init()

	RecvMsg(*TcpClient) INetMsg
	SendMsg(*TcpClient, INetMsg) bool
	SendData(*TcpClient, []byte) bool
	HandleMsg(INetMsg)

	SetServer(*TcpServer)
	AddMsgHandler(cmd uint32, cb MsgHandler)
	RemoveMsgHandler(cmd uint32)
	SetShowClientData(bool)
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
	inited     int32
	Server     *TcpServer
	HandlerMap map[uint32]MsgHandler

	showClientData    bool
	maxPackLen        int
	recvBlockTime     time.Duration
	recvBufLen        int
	sendBlockTime     time.Duration
	sendBufLen        int
	aliveTime         time.Duration
	noDelay           bool
	keepAlive         bool
	delegate          ITcpClientDelegate
	dataInSupervisor  func(*NetMsg)
	dataOutSupervisor func(*NetMsg)
	newConnCB         func(*TcpClient)

	tag string
}

func (dele *DefaultTCDelegate) RecvMsg(client *TcpClient) INetMsg {
	var (
		readLen = 0
		err     error
		msg     = &NetMsg{
			Client:    client,
			buf:       make([]byte, PACK_HEAD_LEN),
			encrypted: 1,
		}
		dataLen = 0
		//server  = dele.Server
	)

	if err = (*client.conn).SetReadDeadline(time.Now().Add(dele.recvBlockTime)); err != nil {
		if dele.showClientData {
			ZLog("RecvMsg %s SetReadDeadline Err: %v.", client.Info(), err)
		}
		goto Exit
	}

	readLen, err = io.ReadFull(client.conn, msg.buf)
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

	dataLen = int(binary.LittleEndian.Uint32(msg.buf[0:4]))

	if dataLen > 0 {
		if dataLen+PACK_HEAD_LEN > dele.maxPackLen {
			ZLog("RecvMsg Read Body Err: Msg Len(%d) > MAXPACK_LEN(%d)", dataLen+PACK_HEAD_LEN, dele.maxPackLen)
			goto Exit
		}
		msg.buf = append(msg.buf, make([]byte, dataLen)...)
		readLen, err := io.ReadFull(client.conn, msg.buf[PACK_HEAD_LEN:])
		if err != nil || readLen != dataLen {
			if dele.showClientData {
				ZLog("RecvMsg %s Read Body Err: %v.", client.Info(), err)
			}
			goto Exit
		}
	}

	if !msg.Decrypt() {
		goto Exit
	}

	return msg

Exit:
	return nil
}

func (dele *DefaultTCDelegate) SendMsg(client *TcpClient, msg INetMsg) bool {
	var (
		writeLen = 0
		buf      []byte
		err      error
	)

	if !msg.Encrypt() {
		return false
	}

	msgLen := msg.GetMsgLen()

	if msgLen > dele.maxPackLen {
		if dele.showClientData {
			ZLog("SendMsg Error: Body Len(%d) > MAXPACK_LEN(%d)", msgLen, dele.maxPackLen)
		}
		goto Exit
	}

	if msgLen < PACK_HEAD_LEN {
		if dele.showClientData {
			ZLog("SendMsg Error: Body Len(%d) > MAXPACK_LEN(%d)", msgLen, dele.maxPackLen)
		}
		goto Exit
	}

	if err := (*client.conn).SetWriteDeadline(time.Now().Add(dele.sendBlockTime)); err != nil {
		if dele.showClientData {
			ZLog("%s SetWriteDeadline Error: %v.", client.Info(), err)
		}
		goto Exit
	}

	buf = msg.GetSendBuf()
	binary.LittleEndian.PutUint32(buf, uint32(msg.GetDataLen()))
	writeLen, err = client.conn.Write(buf)

	if err == nil && writeLen == msgLen {
		if dele.showClientData {
			ZLog("[Send] %s Cmd: %d Len: %d", client.Info(), msg.GetCmd(), msgLen)
		}
		return true
	}

Exit:
	if dele.showClientData {
		ZLog("[Send] %s Cmd: %d Len: %d, Error: %s", client.Info(), msg.GetCmd(), msgLen, err.Error())
	}
	client.Stop()
	return false
}

func (dele *DefaultTCDelegate) SendData(client *TcpClient, data []byte) bool {
	msgLen := len(data)
	if msgLen > 0 {
		if len(data) > dele.maxPackLen {
			if dele.showClientData {
				ZLog("SendMsg Error: Body Len(%d) > MAXPACK_LEN(%d)", msgLen, dele.maxPackLen)
			}
			goto Exit
		}

		if err := (*client.conn).SetWriteDeadline(time.Now().Add(dele.sendBlockTime)); err != nil {
			if dele.showClientData {
				ZLog("%s SetWriteDeadline Error: %v.", client.Info(), err)
			}
			goto Exit
		}

		writeLen, err := client.conn.Write(data)
		if err == nil && writeLen == msgLen {
			if dele.showClientData {
				ZLog("[SendData] %s Len: %d", client.Info(), msgLen)
			}
			return true
		}
	} else {
		if dele.showClientData {
			ZLog("[SendData] %s Warn: Len == 0", client.Info())
		}
		return true
	}

Exit:
	if dele.showClientData {
		ZLog("[SendData] %s Error: Len == %d", client.Info(), msgLen)
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
	if atomic.CompareAndSwapInt32(&(dele.inited), 0, 1) {
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
}

func (dele *DefaultTCDelegate) HandleMsg(msg INetMsg) {
	/*defer HandlePanic(true, func() {
		ZLog("HandleMsg %s panic err!", msg.Client.Info())
		msg.Client.Stop()
	})*/
	client := msg.GetClient()
	cb, ok := dele.HandlerMap[msg.GetCmd()]
	if ok {
		if cb(msg) {
			return
		} else {
			if dele.showClientData {
				ZLog("%s HandleMsg Error, %s Msg Cmd: %d, Data: %v.", dele.tag, client.Info(), msg.GetCmd(), string(msg.GetData()))
			}
			client.Stop()
		}
	} else {
		if dele.showClientData {
			ZLog("%s HandleMsg Error: No Handler For Cmd %d, %s", dele.tag, msg.GetCmd(), client.Info())
		}
	}
}

func (dele *DefaultTCDelegate) AddMsgHandler(cmd uint32, cb MsgHandler) {
	dele.Lock()
	defer dele.Unlock()

	//ZLog("%s AddMsgHandler, Cmd: %d", dele.tag, cmd)
	if dele.HandlerMap == nil {
		dele.HandlerMap = make(map[uint32]MsgHandler)
	}
	dele.HandlerMap[cmd] = cb
}

func (dele *DefaultTCDelegate) RemoveMsgHandler(cmd uint32) {
	dele.Lock()
	defer dele.Unlock()

	//ZLog("%s RemoveMsgHandler, Cmd: %d", dele.tag, cmd)

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
