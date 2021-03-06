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
	mutex    sync.Mutex
	Addr     string
	Listener net.Listener
	Clients  map[string]*ObserverClient
}

func (mgr *OBClusterMgr) AddNode(args *OBAgrs, reply *([]OBRsp)) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	addr, heartbeat := args.Addr, args.Heartbeat
	if _, ok := mgr.Clients[addr]; ok {
		return &OBClusterError{
			description: fmt.Sprintf("OBNode who's addr is %s has been exist.", addr),
		}
	}

	ename := fmt.Sprintf("obclustermgr_%s", addr)
	client := NewOBClient(addr, ename, heartbeat)
	client.SetCloseCB(func() {
		mgr.mutex.Lock()
		defer mgr.mutex.Unlock()
		delete(mgr.Clients, addr)
	})

	if client.RegistCluster() {
		client.NewListener(EventAll, EventAll, func(e interface{}, args []interface{}) {
			//mgr.mutex.Lock()
			//defer mgr.mutex.Unlock()
			eve, ok1 := e.(string)
			if ok1 {
				//fmt.Println("Cluster Mgr On Event --------------------", eve)
				data, ok2 := args[0].([]byte)
				if ok2 {
					//fmt.Println("Cluster Mgr On Event 00000000000000000", eve, data)
					for pubaddr, c := range mgr.Clients {
						//fmt.Println("Cluster Mgr On Event xxxxxxxxxxxxxx 111", pubaddr, addr)
						if pubaddr != addr {
							c.PublishAll(eve, data)
							//fmt.Println("Cluster Mgr On Event xxxxxxxxxxxxxx 222", pubaddr, addr)
						}
					}
				}
			}
		})
	}

	mgr.Clients[addr] = client

	return nil
}

func (mgr *OBClusterMgr) Stop() {
	for _, c := range mgr.Clients {
		c.Stop()
	}
	mgr.Listener.Close()
	zed.ZLog("OBClusterMgr(%s) Stop()", mgr.Addr)
}

func (mgr *OBClusterMgr) DeleteNode(args *OBAgrs, reply *([]OBRsp)) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

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

	mgr.Listener = listener

	zed.ZLog("ClusterMgr Start on: %s", addr)
	http.Serve(listener, nil)

	return mgr
}

type OBClusterNode struct {
	MgrAddr  string
	NodeAddr string
	Server   *ObserverServer
	Client   *rpc.Client
}

func (node *OBClusterNode) Start() bool {
	go node.Server.Start(node.NodeAddr)
	time.Sleep(time.Second / 2)

	client, err := rpc.DialHTTP("tcp", node.MgrAddr)

	if err != nil {
		zed.ZLog("OBClusterNode Start DialHTTP Error: %v", err)
		return false
	}

	node.Client = client

	args := &OBAgrs{
		Addr:      node.NodeAddr,
		Heartbeat: DEFAULT_HEART_BEAT_TIME - time.Second*3,
	}

	err = client.Call("OBClusterMgr.AddNode", args, nil)
	if err != nil {
		zed.ZLog("OBClusterNode Start OBClusterMgr AddNode Error: %v", err)
		return false
	}

	zed.ZLog("OBClusterNode Start on: %s", node.NodeAddr)
	return true
}

func (node *OBClusterNode) Stop() {
	node.Client.Close()
	node.Server.Stop()
	zed.ZLog("OBClusterNode(%s) Stop()", node.NodeAddr)
}

func NewOBClusterNode(mgraddr string, nodeaddr string, heartbeat time.Duration) *OBClusterNode {
	return &OBClusterNode{
		MgrAddr:  mgraddr,
		NodeAddr: nodeaddr,
		Server:   NewOBServer(fmt.Sprintf("OBClusterNode_%s", nodeaddr)),
	}
}
