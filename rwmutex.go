package zed

import (
	"encoding/binary"
	//"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	rwmutexs       = make(map[string]*RWMutex)
	rwmutexconnops = make(map[*TcpClient]map[string]bool)
)

const (
	RWMUTEX_STATE_FREE = iota
	RWMUTEX_STATE_READING
	RWMUTEX_STATE_WRITING
)
const (
	RWMUTEX_CMD_RLOCK = iota
	RWMUTEX_CMD_RUNLOCK
	RWMUTEX_CMD_RLOCK_ERR
	RWMUTEX_CMD_RUNLOCK_ERR
	RWMUTEX_CMD_LOCK
	RWMUTEX_CMD_UNLOCK
	RWMUTEX_CMD_LOCK_ERR
	RWMUTEX_CMD_UNLOCK_ERR
)

type RWMutexTeam struct {
	//events map[string]uint64
	count uint64
}

type RWMutex struct {
	sync.RWMutex
	state      int
	server     *TcpServer
	mtxmap     map[string]map[*TcpClient]*TcpClient
	mtxcurrmap map[string]*TcpClient
}

func (rwmtx *RWMutex) PublicR(key string) {
	mtxmap, ok := rwmtx.mtxmap[key]
	if ok {
		for client, _ := range mtxmap {
			rwmtx.mtxcurrmap[key] = client
			Printf("[PublicR] %s\n", client.conn.RemoteAddr())
			client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_LOCK, Len: 0, Data: nil})
			delete(mtxmap, client)
			return
		}

		rwmtx.mtxcurrmap[key] = nil
	}
}

func NewRWMutexServer(name string, addr string) *RWMutex {
	if rwmtx, ok := rwmutexs[name]; !ok {
		rwmtx = &RWMutex{
			state:      RWMUTEX_STATE_FREE,
			server:     NewTcpServer(name),
			mtxmap:     make(map[string]map[*TcpClient]*TcpClient),
			mtxcurrmap: make(map[string]*TcpClient),
		}
		/*nRLock := 0
		nRUnLock := 0*/

		handleLock := func(msg *NetMsg) bool {
			/*			nRLock = nRLock + 1
						ZLog("handleRLock %d", nRLock)
			*/
			rwmtx.Lock()
			defer rwmtx.Unlock()

			key := string(msg.Data)
			mtxmap, ok := rwmtx.mtxmap[key]
			if key == "" {
				goto Err
			}

			if !ok {
				mtxmap = make(map[*TcpClient]*TcpClient)
				rwmtx.mtxmap[key] = mtxmap
			}

			if _, ok2 := mtxmap[msg.Client]; ok2 {
				goto Err
			}

			if len(mtxmap) == 1 {
				rwmtx.mtxcurrmap[key] = msg.Client
				Printf("[Lock] %s\n", msg.Client.conn.RemoteAddr())
				msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_LOCK, Len: 0, Data: nil})
			} else {
				mtxmap[msg.Client] = msg.Client
			}

			return true
		Err:
			msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RLOCK_ERR, Len: 0, Data: nil})
			return false
		}

		handleUnLock := func(msg *NetMsg) bool {
			key := string(msg.Data)
			if _, ok := rwmtx.mtxcurrmap[key]; ok {
				rwmtx.PublicR(key)
				return true
			}
			return false
		}

		handleRLock := func(msg *NetMsg) bool {
			return true
		}

		handleRUnLock := func(msg *NetMsg) bool {
			return true
		}

		handleConnClose := func(client *TcpClient) {

		}
		rwmtx.server.SetConnCloseCB(handleConnClose)
		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_LOCK, handleLock)
		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_UNLOCK, handleUnLock)
		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_RLOCK, handleRLock)
		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_RUNLOCK, handleRUnLock)

		NewCoroutine(func() {
			rwmtx.server.Start(addr)
		})
		return rwmtx
	} else {
		ZLog("NewRWMutex Error: %s has been exist.", name)
	}
	return nil
}

func DeleRWMutex(name string) {
	if rwmtx, ok := rwmutexs[name]; ok {
		rwmtx.server.Stop()
		delete(rwmutexs, name)
	}
}

type RWMutexClient struct {
	mutex sync.RWMutex
	addr  string
	conn  *net.TCPConn
}

func (client *RWMutexClient) SendMsg(msg *NetMsg) bool {
	var (
		writeLen = 0
		buf      []byte
		err      error
	)

	if client.conn == nil {
		tcpaddr, err2 := net.ResolveTCPAddr("tcp", client.addr)
		if err2 != nil {
			return false
		}

		client.conn, err = net.DialTCP("tcp", nil, tcpaddr)
		if err != nil {
			return false
		}
	}

	if err := (*client.conn).SetWriteDeadline(time.Now().Add(WRITE_BLOCK_TIME)); err != nil {
		ZLog("RWMutexClient SetWriteDeadline Err: %v.", err)
		goto Exit
	}

	buf = make([]byte, PACK_HEAD_LEN+msg.Len)
	binary.LittleEndian.PutUint32(buf, uint32(msg.Len))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))

	writeLen, err = client.conn.Write(buf)

	if err == nil && writeLen == len(buf) {
		return true
	}

Exit:
	return false
}

func (client *RWMutexClient) ReadMsg() *NetMsg {
	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
		msg     *NetMsg
	)

	if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
		ZLog("RWMutexClient SetReadDeadline Err: %v.", err)
		goto Exit
	}

	readLen, err = io.ReadFull(client.conn, head)
	if err != nil || readLen < PACK_HEAD_LEN {
		ZLog("RWMutexClient Read Head Err: %v %d.", err, readLen)
		goto Exit
	}

	if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
		ZLog("RWMutexClient SetReadDeadline Err: %v.", err)
		goto Exit
	}

	msg = &NetMsg{
		Cmd:  CmdType(binary.LittleEndian.Uint32(head[4:8])),
		Len:  0,
		Data: nil,
	}

	return msg

Exit:
	return nil
}

func (client *RWMutexClient) Lock(key string) bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	/*
	   const (
	   	RWMUTEX_CMD_RLOCK = iota
	   	RWMUTEX_CMD_RUNLOCK
	   	RWMUTEX_CMD_RLOCK_ERR
	   	RWMUTEX_CMD_RUNLOCK_ERR
	   	RWMUTEX_CMD_LOCK
	   	RWMUTEX_CMD_UNLOCK
	   	RWMUTEX_CMD_LOCK_ERR
	   	RWMUTEX_CMD_UNLOCK_ERR
	   )
	*/
	if client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_LOCK, Len: len(key), Data: []byte(key)}) {
		if msg := client.ReadMsg(); msg != nil && msg.Cmd == RWMUTEX_CMD_LOCK {
			return true
		}
	}

	return false
}

func (client *RWMutexClient) UnLock(key string) bool {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_UNLOCK, Len: len(key), Data: []byte(key)}) {
		if msg := client.ReadMsg(); msg != nil && msg.Cmd == RWMUTEX_CMD_UNLOCK {
			return true
		}
	}

	return false
}

func NewRWMutexClient(addr string) *RWMutexClient {
	return &RWMutexClient{
		addr: addr,
		conn: nil,
	}
}

func DeleRWMutexClient(client *RWMutexClient) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.conn.Close()
}
