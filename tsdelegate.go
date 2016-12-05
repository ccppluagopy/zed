package zed

import (
	"encoding/binary"
	"io"
	"sync"
	"time"
)

type ZTcpClientDelegate interface {
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
	SetIOBufLen(int, int)
	SetCientAliveTime(time.Duration)
	SetMaxPackLen(int)

	ShowClientData() bool
	MaxPackLen() int
	RecvBufLen() int
	SendBufLen() int
	RecvBlockTime() time.Duration
	SendBlockTime() time.Duration
	AliveTime() time.Duration
}

type DefaultTCDelegate struct {
	sync.Mutex
	Server     *TcpServer
	HandlerMap map[CmdType]MsgHandler

	showClientData    bool
	maxPackLen        int
	recvBlockTime     time.Duration
	recvBufLen        int
	sendBlockTime     time.Duration
	sendBufLen        int
	aliveTime         time.Duration
	delegate          ZTcpClientDelegate
	dataInSupervisor  func(*NetMsg)
	dataOutSupervisor func(*NetMsg)
}

func (dele *DefaultTCDelegate) RecvMsg(client *TcpClient) *NetMsg {
	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
		msg     *NetMsg
		server  = dele.Server
	)

	if err = (*client.conn).SetReadDeadline(time.Now().Add(dele.recvBlockTime)); err != nil {
		if server != nil && server.showClientData {
			ZLog("RecvMsg %s SetReadDeadline Err: %v.", client.Info(), err)
		}
		goto Exit
	}

	readLen, err = io.ReadFull(client.conn, head)
	if err != nil || readLen < PACK_HEAD_LEN {
		if server != nil && server.showClientData {
			ZLog("RecvMsg %s Read Head Err: %v %d.", client.Info(), err, readLen)
		}
		goto Exit
	}

	if err = (*client.conn).SetReadDeadline(time.Now().Add(dele.recvBlockTime)); err != nil {
		if server != nil && server.showClientData {
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
			if server != nil && server.showClientData {
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

	if msg.Len > 0 && (msg.Data == nil || msg.Len != len(msg.Data)) {
		ZLog("SendMsg Err: msg.Len(%d) != len(Data)%v", msg.Len, msg.Data)
		goto Exit
	}

	if msg.Len > dele.maxPackLen {
		ZLog("SendMsg Err: Body Len(%d) > MAXPACK_LEN(%d)", msg.Len, dele.maxPackLen)
		goto Exit
	}

	if err := (*client.conn).SetWriteDeadline(time.Now().Add(dele.sendBlockTime)); err != nil {
		ZLog("%s SetWriteDeadline Err: %v.", client.Info(), err)
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
		return true
	}

Exit:
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

func (dele *DefaultTCDelegate) HandleMsg(msg *NetMsg) {
	/*defer PanicHandle(true, func() {
		ZLog("HandleMsg %s panic err!", msg.Client.Info())
		msg.Client.Stop()
	})*/

	cb, ok := dele.HandlerMap[msg.Cmd]
	if ok {
		if cb(msg) {
			return
		} else {
			ZLog("HandleMsg Error, %s Msg Cmd: %d, Data: %v.", msg.Client.Info(), msg.Cmd, msg.Data)
			msg.Client.Stop()
		}
	} else {
		ZLog("No Handler For Cmd %d, %s", msg.Cmd, msg.Client.Info())
	}
}

func (dele *DefaultTCDelegate) AddMsgHandler(cmd CmdType, cb MsgHandler) {
	dele.Lock()
	defer dele.Unlock()

	ZLog("TcpServer DefaultTCDelegate AddMsgHandler, Cmd: %d", cmd)
	if dele.HandlerMap == nil {
		dele.HandlerMap = make(map[CmdType]MsgHandler)
	}
	dele.HandlerMap[cmd] = cb
}

func (dele *DefaultTCDelegate) RemoveMsgHandler(cmd CmdType) {
	dele.Lock()
	defer dele.Unlock()

	ZLog("TcpServer DefaultTCDelegate RemoveMsgHandler, Cmd: %d", cmd)

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

func (dele *DefaultTCDelegate) SetIOBufLen(recvBL int, sendBL int) {
	dele.Lock()
	defer dele.Unlock()
	dele.recvBufLen = recvBL
	dele.sendBufLen = sendBL
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

func (dele *DefaultTCDelegate) RecvBlockTime() time.Duration {
	return dele.recvBlockTime
}

func (dele *DefaultTCDelegate) SendBlockTime() time.Duration {
	return dele.sendBlockTime
}

func (dele *DefaultTCDelegate) AliveTime() time.Duration {
	return dele.aliveTime
}
