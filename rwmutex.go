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
			wteam, ok2 := rwmtx.rmap[key]

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

				if !ok2 || len(wteam.items) == 0 {
					if rwmtx.state != RWMUTEX_STATE_WRITING {
						rwmtx.state = RWMUTEX_STATE_READING
						msg.Client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RLOCK, Len: 0, Data: nil})
					}
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
			wteam, ok2 := rwmtx.wmap[key]
			if msg.Len <= 0 || !ok {
				goto Err
			}

			if _, ok := rteam.items[msg.Client]; !ok {
				goto Err
			} else {
				if ok2 {
					for _, cli := range wteam.items {
						delete(wteam.items, cli)
						cli.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_LOCK, Len: 0, Data: nil})
						rwmtx.state = RWMUTEX_STATE_WRITING
						return true
					}
				}
				for _, cli := range rteam.items {
					delete(rteam.items, cli)
					cli.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RLOCK, Len: 0, Data: nil})
					return true
				}

				rwmtx.state = RWMUTEX_STATE_FREE
			}

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
			ops := rwmutexconnops[client]

			for key, _ := range ops {
				rteam, ok1 := rwmtx.rmap[key]
				wteam, ok2 := rwmtx.wmap[key]

				if ok2 {
					for _, cli := range wteam.items {
						cli.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RLOCK, Len: 0, Data: nil})
						rwmtx.state = RWMUTEX_STATE_WRITING
						goto CLEAR
					}
				}
				if ok1 {
					for _, cli := range rteam.items {
						cli.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_LOCK, Len: 0, Data: nil})
						rwmtx.state = RWMUTEX_STATE_READING
						goto CLEAR
					}
				}
			CLEAR:
				if ok1 {
					delete(rteam.items, client)
				}
				if ok2 {
					delete(wteam.items, client)
				}
			}

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
