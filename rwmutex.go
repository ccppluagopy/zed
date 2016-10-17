package zed

import (
	"encoding/binary"
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
	items map[*TcpClient]*TcpClient
}

type RWMutex struct {
	sync.RWMutex
	state  int
	server *TcpServer
	rmap   map[string]*RWMutexTeam
	wmap   map[string]*RWMutexTeam
}

func NewRWMutexServer(name string, addr string) *RWMutex {
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

func DeleRWMutex(name string) {
	if rwmtx, ok := rwmutexs[name]; ok {
		rwmtx.server.Stop()
		delete(rwmutexs, name)
	}
}

type RWMutexClient struct {
	sync.RWMutex
	addr string
	conn *net.TCPConn
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

func (client *RWMutexClient) RLock() bool {
	client.Lock()
	defer client.Unlock()
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
	if client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RLOCK, Len: 0, Data: nil}) {
		if msg := client.ReadMsg(); msg != nil && msg.Cmd == RWMUTEX_CMD_RLOCK {
			return true
		}
	}

	return false
}

func (client *RWMutexClient) RUnLock() bool {
	client.Lock()
	defer client.Unlock()

	if client.SendMsg(&NetMsg{Cmd: RWMUTEX_CMD_RUNLOCK, Len: 0, Data: nil}) {
		if msg := client.ReadMsg(); msg != nil && msg.Cmd == RWMUTEX_CMD_RUNLOCK {
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
	client.Lock()
	defer client.Unlock()

	client.conn.Close()
}
