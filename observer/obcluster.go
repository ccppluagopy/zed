package observer

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type OBAgrs struct {
	Addr      string
	Heartbeat time.Duration
}
type OBRsp struct {
}

type OBClusterError struct {
	description string
}

func (mgrErr *OBClusterError) Error() string {
	return mgrErr.description
}

type OBClusterMgr struct {
	sync.Mutex
	Addr    string
	Clients map[string]*ObserverClient
}

func (mgr *OBClusterMgr) AddNode(args *OBAgrs, reply *([]OBRsp)) error {
	mgr.Lock()
	defer mgr.Unlock()

	addr, heartbeat := args.Addr, args.Heartbeat
	if _, ok := mgr.Clients[addr]; ok {
		return &OBClusterError{
			description: fmt.Sprintf("OBNode who's addr is %s has been exist.", addr),
		}
	}

	ename := fmt.Sprintf("obclustermgr_%s", addr)
	client := NewOBClient(addr, ename, heartbeat)
	client.SetCloseCB(func() {
		mgr.Lock()
		defer mgr.Unlock()
		delete(mgr.Clients, addr)
	})

	client.Regist(EventAll, nil)
	client.NewListener(EventAll, EventAll, func(e interface{}, args []interface{}) {
		for _, c := range mgr.Clients {
			eve, ok1 := e.(string)
			if ok1 {
				data, ok2 := args[0].([]byte)
				if ok2 {
					c.Publish2(eve, data)
				}
			}

		}

	})

	mgr.Clients[addr] = client

	return nil
}

func (mgr *OBClusterMgr) DeleteNode(args *OBAgrs, reply *([]OBRsp)) error {
	mgr.Lock()
	defer mgr.Unlock()

	delete(mgr.Clients, args.Addr)

	return nil
}

func NewOBClusterMgr(addr string) *OBClusterMgr {
	mgr := &OBClusterMgr{
		Addr:    addr,
		Clients: make(map[string]*ObserverClient),
	}

	rpc.Register(mgr)
	rpc.HandleHTTP()

	listener, e := net.Listen("tcp", mgr.Addr)
	if e != nil {
		zed.ZLog("NewOBClusterMgr Listen error: %v", e)
		return nil
	}

	http.Serve(listener, nil)

	return mgr
}

type OBClusterNode struct {
	MgrAddr  string
	NodeAddr string
	Server   *ObserverServer
}

func (node *OBClusterNode) Start() bool {
	go node.Server.Start(node.NodeAddr)
	time.Sleep(time.Second / 2)

	client, err := rpc.DialHTTP("tcp", node.MgrAddr)
	if err != nil {
		zed.ZLog("OBClusterNode Start DialHTTP Error: %v", err)
		return false
	}

	args := &OBAgrs{
		Addr:      node.NodeAddr,
		Heartbeat: DEFAULT_HEART_BEAT_TIME - time.Second*3,
	}

	err = client.Call("OBClusterNode Start OBClusterMgr AddNode", args, nil)
	if err != nil {
		zed.ZLog("OBClusterNode Start OBClusterMgr AddNode Error: %v", err)
		return false
	}

	return true
}

func NewOBClusterNode(mgraddr string, nodeaddr string, heartbeat time.Duration) *OBClusterNode {
	return &OBClusterNode{
		MgrAddr:  mgraddr,
		NodeAddr: nodeaddr,
		Server:   NewOBServer(fmt.Sprintf("OBClusterNode_%d", time.Nanosecond)),
	}

}
