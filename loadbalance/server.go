package loadbalance

import (
	"fmt"
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
}

type LBArgs struct {
	ServerType string
	ServerTag  string
	Addr       string
	Num        int
}

type LBError struct {
	description string
}

func (lbError *LBError) Error() string {
	return lbError.description
}

type LBRsp struct {
	Info ServerInfo
}

type LoadbalanceServer struct {
	sync.Mutex
	Addr     string
	ticker   *time.Ticker
	listener net.Listener
	Servers  map[string]map[string]*ServerInfo
}

/*func (server *LoadbalanceServer) Init() {
	server.Servers = make(map[string]map[string]*ServerInfo)
}*/

func (server *LoadbalanceServer) AddServer(args *LBArgs, reply *([]LBRsp)) error {
	server.Lock()
	defer server.Unlock()

	serverType, serverTag, addr, num := args.ServerType, args.ServerTag, args.Addr, args.Num

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
		}
	} else {
		zed.ZLog("LoadbalanceServer Addserver Error: serverTag %s has been exist", serverTag)
		return &LBError{
			description: "LoadbalanceServer Addserver Error: serverTag %s has been exist",
		}
	}

	return nil
}

func (server *LoadbalanceServer) DeleteServer(args *LBArgs, reply *([]LBRsp)) error {
	server.Lock()
	defer server.Unlock()

	serverType, serverTag := args.ServerType, args.ServerTag

	if servers, ok := server.Servers[serverType]; ok {
		delete(servers, serverTag)
	}

	return nil
}

func (server *LoadbalanceServer) Increament(args *LBArgs, reply *([]LBRsp)) error {
	server.Lock()
	defer server.Unlock()

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

func (server *LoadbalanceServer) UpdateLoad(args *LBArgs, reply *([]LBRsp)) error {
	server.Lock()
	defer server.Unlock()

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

func (server *LoadbalanceServer) GetServerAddr(args *LBArgs, reply *([]LBRsp)) error {
	server.Lock()
	defer server.Unlock()

	//fmt.Println("LoadbalanceServer.GetServerAddr 000")

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
		*reply = []LBRsp{
			LBRsp{Info: *info},
		}
		//fmt.Println("LoadbalanceServer.GetServerAddr 444: ", args.ServerType, info.Addr, info.Num)
	}

	return nil
}

func (server *LoadbalanceServer) Stop() {
	server.listener.Close()
	server.ticker.Stop()
	zed.ZLog("NewLoadbalanceServer Stop()")

}

func (server *LoadbalanceServer) startHeartbeat() {
	go func() {
		for {
			select {
			case _, ok := <-server.ticker.C:
				if ok {
					func() {
						defer zed.PanicHandle(true)
						for _, servers := range server.Servers {
							for tag, info := range servers {
								if !zed.Ping(info.Addr) {
									server.Lock()
									delete(servers, tag)
									server.Unlock()
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

func NewLoadbalanceServer(addr string) *LoadbalanceServer {
	server := &LoadbalanceServer{
		Addr:    addr,
		Servers: make(map[string]map[string]*ServerInfo),
		ticker:  time.NewTicker(time.Second * 10),
		//ticker:  time.NewTicker(time.Minute),
	}

	rpc.Register(server)
	rpc.HandleHTTP()

	listener, e := net.Listen("tcp", addr)
	if e != nil {
		zed.ZLog("NewLoadbalanceServer Listen error: %v", e)
		return nil
	}

	server.startHeartbeat()
	server.listener = listener
	zed.NewCoroutine(func() {
		http.Serve(server.listener, nil)
	})

	zed.ZLog("NewLoadbalanceServer Start on: %s", addr)

	return server
}

func GetServerAddr(server *LoadbalanceServer, serverType string) *ServerInfo {
	args := &LBArgs{
		ServerType: serverType,
	}
	reply := []LBRsp{}
	err := server.GetServerAddr(args, &reply)
	if err != nil {
		zed.ZLog("loadbalance GetServerAddr Error:", err)
	}

	if len(reply) == 1 {
		//zed.ZLog("LoadbalanceClient GetServerAddr:", reply[0].Info.Addr, reply[0].Info.Num)
		return &(reply[0].Info)
	}

	return nil
}
