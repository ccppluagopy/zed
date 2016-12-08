package observer

import (
	"encoding/binary"
	"github.com/ccppluagopy/zed"
	//"fmt"
	"encoding/json"
	"io"
	"net"
	"sync"
	"time"
)

type MsgRegist struct {
	Key string `json:"Key"`
}

type MsgPublic struct {
	Key  string `json:"Key"`
	Data []byte `json:"Data"`
}

type ObserverServer struct {
	sync.RWMutex
	server *zed.TcpServer
	events map[string]map[*zed.TcpClient]bool
}

type ObserverClient struct {
	mutex  sync.Mutex
	addr   string
	name   string
	client *zed.TcpClient
}

const (
	ERR_REGIST_EMPTY_KEY = iota
	ERR_UNREGIST_EMPTY_KEY
	ERR_UNREGIST_INVALID_KEY
	ERR_PUBLIC_EMPTY_KEY
	ERR_PUBLIC_EMPTY_KEY_DATA
)

type OBServers struct {
	sync.Mutex
	servers map[string]*ObserverServer
}

func (servers *OBServers) GetServer(name string) *ObserverServer {
	servers.Lock()
	defer servers.Unlock()
	return servers.servers[name]
}

func (servers *OBServers) AddServer(name string, server *ObserverServer) {
	servers.Lock()
	defer servers.Unlock()
	if s, ok := servers.servers[name]; !ok {
		servers.servers[name] = server
	} else {
		zed.ZLog("OBServers AddServer Error: %s has been exist!", name)
	}

}

func (servers *OBServers) DeleServer(name string) {
	servers.Lock()
	defer servers.Unlock()
	delete(servers.servers, name)

}

var (
	observers = &OBServers{

		servers: make(map[string]*ObserverServer),
	}
	errconfig = make(map[int]string{
		ERR_REGIST_EMPTY_KEY:      "ERR_REGIST_EMPTY_KEY",
		ERR_UNREGIST_EMPTY_KEY:    "ERR_UNREGIST_EMPTY_KEY",
		ERR_UNREGIST_INVALID_KEY:  "ERR_UNREGIST_INVALID_KEY",
		ERR_PUBLIC_EMPTY_KEY:      "ERR_PUBLIC_EMPTY_KEY",
		ERR_PUBLIC_EMPTY_KEY_DATA: "ERR_PUBLIC_EMPTY_KEY_DATA",
	})
)

const (
	OBC_CMD_REGIST_REQ = iota
	OBS_CMD_REGIST_RSP

	OBC_CMD_UNREGIST_REQ
	OBS_CMD_UNREGIST_RSP

	OBC_CMD_PUBLIC_REQ
	OBS_CMD_PUBLIC_RSP
	OBS_CMD_PUBLIC
)

func SendError(client *zed.TcpClient, cmd zed.CmdType, errNum int) {
	if desc, ok := errconfig[errNum]; ok {
		rsp := make(map[string]string)

		rsp["Error"] = desc
		rspData, _ := json.Marshal(&rsp)

		client.SendMsg(&zed.NetMsg{
			Cmd:    cmd,
			Len:    len(rspData),
			Client: client,
			Data:   rspData,
		})
	} else {
		zed.ZLog("ObServerServer SendError Error: errNum not found.")
	}
}

type ZObserverServerErr struct {
	errno int
}

func (err *ZObserverServerErr) Error() string {
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

func (observer *ObserverServer) Public(msg *MsgPublic, args ...interface{}) {
	observer.Lock()
	defer observer.Unlock()

	obs, ok := observer.events[msg.Key]
	if ok {
		if len(args) == 1 {
			if data, ok := args[0].([]byte); ok {
				for ob, _ := range obs {
					ob.SendMsgAsync(&zed.NetMsg{Cmd: OBS_CMD_PUBLIC, Len: len(data), Data: data})
				}
			}
		} else if len(args) == 0 {
			if data, err := json.Marshal(msg); err == nil {
				for ob, _ := range obs {
					ob.SendMsgAsync(&zed.NetMsg{Cmd: OBS_CMD_PUBLIC, Len: len(data), Data: data})
				}
			} else {

			}
		} else {
			zed.ZLog("ObserverServer Public Error: Invalid args num: %d", len(args))
		}
	}
}

func NewObserverServer(name string, addr string) *ObserverServer {
	if observer := observers.GetServer(name); observer == nil {
		observer = &ObserverServer{
			server: zed.NewTcpServer(name),
			events: make(map[string]map[*zed.TcpClient]bool),
		}

		handleRegist := func(msg *zed.NetMsg) bool {
			if msg.Data == nil || msg.Len <= 0 {
				SendError(msg.Client, OBS_CMD_REGIST_RSP, ERR_REGIST_EMPTY_KEY)
				return true
			}

			observer.Lock()
			defer observer.Unlock()
			req := &MsgRegist{}
			if err := json.Unmarshal(msg.Data, req); err == nil {
				if req.Key == "" {
					SendError(msg.Client, OBS_CMD_REGIST_RSP, ERR_REGIST_EMPTY_KEY)
					return true
				} else {
					obs, ok := observer.events[msg.Key]
					if !ok {
						obs = make(map[*zed.TcpClient]bool)
						observer.events[msg.Key] = obs
					}
					obs[msg.Client] = true
					msg.Client.AddCloseCB(zed.Spritf("rme%s", req.Key), func(c *zed.TcpClient) {
						delete(obs, msg.Client)
					})
					msg.Client.SendMsgAsync(&zed.NetMsg{Cmd: OBS_CMD_REGIST_RSP})
				}
			}
			return true
		}

		handleUnregist := func(msg *zed.NetMsg) bool {
			if msg.Data == nil || msg.Len <= 0 {
				SendError(msg.Client, OBS_CMD_UNREGIST_RSP, ERR_UNREGIST_EMPTY_KEY)
				return true
			}

			observer.Lock()
			defer observer.Unlock()
			req := &MsgRegist{}
			if err := json.Unmarshal(msg.Data, req); err == nil {
				if req.Key == "" {
					SendError(msg.Client, OBS_CMD_UNREGIST_RSP, ERR_UNREGIST_EMPTY_KEY)
					return true
				} else {
					obs, ok := observer.events[msg.Key]
					if !ok {
						SendError(msg.Client, OBS_CMD_UNREGIST_RSP, ERR_UNREGIST_INVALID_KEY)
					} else {
						delete(obs, msg.Client)
						msg.Client.RemoveCloseCB(zed.Spritf("rme%s", req.Key))
						msg.Client.SendMsgAsync(&zed.NetMsg{Cmd: OBS_CMD_UNREGIST_RSP})
					}
				}
			}
			return true
		}

		handlePublic := func(msg *zed.NetMsg) bool {
			if msg.Data == nil || msg.Len <= 0 {
				SendError(msg.Client, OBS_CMD_PUBLIC_RSP, ERR_PUBLIC_EMPTY_KEY_DATA)
				return true
			}
			req := &MsgRegist{}
			if err := json.Unmarshal(msg.Data, req); err == nil {
				if req.Key == "" {
					SendError(msg.Client, OBS_CMD_PUBLIC_RSP, ERR_PUBLIC_EMPTY_KEY)
					return true
				} else {
					observer.Lock()
					defer observer.Unlock()

					obs, ok := observer.events[req.Key]
					if ok {
						for ob, _ := range obs {
							ob.SendMsgAsync(&zed.NetMsg{Cmd: OBS_CMD_PUBLIC, Len: len(msg.Data), Data: msg.Data})
						}
					}
				}
			}
			return true
		}

		handleConnClose := func(client *zed.TcpClient) {

		}

		observer.server.SetConnCloseCB(handleConnClose)
		observer.server.AddMsgHandler(OBC_CMD_REGIST_REQ, handleRegist)
		observer.server.AddMsgHandler(OBC_CMD_UNREGIST_REQ, handleUnregist)
		observer.server.AddMsgHandler(OBC_CMD_PUBLIC_REQ, handlePublic)

		zed.NewCoroutine(func() {
			observer.server.Start(addr)
		})

		observers.AddServer(name, observer)
		return observer
	} else {
		zed.ZLog("NewObserverServer Error: %s has been exist.", name)
	}
	return nil
}

func DeleObserverServer(name string) {
	observers.DeleServer(name)
}

func NewOBClient(name string, addr string) *ObserverClient {

	client := &zed.TcpClient{
		conn:    conn,
		parent:  parent,
		ID:      NullID,
		Idx:     parent.ClientNum,
		Addr:    addr,
		closeCB: make(map[interface{}]ClientCloseCB),
		chSend:  make(chan *AsyncMsg, 10),
		Valid:   false,
		running: true,
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil
	}

	client.StartReader()
	client.StartWriter()

	observer := &ObserverClient{
		addr:   addr,
		client: client,
		name:   name,
	}
}

func DeleOBClient(client *MutexClient) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.conn.Close()
}
