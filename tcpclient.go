package zed

import (
	"encoding/binary"
	//"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type TcpClient struct {
	sync.RWMutex
	conn    *net.TCPConn
	parent  *TcpServer
	Id      ClientIDType
	Idx     int
	Addr    string
	closeCB map[interface{}]ClientCloseCB
	running bool
}

func (client *TcpClient) AddCloseCB(key interface{}, cb ClientCloseCB) {
	client.Lock()
	defer client.Unlock()
	if client.running {
		client.closeCB[key] = cb
	}
}

func (client *TcpClient) RemoveCloseCB(key interface{}) {
	client.Lock()
	defer client.Unlock()
	if client.running {
		delete(client.closeCB, key)
	}
}

func (client *TcpClient) Stop() {
	NewCoroutine(func() {
		client.Lock()
		defer client.Unlock()

		if client.running {
			client.conn.Close()

			for _, cb := range client.closeCB {
				cb(client)
			}

			for key, _ := range client.closeCB {
				delete(client.closeCB, key)
			}

			client.running = false
			LogInfo(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) Stopped.", client.Id, client.Addr)
		}
	})
}

func (client *TcpClient) SendMsg(msg *NetMsg) {
	client.RLock()
	defer client.RUnlock()

	if client.running {
		if err := (*client.conn).SetWriteDeadline(time.Now().Add(WRITE_BLOCK_TIME)); err != nil {
			LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) SetWriteDeadline Err: %v.", client.Id, client.Addr, err)
			goto Exit
		}

		buf := make([]byte, PACK_HEAD_LEN+len(msg.Buf))
		binary.LittleEndian.PutUint32(buf, uint32(len(msg.Buf)))
		binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))
		copy(buf[PACK_HEAD_LEN:], msg.Buf)

		writeLen, err := client.conn.Write(buf)

		LogInfo(LOG_IDX, client.Idx, "Send Msg Client(Id: %s, Addr: %s) Cmd: %d, BufLen: %d, Buf: %s", client.Id, client.Addr, msg.Cmd, msg.BufLen, string(msg.Buf))

		if err != nil || writeLen != len(buf) {
			goto Exit
		}

		return
	}

Exit:
	client.Stop()
}

func (client *TcpClient) reader() {
	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
	)

	for {
		if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
			LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) SetReadDeadline Err: %v.", client.Id, client.Addr, err)
			goto Exit
		}

		readLen, err = io.ReadFull(client.conn, head)
		if err != nil || readLen < PACK_HEAD_LEN {
			LogInfo(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) Read Head Err: %v.", client.Id, client.Addr, err)
			goto Exit
		}

		if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
			LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) SetReadDeadline Err: %v.", client.Id, client.Addr, err)
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
				LogInfo(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) Read Body Err: %v.", client.Id, client.Addr, err)
				goto Exit
			}
		}

		LogInfo(LOG_IDX, client.Idx, "Recv Msg Client(Id: %s, Addr: %s) Cmd: %d, BufLen: %d, Data: %s", client.Id, client.Addr, msg.Cmd, msg.BufLen, string(msg.Buf))

		client.parent.HandleMsg(msg)
	}

Exit:
	client.Stop()
}

func (client *TcpClient) start() bool {
	if err := client.conn.SetKeepAlive(true); err != nil {
		LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) SetKeepAlive Err: %v.", client.Id, client.Addr, err)
		return false
	}

	if err := client.conn.SetKeepAlivePeriod(KEEP_ALIVE_TIME); err != nil {
		LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) SetKeepAlivePeriod Err: %v.", client.Id, client.Addr, err)
		return false
	}

	if err := (*client.conn).SetReadBuffer(RECV_BUF_LEN); err != nil {
		LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) SetReadBuffer Err: %v.", client.Id, client.Addr, err)
		return false
	}
	if err := (*client.conn).SetWriteBuffer(SEND_BUF_LEN); err != nil {
		LogError(LOG_IDX, client.Idx, "Client(Id: %s, Addr: %s) SetWriteBuffer Err: %v.", client.Id, client.Addr, err)
		return false
	}

	NewCoroutine(func() {
		client.reader()
	})

	LogInfo(LOG_IDX, client.Idx, "New Client Start Client(Id: %s, Addr: %s)", client.Id, client.Addr)

	return true
}

func newTcpClient(parent *TcpServer, conn *net.TCPConn) *TcpClient {
	client := &TcpClient{
		conn:    conn,
		parent:  parent,
		Id:      NullId,
		Idx:     parent.ClientNum,
		Addr:    conn.RemoteAddr().String(),
		closeCB: make(map[interface{}]ClientCloseCB),
		running: true,
	}

	return client
}
