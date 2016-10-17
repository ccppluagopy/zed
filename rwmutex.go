package zed

import (
	"sync"
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
	items map[*TcpClient]*TcpClient
}

type RWMutex struct {
	sync.RWMutex
	state  int
	server *TcpServer
	rmap   map[string]*RWMutexTeam
	wmap   map[string]*RWMutexTeam
}

func NewRWMutex(name string, addr string) *RWMutex {
	if rwmtx, ok := rwmutexs[name]; !ok {
		rwmtx = &RWMutex{
			state:  RWMUTEX_STATE_FREE,
			server: NewTcpServer(name),
			rmap:   make(map[string]*RWMutexTeam),
			wmap:   make(map[string]*RWMutexTeam),
		}

		handleRLock := func(msg *NetMsg) bool {
			rwmtx.Lock()
			defer rwmtx.Unlock()

			key := string(msg.Data)
			rteam, ok := rwmtx.rmap[key]

			if msg.Len <= 0 {
				goto Err
			}

			if !ok {
				rteam = &RWMutexTeam{
					items: make(map[*TcpClient]*TcpClient),
				}
				rwmtx.rmap[key] = rteam
			}
			if _, ok := rteam.items[msg.Client]; !ok {
				rteam.items[msg.Client] = msg.Client
				ops, ok := rwmutexconnops[msg.Client]
				if !ok {
					ops = make(map[string]bool)
					rwmutexconnops[msg.Client] = ops
				}
				if _, ok := ops[key]; ok {
					goto Err
				}
				ops[key] = true

				if rwmtx.state != RWMUTEX_STATE_WRITING {
					rwmtx.state = RWMUTEX_STATE_READING
					msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RLOCK, Len: 0, Data: nil})
				}
			} else {
				goto Err
			}
			return true
		Err:
			msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RLOCK_ERR, Len: 0, Data: nil})
			return false
		}

		handleRUnLock := func(msg *NetMsg) bool {
			rwmtx.Lock()
			defer rwmtx.Unlock()

			key := string(msg.Data)
			rteam, ok := rwmtx.rmap[key]

			if msg.Len <= 0 {
				goto Err
			}

			if !ok {
				rteam = &RWMutexTeam{
					items: make(map[*TcpClient]*TcpClient),
				}
				rwmtx.rmap[key] = rteam
			}
			if _, ok := rteam.items[msg.Client]; !ok {
				rteam.items[msg.Client] = msg.Client
				ops, ok := rwmutexconnops[msg.Client]
				if !ok {
					ops = make(map[string]bool)
					rwmutexconnops[msg.Client] = ops
				}
				if _, ok := ops[key]; ok {
					goto Err
				}
				ops[key] = true

				if rwmtx.state != RWMUTEX_STATE_WRITING {
					rwmtx.state = RWMUTEX_STATE_READING
					msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RLOCK, Len: 0, Data: nil})
				}
			} else {
				goto Err
			}
			return true

			return true
		Err:
			msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RUNLOCK_ERR, Len: 0, Data: nil})
			return false
		}

		handleLock := func(msg *NetMsg) bool {
			rwmtx.Lock()
			defer rwmtx.Unlock()

			key := string(msg.Data)
			wteam, ok := rwmtx.wmap[key]

			if msg.Len <= 0 {
				goto Err
			}

			if !ok {
				wteam = &RWMutexTeam{
					items: make(map[*TcpClient]*TcpClient),
				}
				rwmtx.wmap[key] = wteam
			}
			if _, ok := wteam.items[msg.Client]; !ok {
				wteam.items[msg.Client] = msg.Client
				ops, ok := rwmutexconnops[msg.Client]
				if !ok {
					ops = make(map[string]bool)
					rwmutexconnops[msg.Client] = ops
				}
				if _, ok := ops[key]; ok {
					goto Err
				}
				ops[key] = true

				if rwmtx.state == RWMUTEX_STATE_FREE {
					rwmtx.state = RWMUTEX_STATE_WRITING
					msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_LOCK, Len: 0, Data: nil})
				}
			} else {
				goto Err
			}
			return true
		Err:
			msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_LOCK_ERR, Len: 0, Data: nil})
			return false
		}

		handleUnLock := func(msg *NetMsg) bool {
			if msg.Len <= 0 {
				goto Err
			}

			return true
		Err:
			msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RLOCK_ERR, Len: 0, Data: nil})
			return false
		}

		handleConnClose := func(client *TcpClient) {
			delete(rwmutexconnops, client)
		}
		rwmtx.server.SetConnCloseCB(handleConnClose)
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
