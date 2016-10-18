package mutex

import (
	//"encoding/binary"
	"github.com/ccppluagopy/zed"
	//"io"
	//"net"
	"sync"
	//"time"
)

var (
	rwmutexs       = make(map[string]*RWMutex)
	rwmutexconnops = make(map[*zed.TcpClient]map[string]int)
)

const (
	RWMUTEX_STATE_FREE = iota
	RWMUTEX_STATE_READING
	RWMUTEX_STATE_WRITING
)
const (
	RWMUTEX_CMD_LOCK = iota
	RWMUTEX_CMD_UNLOCK
	RWMUTEX_CMD_LOCK_ERR
	RWMUTEX_CMD_UNLOCK_ERR

	RWMUTEX_NET_ERR
	RWMUTEX_LOCK_EMPTY_KEY_ERR
	RWMUTEX_TWICE_LOCK_ERR
	RWMUTEX_LOCK_RECV_INVALID_CMD_ERR
	RWMUTEX_INVALID_UNLOCK_ERR
	RWMUTEX_UNLOCK_EMPTY_KEY_ERR
	RWMUTEX_UNLOCK_RECV_INVALID_CMD_ERR
)

type RWMutex struct {
	sync.Mutex
	state        int
	server       *zed.TcpServer
	rwmtxrmap    map[string]map[*zed.TcpClient]int
	rwmtxwmap    map[string]map[*zed.TcpClient]int
	rwmtxcurrmap map[string]*zed.TcpClient
}

type ZRWMutexErr struct {
	errno int
}

func (err *ZRWMutexErr) Error() string {
	switch err.errno {
	case RWMUTEX_NET_ERR:
		return "Error: ZRWMutex Operation Net Unavailable."
	case RWMUTEX_LOCK_EMPTY_KEY_ERR:
		return "Error: ZRWMutex Lock key is empty."
	case RWMUTEX_TWICE_LOCK_ERR:
		return "Error: ZRWMutex Twice Lock."
	case RWMUTEX_LOCK_RECV_INVALID_CMD_ERR:
		return "Error: ZRWMutex Lock Recv Invalid Cmd."
	case RWMUTEX_INVALID_UNLOCK_ERR:
		return "Error: ZRWMutex Invalid UnLock Operation."
	case RWMUTEX_UNLOCK_EMPTY_KEY_ERR:
		return "Error: ZRWMutex UnLock key is empty."
	case RWMUTEX_UNLOCK_RECV_INVALID_CMD_ERR:
		return "Error: ZRWMutex UnLock Recv Invalid Cmd."
	}

	return "ZMutexError"
}

func (rwmtx *RWMutex) Public(key string) {
	rwmtxwmap, ok := rwmtx.rwmtxwmap[key]
	if ok {
		for client, _ := range rwmtxwmap {
			rwmtx.rwmtxcurrmap[key] = client
			Printff("[Public] %s\n", client.GetConn().RemoteAddr())
			client.SendMsg(&zed.NetMsg{Cmd: MUTEX_CMD_LOCK, Len: 0, Data: nil})
			rwmtx.state = RWMUTEX_STATE_WRITING
			return
		}

		for client, _ := range rwmtx.rwmtxrmap[key] {
			rwmtx.rwmtxcurrmap[key] = client
			Printff("[Public] %s\n", client.GetConn().RemoteAddr())
			client.SendMsg(&zed.NetMsg{Cmd: MUTEX_CMD_LOCK, Len: 0, Data: nil})
			rwmtx.state = RWMUTEX_STATE_READING
			return
		}

		rwmtx.state = RWMUTEX_STATE_FREE
		rwmtx.rwmtxcurrmap[key] = nil
	}
}

func NewRWMutexServer(name string, addr string) *RWMutex {
	if rwmtx, ok := rwmutexs[name]; !ok {
		rwmtx = &RWMutex{
			state:        RWMUTEX_STATE_FREE,
			server:       zed.NewTcpServer(name),
			rwmtxrmap:    make(map[string]map[*zed.TcpClient]int),
			rwmtxwmap:    make(map[string]map[*zed.TcpClient]int),
			rwmtxcurrmap: make(map[string]*zed.TcpClient),
		}

		handleLock := func(msg *zed.NetMsg) bool {
			rwmtx.Lock()
			defer rwmtx.Unlock()

			key := string(msg.Data)
			if key == "" {
				msg.Client.SendMsg(&zed.NetMsg{Cmd: RWMUTEX_LOCK_EMPTY_KEY_ERR, Len: 0, Data: nil})
				return false
			}

			rwmtxwmap, ok := rwmtx.rwmtxwmap[key]
			if !ok {
				rwmtxwmap = make(map[*zed.TcpClient]int)
				rwmtx.rwmtxwmap[key] = rwmtxwmap
			}

			if _, ok2 := rwmtxwmap[msg.Client]; ok2 {
				msg.Client.SendMsg(&zed.NetMsg{Cmd: RWMUTEX_TWICE_LOCK_ERR, Len: 0, Data: nil})
				return false
			}

			rwmtxrmap, ok3 := rwmtx.rwmtxrmap[key]
			if ok3 {
				if _, ok4 := rwmtxrmap[msg.Client]; ok4 {
					msg.Client.SendMsg(&zed.NetMsg{Cmd: RWMUTEX_TWICE_LOCK_ERR, Len: 0, Data: nil})
					return false
				}
			}

			rwmtxwmap[msg.Client] = RWMUTEX_STATE_WRITING

			_, ok4 := rwmtx.rwmtxcurrmap[key]
			if !ok4 {
				if len(rwmtxrmap) == 0 {
					rwmtx.rwmtxcurrmap[key] = msg.Client
					msg.Client.SendMsg(&zed.NetMsg{Cmd: RWMUTEX_CMD_LOCK, Len: 0, Data: nil})
				}
			}

			return true
		}

		handleUnLock := func(msg *zed.NetMsg) bool {
			rwmtx.Lock()
			defer rwmtx.Unlock()

			key := string(msg.Data)
			if key == "" {
				msg.Client.SendMsg(&zed.NetMsg{Cmd: RWMUTEX_UNLOCK_EMPTY_KEY_ERR, Len: 0, Data: nil})
				return false
			}

			if curr, ok := rwmtx.rwmtxcurrmap[key]; ok && curr == msg.Client {
				if rwmtxwmap, ok := rwmtx.rwmtxwmap[key]; ok {
					delete(rwmtxwmap, msg.Client)
				}

				msg.Client.SendMsg(&zed.NetMsg{Cmd: RWMUTEX_CMD_UNLOCK, Len: 0, Data: nil})
				rwmtx.Public(key)
				return true
			}

			msg.Client.SendMsg(&zed.NetMsg{Cmd: RWMUTEX_INVALID_UNLOCK_ERR, Len: 0, Data: nil})
			return false
		}

		handleConnClose := func(client *zed.TcpClient) {

		}
		rwmtx.server.SetConnCloseCB(handleConnClose)
		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_LOCK, handleLock)
		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_UNLOCK, handleUnLock)

		zed.NewCoroutine(func() {
			rwmtx.server.Start(addr)
		})
		return rwmtx
	} else {
		zed.ZLog("NewMutex Error: %s has been exist.", name)
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
	MutexClient
}

func (client *RWMutexClient) Lock(key string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.SendMsg(&zed.NetMsg{Cmd: MUTEX_CMD_LOCK, Len: len(key), Data: []byte(key)}) {
		msg := client.ReadMsg()
		if msg == nil {
			panic(&ZRWMutexErr{errno: MUTEX_NET_ERR})
		}
		switch msg.Cmd {
		case MUTEX_CMD_LOCK:
			return
		case MUTEX_TWICE_LOCK_ERR:
			panic(&ZRWMutexErr{errno: MUTEX_TWICE_LOCK_ERR})
		case MUTEX_LOCK_EMPTY_KEY_ERR:
			panic(&ZRWMutexErr{errno: MUTEX_LOCK_EMPTY_KEY_ERR})
		default:
			panic(&ZRWMutexErr{errno: MUTEX_LOCK_RECV_INVALID_CMD_ERR})
		}
	} else {
		panic(&ZRWMutexErr{errno: MUTEX_NET_ERR})
	}
}

func (client *RWMutexClient) UnLock(key string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.SendMsg(&zed.NetMsg{Cmd: MUTEX_CMD_UNLOCK, Len: len(key), Data: []byte(key)}) {
		msg := client.ReadMsg()
		if msg == nil {
			panic(&ZRWMutexErr{errno: MUTEX_NET_ERR})
		}
		switch msg.Cmd {
		case MUTEX_CMD_UNLOCK:

		case MUTEX_INVALID_UNLOCK_ERR:
			panic(&ZRWMutexErr{errno: MUTEX_INVALID_UNLOCK_ERR})
		case MUTEX_UNLOCK_EMPTY_KEY_ERR:
			panic(&ZRWMutexErr{errno: MUTEX_UNLOCK_EMPTY_KEY_ERR})
		default:
			panic(&ZRWMutexErr{errno: MUTEX_UNLOCK_RECV_INVALID_CMD_ERR})
		}
	} else {
		panic(&ZRWMutexErr{errno: MUTEX_NET_ERR})
	}
}

func NewRWMutexClient(name string, addr string) *RWMutexClient {
	return &RWMutexClient{
		MutexClient: MutexClient{
			addr: addr,
			conn: nil,
			name: name,
		},
	}
}

func DeleRWMutexClient(client *RWMutexClient) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.conn.Close()
}

/*
func TestRWMutex() {
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
*/
