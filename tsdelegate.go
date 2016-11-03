package zed

import (
	"encoding/binary"
	"io"
	"sync"
	"time"
)

type ZServerDelegate interface {
	RecvMsg(*TcpClient) *NetMsg
	SendMsg(*NetMsg) bool
	HandleMsg(*NetMsg)
	SetServer(*TcpServer)
	AddMsgHandler(cmd CmdType, cb MsgHandler)
	RemoveMsgHandler(cmd CmdType)
	/*SetShowClientData(bool)
	SetDataInSupervisor(func(msg *NetMsg))
	SetDataOutSupervisor(func(msg *NetMsg))
	SetIOBlockTime(time.Duration, time.Duration)
	SetIOBufLen(int, int)
	SetCientAliveTime(time.Duration)
	SetMaxPackLen(int)*/
}

type DefaultTSDelegate struct {
	sync.Mutex
	Server     *TcpServer
	HandlerMap map[CmdType]MsgHandler
}

func (dele *DefaultTSDelegate) RecvMsg(client *TcpClient) *NetMsg {
	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
		msg     *NetMsg
		server  = dele.Server
	)

	if err = (*client.conn).SetReadDeadline(time.Now().Add(client.parent.recvBlockTime)); err != nil {
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

	if err = (*client.conn).SetReadDeadline(time.Now().Add(client.parent.recvBlockTime)); err != nil {
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
	if msg.Len > client.parent.maxPackLen {
		ZLog("RecvMsg Read Body Err: Body Len(%d) > MAXPACK_LEN(%d)", msg.Len, client.parent.maxPackLen)
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

func (dele *DefaultTSDelegate) SendMsg(msg *NetMsg) bool {
	var (
		writeLen = 0
		buf      []byte
		err      error
		client   = msg.Client
	)

	if msg.Len > 0 && (msg.Data == nil || msg.Len != len(msg.Data)) {
		ZLog("SendMsg Err: msg.Len(%d) != len(Data)%v", msg.Len, msg.Data)
		goto Exit
	}

	if msg.Len > client.parent.maxPackLen {
		ZLog("SendMsg Err: Body Len(%d) > MAXPACK_LEN(%d)", msg.Len, client.parent.maxPackLen)
		goto Exit
	}

	if err := (*client.conn).SetWriteDeadline(time.Now().Add(client.parent.sendBlockTime)); err != nil {
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

func (dele *DefaultTSDelegate) MsgFilter(msg *NetMsg) bool {
	Println("DefaultTSDelegate MsgFilter 00000000000000000000000")
	server := dele.Server
	if server != nil && server.msgFilter != nil {
		return server.msgFilter(msg)
	}
	return true
}

func (dele *DefaultTSDelegate) HandleMsg(msg *NetMsg) {
	//Println("DefaultTSDelegate HandleMsg 2222")
	defer PanicHandle(true, func() {
		ZLog("HandleMsg %s panic err!", msg.Client.Info())
		msg.Client.Stop()
	})

	cb, ok := dele.HandlerMap[msg.Cmd]
	if ok && dele.MsgFilter(msg) {

		if cb(msg) {
			return
		} else {
			ZLog("HandleMsg Error, %s Msg Cmd: %d, Data: %v.", msg.Client.Info(), msg.Cmd, msg.Data)
		}
	} else {
		ZLog("No Handler For Cmd %d, %s", msg.Cmd, msg.Client.Info())
	}

	msg.Client.Stop()
}

func (dele *DefaultTSDelegate) AddMsgHandler(cmd CmdType, cb MsgHandler) {
	dele.Lock()
	defer dele.Unlock()

	ZLog("TcpServer DefaultTSDelegate AddMsgHandler, Cmd: %d", cmd)
	if dele.HandlerMap == nil {
		dele.HandlerMap = make(map[CmdType]MsgHandler)
	}
	dele.HandlerMap[cmd] = cb
}

func (dele *DefaultTSDelegate) RemoveMsgHandler(cmd CmdType) {
	dele.Lock()
	defer dele.Unlock()

	ZLog("TcpServer DefaultTSDelegate RemoveMsgHandler, Cmd: %d", cmd)

	delete(dele.HandlerMap, cmd)
}

func (dele *DefaultTSDelegate) SetServer(server *TcpServer) {
	dele.Server = server
}
