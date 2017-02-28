package loadbalance

import (
	//"fmt"
	"github.com/ccppluagopy/zed"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

var (
	mtx = sync.Mutex{}
)

type ServerInfo struct {
	Addr string
	Num  int
	Attr []byte
}

type LBArgs struct {
	ServerType string
	ServerTag  string
	Addr       string
	Num        int
	Attr       []byte
}

type LBError struct {
	description string
}

func (lbError *LBError) Error() string {
	return lbError.description
}

type LBRsp struct {
	Infos []ServerInfo
}

type LoadbalanceServer struct {
	mutex    sync.Mutex
	Addr     string
	ticker   *time.Ticker
	listener net.Listener
	Servers  map[string]map[string]*ServerInfo
}

/*func (server *LoadbalanceServer) Init() {
	server.Servers = make(map[string]map[string]*ServerInfo)
}*/

func (server *LoadbalanceServer) AddServer(args *LBArgs, reply *LBRsp) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	serverType, serverTag, addr, num, attr := args.ServerType, args.ServerTag, args.Addr, args.Num, args.Attr

	servers, ok := server.Servers[serverType]
	if !ok {
		//fmt.Println("LoadbalanceServer.AddServer 111: ", serverType, serverTag, addr)
		servers = make(map[string]*ServerInfo)
		server.Servers[serverType] = servers
	}

	if _, ok2 := servers[serverTag]; !ok2 {
		//fmt.Println("LoadbalanceServer.AddServer 222: ", servers)
		servers[serverTag] = &ServerInfo{
			Addr: addr,
			Num:  num,
			Attr: attr,
		}
	} else {
		zed.ZLog("LoadbalanceServer Addserver Error: serverTag %s has been exist", serverTag)
		return &LBError{
			description: "LoadbalanceServer Addserver Error: serverTag %s has been exist",
		}
	}

	return nil
}

func (server *LoadbalanceServer) DeleteServer(args *LBArgs, reply *LBRsp) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	serverType, serverTag := args.ServerType, args.ServerTag

	if servers, ok := server.Servers[serverType]; ok {
		//deleteServer(server, serverType, serverTag)
		server.mutex.Lock()
		delete(servers, serverTag)
		server.mutex.Unlock()
	}

	return nil
}

func (server *LoadbalanceServer) Increament(args *LBArgs, reply *LBRsp) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	serverType, serverTag, num := args.ServerType, args.ServerTag, args.Num

	if servers, ok := server.Servers[serverType]; ok {
		server, ok2 := servers[serverTag]
		if ok2 {
			server.Num += num
			return nil
		}
	}

	//zed.ZLog("LoadbalanceServer Increament Error: serverTag %s does not exist", serverTag)
	return &LBError{
		description: "LoadbalanceServer Increament Error: serverTag %s does not exist",
	}
}

func (server *LoadbalanceServer) UpdateLoad(args *LBArgs, reply *LBRsp) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	serverType, serverTag, num := args.ServerType, args.ServerTag, args.Num

	if servers, ok := server.Servers[serverType]; ok {
		server, ok2 := servers[serverTag]
		if ok2 {
			server.Num = num
			return nil
		}
	}

	//zed.ZLog("LoadbalanceServer UpdateLoad Error: serverTag %s does not exist", serverTag)
	return &LBError{
		description: "LoadbalanceServer UpdateLoad Error: serverTag %s does not exist",
	}
}

func (server *LoadbalanceServer) GetMinLoadServerInfoByType(args *LBArgs, reply *LBRsp) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	//fmt.Println("LoadbalanceServer.GetMinLoadServerInfoByType 000")

	var info *ServerInfo = nil

	if servers, ok := server.Servers[args.ServerType]; ok {
		min := 0xFFFFFFF
		for _, v := range servers {
			if v.Num < min {
				min = v.Num
				info = v
			}
		}
		//*reply = append(*reply, addr)
		*reply = LBRsp{
			Infos: []ServerInfo{*info},
		}
		//fmt.Println("LoadbalanceServer.GetMinLoadServerInfoByType 444: ", args.ServerType, info.Addr, info.Num)
	}

	return nil
}

func (server *LoadbalanceServer) GetServersInfoByType(args *LBArgs, reply *LBRsp) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	//fmt.Println("LoadbalanceServer.GetServersInfoByType 000")

	tmpreply := *reply
	if servers, ok := server.Servers[args.ServerType]; ok {
		for _, info := range servers {
			tmpreply.Infos = append(tmpreply.Infos, *info)
		}

		//fmt.Println("LoadbalanceServer.GetServersInfoByType 444: ", args.ServerType, info.Addr, info.Num)
	}

	return nil
}

func deleteServer(server *LoadbalanceServer, stype string, tag string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	servers, ok := server.Servers[stype]
	if ok {
		delete(servers, tag)
	}
}

func (server *LoadbalanceServer) startHeartbeat() {
	go func() {
		for {
			select {
			case _, ok := <-server.ticker.C:
				if ok {
					func() {
						defer zed.HandlePanic(true)
						for _, servers := range server.Servers {
							for tag, info := range servers {
								if !zed.Ping(info.Addr) {
									//deleteServer(server, stype, tag)
									server.mutex.Lock()
									delete(servers, tag)
									server.mutex.Unlock()
								}
							}
						}
					}()
				} else {
					return
				}
			}
		}
	}()
}

func NewLoadbalanceServer(addr string, pingTime time.Duration) *LoadbalanceServer {
	server := &LoadbalanceServer{
		Addr:    addr,
		Servers: make(map[string]map[string]*ServerInfo),
		ticker:  time.NewTicker(pingTime),
		//ticker:  time.NewTicker(time.Minute),
	}

	return server
}

func StartLBServer(server *LoadbalanceServer) {
	rpc.Register(server)
	rpc.HandleHTTP()

	listener, e := net.Listen("tcp", server.Addr)
	if e != nil {
		zed.ZLog("LoadbalanceServer Listen error: %v", e)
		return
	}

	server.startHeartbeat()
	server.listener = listener

	zed.ZLog("LoadbalanceServer Start on: %s", server.Addr)
	http.Serve(server.listener, nil)
}

func StopLBServer(server *LoadbalanceServer) {
	server.listener.Close()
	server.ticker.Stop()
	zed.ZLog("LoadbalanceServer Stop()")
}

func GetMinLoadServerInfoByType(server *LoadbalanceServer, serverType string) *ServerInfo {
	args := &LBArgs{
		ServerType: serverType,
	}
	reply := LBRsp{}
	err := server.GetMinLoadServerInfoByType(args, &reply)
	if err != nil {
		zed.ZLog("loadbalance GetMinLoadServerInfoByType Error:", err)
	}

	if len(reply.Infos) == 1 {
		//zed.ZLog("LoadbalanceClient GetMinLoadServerInfoByType:", reply[0].Info.Addr, reply[0].Info.Num)
		return &(reply.Infos[0])
	}

	return nil
}
