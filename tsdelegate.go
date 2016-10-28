package zed

/*
import (
	"sync"
)

type ZServerDelegate interface {
	RecvMsg(*TcpClient) *NetMsg
	SendMsg(*TcpClient) *NetMsg
	HandleMsg(*NetMsg)
}

type DefaultTSDelegate struct {
	sync.Mutex
	handlerMap map[CmdType]MsgHandler
}

func (dele *DefaultTSDelegate) RecvMsg(client *TcpClient) *NetMsg {
	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
		msg     *NetMsg
	)

	if err = (*client.conn).SetReadDeadline(time.Now().Add(client.parent.recvBlockTime)); err != nil {
		if showClientData {
			ZLog("%s SetReadDeadline Err: %v.", client.Info(), err)
		}
		goto Exit
	}

	readLen, err = io.ReadFull(client.conn, head)
	if err != nil || readLen < PACK_HEAD_LEN {
		if showClientData {
			ZLog("%s Read Head Err: %v %d.", client.Info(), err, readLen)
		}
		goto Exit
	}

	if err = (*client.conn).SetReadDeadline(time.Now().Add(client.parent.recvBlockTime)); err != nil {
		if showClientData {
			ZLog("%s SetReadDeadline Err: %v.", client.Info(), err)
		}
		goto Exit
	}

	msg = &NetMsg{
		Cmd:    CmdType(binary.LittleEndian.Uint32(head[4:8])),
		Len:    int(binary.LittleEndian.Uint32(head[0:4])),
		Client: client,
	}
	if msg.Len > client.parent.maxPackLen {
		ZLog("Read Body Err: Body Len(%d) > MAXPACK_LEN(%d)", msg.Len, client.parent.maxPackLen)
		goto Exit
	}
	if msg.Len > 0 {
		msg.Data = make([]byte, msg.Len)
		readLen, err := io.ReadFull(client.conn, msg.Data)
		if err != nil || readLen != int(msg.Len) {
			if showClientData {
				ZLog("%s Read Body Err: %v.", client.Info(), err)
			}
			goto Exit
		}
	}

	return msg

Exit:
	return nil
}

func (dele *DefaultTSDelegate) SendMsg(client *TcpClient) {
	client.Lock()
	defer client.Unlock()

	var (
		writeLen = 0
		buf      []byte
		err      error
	)

	if msg.Len > client.parent.maxPackLen {
		ZLog("SendMsg Err: Body Len(%d) > MAXPACK_LEN(%d)", msg.Len, client.parent.maxPackLen)
		goto Exit
	}

	if err := (*client.conn).SetWriteDeadline(time.Now().Add(client.parent.sendBlockTime)); err != nil {
		ZLog("%s SetWriteDeadline Err: %v.", client.Info(), err)
		goto Exit
	}

	msg.Client = client
	buf = make([]byte, PACK_HEAD_LEN+msg.Len)
	binary.LittleEndian.PutUint32(buf, uint32(msg.Len))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))
	if msg.Len > 0 {
		copy(buf[PACK_HEAD_LEN:], msg.Data)
	}

	writeLen, err = client.conn.Write(buf)

	if dataOutSupervisor != nil {
		dataOutSupervisor(msg)
	} else if showClientData {
		ZLog("[Send] %s Cmd: %d, Len: %d, Data: %s", client.Info(), msg.Cmd, msg.Len, string(msg.Data))
	}

	if err == nil && writeLen == len(buf) {
		return
	}

Exit:
	client.Stop()
}

func (dele *DefaultTSDelegate) MsgFilter(*NetMsg) bool {
	return true
}

func (dele *DefaultTSDelegate) HandleMsg(*NetMsg) {
	defer PanicHandle(true, func() {
		ZLog("HandleMsg %s panic err!", msg.Client.Info())
		msg.Client.Stop()
	})

	cb, ok := dele.handlerMap[msg.Cmd]
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

	dele.handlerMap[cmd] = cb
}

func (dele *DefaultTSDelegate) RemoveMsgHandler(cmd CmdType, cb MsgHandler) {
	dele.Lock()
	defer dele.Unlock()

	ZLog("TcpServer DefaultTSDelegate RemoveMsgHandler, Cmd: %d", cmd)

	delete(dele.handlerMap, cmd)
}
*/
