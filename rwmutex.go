package zed

/*
import (
	"sync"
)

var (
	rwmutexs = make(map[string]*RWMutex)
)

const (
	RWMUTEX_STATE_FREE = iota
	RWMUTEX_STATE_READING
	RWMUTEX_STATE_WRITING
)
const (
	RWMUTEX_CMD_RLOCK = iota
	RWMUTEX_CMD_RUNLOCK
	RWMUTEX_CMD_LOCK
	RWMUTEX_CMD_UNLOCK
)

type RWMutex struct {
	sync.RWMutex
	state  int
	server *TcpServer
	rmap   map[string]map[*TcpClient]*TcpClient
	wmap   map[string]map[*TcpClient]*TcpClient
}

func NewRWMutex(name string, addr string) *RWMutex {
	if rwmtx, ok := rwmutexs[name]; !ok {
		rwmtx = &RWMutex{
			state:  RWMUTEX_STATE_FREE,
			server: zed.NewTcpServer(name),
			rmap:   make(map[string]map[*TcpClient]*TcpClient),
			wmap:   make(map[string]map[*TcpClient]*TcpClient),
		}

		handleRLock := func(msg *NetMsg) bool {
			if msg.Len <= 0 {
				return false
			}
			key := string(msg.Data)
			ritems, ok := rwmtx.rmap[key]
			if !ok {
				ritems = make(map[*TcpClient]*TcpClient)
			}
			if ritem, ok := ritems[key]; !ok {
				ritems[msg.Client] = msg.Client
			} else {
				return false
			}
			return true
		}
		handleRUnLock := func(msg *NetMsg) bool {
			if msg.Len <= 0 {
				return false
			}
		}
		handleLock := func(msg *NetMsg) bool {
			if msg.Len <= 0 {
				return false
			}
		}
		handleUnLock := func(msg *NetMsg) bool {
			if msg.Len <= 0 {
				return false
			}
		}

		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_RLOCK, handleRLock)
		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_RUNLOCK, handleRUnLock)
		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_LOCK, handleLock)
		rwmtx.server.AddMsgHandler(RWMUTEX_CMD_UNLOCK, handleUnLock)

		NewCoroutine(func() {
			rwmtx.server.Start(addr)
		})
		return rwmtx
	} else {
		ZLog("NewRWMutex Error: %s has been exist.", name)
	}
	return nil
}
*/
