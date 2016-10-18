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
	rwmutexs       = make(map[string]*Mutex)
	rwmutexconnops = make(map[*TcpClient]map[string]bool)
)

const (
	MUTEX_STATE_FREE = iota
	MUTEX_STATE_READING
	MUTEX_STATE_WRITING
)
const (
	MUTEX_CMD_LOCK = iota
	MUTEX_CMD_UNLOCK
	MUTEX_CMD_LOCK_ERR
	MUTEX_CMD_UNLOCK_ERR

	MUTEX_NET_ERR
	MUTEX_LOCK_EMPTY_KEY_ERR
	MUTEX_TWICE_LOCK_ERR
	MUTEX_INVALID_UNLOCK_ERR
	MUTEX_UNLOCK_EMPTY_KEY_ERR
)

type Mutex struct {
	sync.Mutex
	state      int
	server     *TcpServer
	mtxmap     map[string]map[*TcpClient]*TcpClient
	mtxcurrmap map[string]*TcpClient
}

func Printff(fmt string, v ...interface{}) {

}

type ZMutexErr struct {
	errno int
}

func (err *ZMutexErr) Error() string {
	switch err.errno {
	case MUTEX_NET_ERR:
		return "Error: ZMutex Operation Net Unavailable."
	case MUTEX_LOCK_EMPTY_KEY_ERR:
		return "Error: ZMutex Lock key is empty."
	case MUTEX_TWICE_LOCK_ERR:
		return "Error: ZMutex Twice Lock."
	case MUTEX_INVALID_UNLOCK_ERR:
		return "Error: ZMutex Invalid UnLock Operation."
	case MUTEX_UNLOCK_EMPTY_KEY_ERR:
		return "Error: ZMutex UnLock key is empty."
	}

	return "ZMutexError"
}

func (rwmtx *Mutex) PublicR(key string) {
	mtxmap, ok := rwmtx.mtxmap[key]
	if ok {
		for client, _ := range mtxmap {
			rwmtx.mtxcurrmap[key] = client
			Printff("[PublicR] %s\n", client.conn.RemoteAddr())
			client.SendMsg(&NetMsg{Cmd: MUTEX_CMD_LOCK, Len: 0, Data: nil})
			//delete(mtxmap, client)
			return
		}

		rwmtx.mtxcurrmap[key] = nil
	}
}

func NewMutexServer(name string, addr string) *Mutex {
	if rwmtx, ok := rwmutexs[name]; !ok {
		rwmtx = &Mutex{
			state:      MUTEX_STATE_FREE,
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
				msg.Client.SendMsg(&NetMsg{Cmd: MUTEX_LOCK_EMPTY_KEY_ERR, Len: 0, Data: nil})
				return false
			}

			if !ok {
				mtxmap = make(map[*TcpClient]*TcpClient)
				rwmtx.mtxmap[key] = mtxmap
			}

			if _, ok2 := mtxmap[msg.Client]; ok2 {
				msg.Client.SendMsg(&NetMsg{Cmd: MUTEX_TWICE_LOCK_ERR, Len: 0, Data: nil})
				return false
			}
			mtxmap[msg.Client] = msg.Client
			if len(mtxmap) == 1 {
				rwmtx.mtxcurrmap[key] = msg.Client
				Printff("[HandleLock] %s\n", msg.Client.conn.RemoteAddr())
				msg.Client.SendMsg(&NetMsg{Cmd: MUTEX_CMD_LOCK, Len: 0, Data: nil})
			}

			return true
		}

		handleUnLock := func(msg *NetMsg) bool {
			Printff("[HandleUnLock] %s\n", msg.Client.conn.RemoteAddr())
			key := string(msg.Data)
			if key == "" {
				msg.Client.SendMsg(&NetMsg{Cmd: MUTEX_UNLOCK_EMPTY_KEY_ERR, Len: 0, Data: nil})
				return false
			}
			if curr, ok := rwmtx.mtxcurrmap[key]; ok && curr == msg.Client {
				delete(rwmtx.mtxmap[key], msg.Client)
				msg.Client.SendMsg(&NetMsg{Cmd: MUTEX_CMD_UNLOCK, Len: 0, Data: nil})
				rwmtx.PublicR(key)
				return true
			}

			msg.Client.SendMsg(&NetMsg{Cmd: MUTEX_INVALID_UNLOCK_ERR, Len: 0, Data: nil})
			return false
		}

		handleConnClose := func(client *TcpClient) {

		}
		rwmtx.server.SetConnCloseCB(handleConnClose)
		rwmtx.server.AddMsgHandler(MUTEX_CMD_LOCK, handleLock)
		rwmtx.server.AddMsgHandler(MUTEX_CMD_UNLOCK, handleUnLock)

		NewCoroutine(func() {
			rwmtx.server.Start(addr)
		})
		return rwmtx
	} else {
		ZLog("NewMutex Error: %s has been exist.", name)
	}
	return nil
}

func DeleMutex(name string) {
	if rwmtx, ok := rwmutexs[name]; ok {
		rwmtx.server.Stop()
		delete(rwmutexs, name)
	}
}

type MutexClient struct {
	mutex sync.Mutex
	addr  string
	conn  *net.TCPConn
	name  string
}

func (client *MutexClient) SendMsg(msg *NetMsg) bool {
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
		ZLog("MutexClient SetWriteDeadline Err: %v.", err)
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

func (client *MutexClient) ReadMsg() *NetMsg {
	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
		msg     *NetMsg
	)

	if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
		ZLog("MutexClient SetReadDeadline Err: %v.", err)
		goto Exit
	}

	readLen, err = io.ReadFull(client.conn, head)
	if err != nil || readLen < PACK_HEAD_LEN {
		ZLog("MutexClient Read Head Err: %v %d.", err, readLen)
		goto Exit
	}

	if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
		ZLog("MutexClient SetReadDeadline Err: %v.", err)
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

func (client *MutexClient) Lock(key string) {
	Printff("[Lock] %s %s 111\n", client.name)
	client.mutex.Lock()
	defer client.mutex.Unlock()

	Printff("[Lock] %s %s 222\n", client.name)
	if client.SendMsg(&NetMsg{Cmd: MUTEX_CMD_LOCK, Len: len(key), Data: []byte(key)}) {
		Printff("[Lock] %s %s 333\n", client.name, client.conn.LocalAddr())
		msg := client.ReadMsg()
		if msg == nil {
			panic(&ZMutexErr{errno: MUTEX_NET_ERR})
		}
		switch msg.Cmd {
		case MUTEX_TWICE_LOCK_ERR:
			panic(&ZMutexErr{errno: MUTEX_TWICE_LOCK_ERR})
		case MUTEX_LOCK_EMPTY_KEY_ERR:
			panic(&ZMutexErr{errno: MUTEX_LOCK_EMPTY_KEY_ERR})
		default:
		}
		Printff("[Lock] %s %s 444 cmd: %d, data: %s\n", client.name, client.conn.LocalAddr(), msg.Cmd, string(msg.Data))

	} else {
		panic(&ZMutexErr{errno: MUTEX_NET_ERR})
	}
}

func (client *MutexClient) UnLock(key string) {
	Printff("[UnLock] %s %s 111\n", client.name)
	client.mutex.Lock()
	defer client.mutex.Unlock()

	Printff("[UnLock] %s %s 222\n", client.name)
	if client.SendMsg(&NetMsg{Cmd: MUTEX_CMD_UNLOCK, Len: len(key), Data: []byte(key)}) {
		Printff("[UnLock] %s %s 333\n", client.name, client.conn.LocalAddr())
		msg := client.ReadMsg()
		if msg == nil {
			panic(&ZMutexErr{errno: MUTEX_NET_ERR})
		}
		switch msg.Cmd {
		case MUTEX_INVALID_UNLOCK_ERR:
			panic(&ZMutexErr{errno: MUTEX_INVALID_UNLOCK_ERR})
		case MUTEX_UNLOCK_EMPTY_KEY_ERR:
			panic(&ZMutexErr{errno: MUTEX_UNLOCK_EMPTY_KEY_ERR})
		default:
		}
		Printff("[UnLock] %s %s 444 cmd: %d, data: %s\n", client.name, client.conn.LocalAddr(), msg.Cmd, string(msg.Data))
	} else {
		panic(&ZMutexErr{errno: MUTEX_NET_ERR})
	}
}

func NewMutexClient(name string, addr string) *MutexClient {
	return &MutexClient{
		addr: addr,
		conn: nil,
		name: name,
	}
}

func DeleMutexClient(client *MutexClient) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.conn.Close()
}
