package mutex

import (
	"encoding/binary"
	"github.com/ccppluagopy/zed"
	//"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	rwmutexs       = make(map[string]*Mutex)
	rwmutexconnops = make(map[*zed.TcpClient]map[string]bool)
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
	MUTEX_LOCK_RECV_INVALID_CMD_ERR
	MUTEX_INVALID_UNLOCK_ERR
	MUTEX_UNLOCK_EMPTY_KEY_ERR
	MUTEX_UNLOCK_RECV_INVALID_CMD_ERR
)

type Mutex struct {
	sync.Mutex
	state      int
	server     *zed.TcpServer
	mtxmap     map[string]map[*zed.TcpClient]*zed.TcpClient
	mtxcurrmap map[string]*zed.TcpClient
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
	case MUTEX_LOCK_RECV_INVALID_CMD_ERR:
		return "Error: ZMutex Lock Recv Invalid Cmd."
	case MUTEX_INVALID_UNLOCK_ERR:
		return "Error: ZMutex Invalid UnLock Operation."
	case MUTEX_UNLOCK_EMPTY_KEY_ERR:
		return "Error: ZMutex UnLock key is empty."
	case MUTEX_UNLOCK_RECV_INVALID_CMD_ERR:
		return "Error: ZMutex UnLock Recv Invalid Cmd."
	}

	return "ZMutexError"
}

func (rwmtx *Mutex) PublicR(key string) {
	mtxmap, ok := rwmtx.mtxmap[key]
	if ok {
		for client, _ := range mtxmap {
			rwmtx.mtxcurrmap[key] = client
			Printff("[PublicR] %s\n", client.GetConn().RemoteAddr())
			client.SendMsg(&zed.NetMsg{Cmd: MUTEX_CMD_LOCK, Len: 0, Data: nil})
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
			server:     zed.NewTcpServer(name),
			mtxmap:     make(map[string]map[*zed.TcpClient]*zed.TcpClient),
			mtxcurrmap: make(map[string]*zed.TcpClient),
		}

		handleLock := func(msg *zed.NetMsg) bool {
			rwmtx.Lock()
			defer rwmtx.Unlock()

			key := string(msg.Data)
			mtxmap, ok := rwmtx.mtxmap[key]
			if key == "" {
				msg.Client.SendMsg(&zed.NetMsg{Cmd: MUTEX_LOCK_EMPTY_KEY_ERR, Len: 0, Data: nil})
				return false
			}

			if !ok {
				mtxmap = make(map[*zed.TcpClient]*zed.TcpClient)
				rwmtx.mtxmap[key] = mtxmap
			}

			if _, ok2 := mtxmap[msg.Client]; ok2 {
				msg.Client.SendMsg(&zed.NetMsg{Cmd: MUTEX_TWICE_LOCK_ERR, Len: 0, Data: nil})
				return false
			}
			mtxmap[msg.Client] = msg.Client
			if len(mtxmap) == 1 {
				rwmtx.mtxcurrmap[key] = msg.Client
				msg.Client.SendMsg(&zed.NetMsg{Cmd: MUTEX_CMD_LOCK, Len: 0, Data: nil})
			}

			return true
		}

		handleUnLock := func(msg *zed.NetMsg) bool {
			key := string(msg.Data)
			if key == "" {
				msg.Client.SendMsg(&zed.NetMsg{Cmd: MUTEX_UNLOCK_EMPTY_KEY_ERR, Len: 0, Data: nil})
				return false
			}
			if curr, ok := rwmtx.mtxcurrmap[key]; ok && curr == msg.Client {
				delete(rwmtx.mtxmap[key], msg.Client)
				msg.Client.SendMsg(&zed.NetMsg{Cmd: MUTEX_CMD_UNLOCK, Len: 0, Data: nil})
				rwmtx.PublicR(key)
				return true
			}

			msg.Client.SendMsg(&zed.NetMsg{Cmd: MUTEX_INVALID_UNLOCK_ERR, Len: 0, Data: nil})
			return false
		}

		handleConnClose := func(client *zed.TcpClient) {

		}
		rwmtx.server.SetConnCloseCB(handleConnClose)
		rwmtx.server.AddMsgHandler(MUTEX_CMD_LOCK, handleLock)
		rwmtx.server.AddMsgHandler(MUTEX_CMD_UNLOCK, handleUnLock)

		zed.NewCoroutine(func() {
			rwmtx.server.Start(addr)
		})
		return rwmtx
	} else {
		zed.ZLog("NewMutex Error: %s has been exist.", name)
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

func (client *MutexClient) SendMsg(msg *zed.NetMsg) bool {
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

	if err := (*client.conn).SetWriteDeadline(time.Now().Add(zed.WRITE_BLOCK_TIME)); err != nil {
		zed.ZLog("MutexClient SetWriteDeadline Err: %v.", err)
		goto Exit
	}

	buf = make([]byte, zed.PACK_HEAD_LEN+msg.Len)
	binary.LittleEndian.PutUint32(buf, uint32(msg.Len))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))

	writeLen, err = client.conn.Write(buf)

	if err == nil && writeLen == len(buf) {
		return true
	}

Exit:
	return false
}

func (client *MutexClient) ReadMsg() *zed.NetMsg {
	var (
		head    = make([]byte, zed.PACK_HEAD_LEN)
		readLen = 0
		err     error
		msg     *zed.NetMsg
	)

	if err = (*client.conn).SetReadDeadline(time.Now().Add(zed.READ_BLOCK_TIME)); err != nil {
		zed.ZLog("MutexClient SetReadDeadline Err: %v.", err)
		goto Exit
	}

	readLen, err = io.ReadFull(client.conn, head)
	if err != nil || readLen < zed.PACK_HEAD_LEN {
		zed.ZLog("MutexClient Read Head Err: %v %d.", err, readLen)
		goto Exit
	}

	/*if err = (*client.conn).SetReadDeadline(time.Now().Add(zed.READ_BLOCK_TIME)); err != nil {
		zed.ZLog("MutexClient SetReadDeadline Err: %v.", err)
		goto Exit
	}*/

	msg = &zed.NetMsg{
		Cmd:  zed.CmdType(binary.LittleEndian.Uint32(head[4:8])),
		Len:  0,
		Data: nil,
	}

	return msg

Exit:
	return nil
}

func (client *MutexClient) Lock(key string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.SendMsg(&zed.NetMsg{Cmd: MUTEX_CMD_LOCK, Len: len(key), Data: []byte(key)}) {
		msg := client.ReadMsg()
		if msg == nil {
			panic(&ZMutexErr{errno: MUTEX_NET_ERR})
		}
		switch msg.Cmd {
		case MUTEX_CMD_LOCK:

		case MUTEX_TWICE_LOCK_ERR:
			panic(&ZMutexErr{errno: MUTEX_TWICE_LOCK_ERR})
		case MUTEX_LOCK_EMPTY_KEY_ERR:
			panic(&ZMutexErr{errno: MUTEX_LOCK_EMPTY_KEY_ERR})
		default:
			panic(&ZMutexErr{errno: MUTEX_LOCK_RECV_INVALID_CMD_ERR})
		}
	} else {
		panic(&ZMutexErr{errno: MUTEX_NET_ERR})
	}
}

func (client *MutexClient) UnLock(key string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.SendMsg(&zed.NetMsg{Cmd: MUTEX_CMD_UNLOCK, Len: len(key), Data: []byte(key)}) {
		msg := client.ReadMsg()
		if msg == nil {
			panic(&ZMutexErr{errno: MUTEX_NET_ERR})
		}
		switch msg.Cmd {
		case MUTEX_CMD_UNLOCK:

		case MUTEX_INVALID_UNLOCK_ERR:
			panic(&ZMutexErr{errno: MUTEX_INVALID_UNLOCK_ERR})
		case MUTEX_UNLOCK_EMPTY_KEY_ERR:
			panic(&ZMutexErr{errno: MUTEX_UNLOCK_EMPTY_KEY_ERR})
		default:
			panic(&ZMutexErr{errno: MUTEX_UNLOCK_RECV_INVALID_CMD_ERR})
		}
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

func TestMutex() {
	const (
		key  = "key"
		addr = "127.0.0.1:33333"
	)

	mutex1 := func() {
		client := NewMutexClient("mutex1", addr)
		for {
			client.Lock("")
			client.Lock(key)
			time.Sleep(time.Second * 1)
			client.UnLock(key)
		}
	}

	mutex2 := func() {
		time.Sleep(time.Second)
		client := NewMutexClient("mutex2", addr)
		n := 0
		for {
			client.Lock(key)
			client.UnLock(key)
			n = n + 1
			zed.Println("mutex2 action .......", n)
		}
	}

	mutex3 := func() {
		time.Sleep(time.Second * 1)
		client := NewMutexClient("mutex3", addr)
		n := 0
		for {
			client.Lock(key)
			client.UnLock(key)
			n = n + 1
			zed.Println("mutex3 action .......", n)
			time.Sleep(time.Second * 1)
		}
	}

	NewMutexServer("test", addr)
	time.Sleep(time.Second)

	go mutex1()
	go mutex2()
	go mutex3()

	time.Sleep(time.Hour)
}
