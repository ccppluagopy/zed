package zed

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	//"sync"
	"sync/atomic"
	"time"
)

type TcpClient struct {
	//sync.RWMutex
	conn    *net.TCPConn
	parent  *TcpServer
	recvQ   chan *NetMsg
	sendQ   chan *NetMsg
	Id      ClientIDType
	Idx     int
	Addr    string
	closeCB map[interface{}]ClientCloseCB
	running uint32
}

func (client *TcpClient) AddCloseCB(key interface{}, cb ClientCloseCB) {
	client.closeCB[key] = cb
}

func (client *TcpClient) RemoveCloseCB(key interface{}) {
	delete(client.closeCB, key)
}

func (client *TcpClient) Stop() {
	atomic.AddUint32(&client.running, 1)
	if client.running == 1 {
		/*client.Lock()
		client.running = false
		client.Unlock()*/

		client.conn.Close()

		if client.recvQ != nil {
			close(client.recvQ)
		}

		if client.sendQ != nil {
			close(client.sendQ)
		}

		for _, cb := range client.closeCB {
			cb(client)
		}

		for key, _ := range client.closeCB {
			delete(client.closeCB, key)
		}

		ZLog("Client(Id: %s, Addr: %s) Stopped.", client.Id, client.Addr)
	}
}

func (client *TcpClient) SendMsg(msg *NetMsg) {
	if client.parent.msgSendCorNum == 0 {
		/*client.RLock()
		defer client.RUnlock()*/
		vrun := atomic.LoadUint32(&client.running)
		if vrun == 0 {
			client.sendQ <- msg
		}
	} else {
		client.parent.SendMsg(msg)
	}
}

func (client *TcpClient) startReader(enableMsgHandleCor bool) {
	defer PanicHandle(true, fmt.Sprintf("Client(Id: %s, Addr: %s) Msg Reader exit.", client.Id, client.Addr))

	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
	)

	for {
		if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
			goto Exit
		}

		readLen, err = io.ReadFull(client.conn, head)
		if err != nil || readLen < PACK_HEAD_LEN {
			ZLog("Client(Id: %s, Addr: %s) Read Head Error: %v!", client.Id, client.Addr, err)
			goto Exit
		}

		if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
			ZLog("Client(Id: %s, Addr: %s) SetReadDeadline Error: %v!", client.Id, client.Addr, err)
			goto Exit
		}

		var msg = &NetMsg{
			Cmd:    CmdType(binary.LittleEndian.Uint32(head[4:8])),
			BufLen: binary.LittleEndian.Uint32(head[0:4]),
			Client: client,
		}

		if msg.BufLen > 0 {
			msg.Buf = make([]byte, msg.BufLen)
			readLen, err := io.ReadFull(client.conn, msg.Buf)
			if err != nil || readLen != int(msg.BufLen) {
				ZLog("Client(Id: %s, Addr: %s) Read body error: %v", client.Id, client.Addr, err)
				goto Exit
			}
		}

		LogInfo(LOG_IDX, LOG_IDX, "Recv Msg, Cmd: %d, BufLen: %d, Data: %s", msg.Cmd, msg.BufLen, string(msg.Buf))

		if enableMsgHandleCor {
			client.recvQ <- msg
		} else {
			client.parent.RelayMsg(msg)
		}

	}

Exit:
	client.Stop()
	return
}

func (client *TcpClient) startWriter() {
	defer PanicHandle(true, fmt.Sprintf("Client(Id: %s, Addr: %s) Msg Writer exit.", client.Id, client.Addr))

	ZLog(fmt.Sprintf("startWriter Client(Id: %s, Addr: %s) Msg Handler exit.", client.Id, client.Addr))

	var (
		msg *NetMsg
		err error
		buf []byte
	)

	client.sendQ = make(chan *NetMsg, 5)

	for {
		msg = <-client.sendQ

		if msg == nil {
			goto Exit
		}

		if err = (*client.conn).SetWriteDeadline(time.Now().Add(WRITE_BLOCK_TIME)); err != nil {
			ZLog("Write Failed Cmd: %d, Len: %d, Buf: %s", msg.Cmd, msg.BufLen, string(msg.Buf))
			ZLog("Client(Id: %s, Addr: %s) SetWriteDeadline Error: %v!", client.Id, client.Addr, err)
			goto Exit
		}

		buf = make([]byte, PACK_HEAD_LEN+len(msg.Buf))
		binary.LittleEndian.PutUint32(buf, uint32(len(msg.Buf)))
		binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))
		copy(buf[PACK_HEAD_LEN:], msg.Buf)

		writeLen, err := client.conn.Write(buf)
		//ZLog("Write Success Cmd: %d, Len: %d, Buf: %s", msg.Cmd, msg.BufLen, string(msg.Buf))

		LogInfo(LOG_IDX, LOG_IDX, "Send Success Cmd: %d, BufLen: %d, Buf: %s", msg.Cmd, msg.BufLen, string(msg.Buf))

		if err != nil || writeLen != len(buf) {
			goto Exit
		}
	}

Exit:
	client.Stop()
}

func (client *TcpClient) startMsgHandler() {
	defer PanicHandle(true, fmt.Sprintf("Client(Id: %s, Addr: %s) Msg Handler exit.", client.Id, client.Addr))

	ZLog("startMsgHandler Client(Id: %s, Addr: %s) Msg Handler exit.", client.Id, client.Addr)

	var msg *NetMsg

	client.recvQ = make(chan *NetMsg, 5)

	for {
		msg = <-client.recvQ

		if msg == nil {
			return
		}

		client.parent.HandleMsg(msg)
	}

	client.Stop()
}

func (client *TcpClient) start() bool {
	if err := client.conn.SetKeepAlive(true); err != nil {
		ZLog("%d SetKeepAlive error: %v", client.Idx, err)
		return false
	}

	if err := client.conn.SetKeepAlivePeriod(KEEP_ALIVE_TIME); err != nil {
		ZLog("%d SetKeepAlivePeriod error: %v", client.Idx, err)
		return false
	}

	if err := (*client.conn).SetReadBuffer(RECV_BUF_LEN); err != nil {
		ZLog("%d SetReadBuffer error: %v", client.Idx, err)
		return false
	}
	if err := (*client.conn).SetWriteBuffer(SEND_BUF_LEN); err != nil {
		ZLog("%d SetWriteBuffer error: %v", client.Idx, err)
		return false
	}

	if client.parent.msgSendCorNum == 0 {
		go client.startWriter()
	}

	if client.parent.msgHandleCorNum == 0 {
		go client.startMsgHandler()
	}

	if client.parent.msgHandleCorNum == 0 {
		go client.startReader(true)
	} else {
		go client.startReader(false)
	}

	ZLog("New Client Start, Idx: %d.", client.Idx)

	return true
}

func newTcpClient(parent *TcpServer, conn *net.TCPConn) *TcpClient {
	//parent.ClientNum = parent.ClientNum + 1

	client := &TcpClient{
		conn:    conn,
		parent:  parent,
		recvQ:   nil,
		sendQ:   nil,
		Id:      NullId,
		Idx:     parent.ClientNum,
		Addr:    conn.RemoteAddr().String(),
		closeCB: make(map[interface{}]ClientCloseCB),
		running: 0,
	}

	return client
}
