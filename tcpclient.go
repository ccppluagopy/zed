package zed

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type ClientCloseCB func(client *TcpClient)

type TcpClient struct {
	conn    *net.TCPConn
	parent  *TcpServer
	recvQ   chan *NetMsg
	sendQ   chan *NetMsg
	Id      ClientIDType
	Idx     int
	Addr    string
	closeCB map[interface{}]ClientCloseCB
	running bool
}

func (client *TcpClient) AddCloseCB(key interface{}, cb ClientCloseCB) {
	client.closeCB[key] = cb
}

func (client *TcpClient) RemoveCloseCB(key interface{}) {
	delete(client.closeCB, key)
}

func (client *TcpClient) HandleMsg(msg *NetMsg) bool {
	cb, ok := client.parent.handlerMap[msg.Cmd]
	if ok {
		return cb(client, msg)
	} else {
		LogError(LOG_IDX, client.Idx, "No Handler For Cmd %d From Client(Id: %s, Addr: %s.", client.Id, client.Addr)
		goto Err
	}

Err:
	//client.Stop()
	return false
}

func (client *TcpClient) Stop() {
	if client.running {
		client.running = false

		client.conn.Close()

		close(client.recvQ)
		close(client.sendQ)

		for _, cb := range client.closeCB {
			cb(client)
		}

		for key, _ := range client.closeCB {
			delete(client.closeCB, key)
		}

		LogInfo(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) Stopped.", client.Id, client.Addr)
	}
}

func (client *TcpClient) SendMsg(msg *NetMsg) {

}

func (client *TcpClient) startReader(enableMsgHandleCor bool) {
	defer PanicHandle(false, fmt.Sprintf("Client(Id: %s, Addr: %s) Msg Reader exit.", client.Id, client.Addr))

	var (
		head     = make([]byte, PACK_HEAD_LEN)
		readLen  = 0
		err      error
		lastRead time.Time
	)

	for {
		if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
			goto Exit
		}

		lastRead = time.Now().UTC()
		readLen, err = io.ReadFull(client.conn, head)
		if err != nil || readLen < PACK_HEAD_LEN {
			LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) Read Head Error: %v!", err)
			goto Exit
		}

		if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
			LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) SetReadDeadline Error: %v!", client.Id, client.Addr, err)
			goto Exit
		}

		var msg = &NetMsg{
			BufLen: binary.LittleEndian.Uint32(head[0:4]),
			Cmd:    CmdType(binary.LittleEndian.Uint32(head[4:8])),
		}

		if msg.BufLen > 0 {
			msg.Buf = make([]byte, msg.BufLen)
			readLen, err := io.ReadFull(client.conn, msg.Buf)
			if err != nil || readLen != int(msg.BufLen) {
				LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) Read body error: %v", client.Id, client.Addr, err)
				goto Exit
			}
		}

		if enableMsgHandleCor {
			client.recvQ <- msg
		} else {
			client.HandleMsg(msg)
		}

	}

Exit:
	client.Stop()
	return
}

func (client *TcpClient) startWriter() {
	defer PanicHandle(false, fmt.Sprintf("Client(Id: %s, Addr: %s) Msg Writer exit.", client.Id, client.Addr))

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
			LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) SetWriteDeadline Error: %v!", client.Id, client.Addr, err)
			goto Exit
		}

		buf = make([]byte, PACK_HEAD_LEN+len(msg.Buf))
		binary.LittleEndian.PutUint32(buf, uint32(len(msg.Buf)))
		binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))
		copy(buf[PACK_HEAD_LEN:], msg.Buf)

		writeLen, err := client.conn.Write(buf)
		if err != nil || writeLen != len(buf) {
			goto Exit
		}
	}

Exit:
	client.Stop()
}

func (client *TcpClient) startMsgHandler() {
	defer PanicHandle(false, fmt.Sprintf("Client(Id: %s, Addr: %s) Msg Handler exit.", client.Id, client.Addr))

	var msg *NetMsg

	client.recvQ = make(chan *NetMsg, 5)

	for {
		msg = <-client.recvQ

		if msg == nil {
			LogInfo(LOG_IDX, client.Idx, "%d msgCoroutine exit.", client.Idx)
			return
		}

		client.HandleMsg(msg)
	}
}

func (client *TcpClient) start() bool {
	if err := client.conn.SetKeepAlive(true); err != nil {
		LogError(LOG_IDX, client.Idx, "%d SetKeepAlive error: %v", client.Idx, err)
		return false
	}

	if err := client.conn.SetKeepAlivePeriod(KEEP_ALIVE_TIME); err != nil {
		LogError(LOG_IDX, client.Idx, "%d SetKeepAlivePeriod error: %v", client.Idx, err)
		return false
	}

	if err := (*client.conn).SetReadBuffer(RECV_BUF_LEN); err != nil {
		LogError(LOG_IDX, client.Idx, "%d SetReadBuffer error: %v", client.Idx, err)
		return false
	}
	if err := (*client.conn).SetWriteBuffer(SEND_BUF_LEN); err != nil {
		LogError(LOG_IDX, client.Idx, "%d SetWriteBuffer error: %v", client.Idx, err)
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

	LogInfo(LOG_IDX, client.Idx, "New Client Start, Idx: %d.", client.Idx)

	return true
}

func newTcpClient(parent *TcpServer, conn *net.TCPConn) *TcpClient {
	parent.ClientNum = parent.ClientNum + 1

	client := &TcpClient{
		conn:    conn,
		parent:  parent,
		recvQ:   nil,
		sendQ:   nil,
		Id:      NullId,
		Idx:     parent.ClientNum,
		Addr:    conn.RemoteAddr().String(),
		closeCB: make(map[interface{}]ClientCloseCB),
		running: true,
	}

	return client
}
