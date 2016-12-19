package loadbalance

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"net"
	"net/http"
	"net/rpc"
	"sync"
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

type LBRsp struct {
	Info ServerInfo
}

type LoadbalanceServer struct {
	Addr    string
	Servers map[string]map[string]*ServerInfo
}

/*func (server *LoadbalanceServer) Init() {
	server.Servers = make(map[string]map[string]*ServerInfo)
}*/

func (server *LoadbalanceServer) AddServer(args *LBArgs, reply *([]LBRsp)) error {
	mtx.Lock()
	defer mtx.Unlock()

	serverType, serverTag, addr, num := args.ServerType, args.ServerTag, args.Addr, args.Num

	fmt.Println("LoadbalanceServer.AddServer 000: ", serverType, serverTag, addr)

	servers, ok := server.Servers[serverType]
	if !ok {
		fmt.Println("LoadbalanceServer.AddServer 111: ", serverType, serverTag, addr)
		servers = make(map[string]*ServerInfo)
		server.Servers[serverType] = servers
	}

	if _, ok2 := servers[serverTag]; !ok2 {
		fmt.Println("LoadbalanceServer.AddServer 222: ", servers)
		servers[serverTag] = &ServerInfo{
			Addr: addr,
			Num:  num,
		}
		fmt.Println("LoadbalanceServer.AddServer 333: ", serverType, serverTag, addr)
	}

	fmt.Println("LoadbalanceServer.AddServer 444: ", serverType, serverTag, addr)

	return nil
}

func (server *LoadbalanceServer) DeleteServer(args *LBArgs, reply *([]LBRsp)) error {
	mtx.Lock()
	defer mtx.Unlock()

	fmt.Println("LoadbalanceServer.DeleteServer 000")

	serverType, serverTag := args.ServerType, args.ServerTag

	if servers, ok := server.Servers[serverType]; ok {
		delete(servers, serverTag)
		fmt.Println("LoadbalanceServer.DeleteServer 222: ", serverType, serverTag)
	}

	return nil
}

func (server *LoadbalanceServer) Increament(args *LBArgs, reply *([]LBRsp)) error {
	mtx.Lock()
	defer mtx.Unlock()

	fmt.Println("LoadbalanceServer.Increament 000")

	serverType, serverTag, num := args.ServerType, args.ServerTag, args.Num

	if servers, ok := server.Servers[serverType]; ok {
		server, ok2 := servers[serverTag]
		if ok2 {
			server.Num += num
			fmt.Println("LoadbalanceServer.Increament 333: ", serverType, serverTag, num)
		}
	}

	return nil
}

func (server *LoadbalanceServer) UpdateLoad(args *LBArgs, reply *([]LBRsp)) error {
	mtx.Lock()
	defer mtx.Unlock()

	fmt.Println("LoadbalanceServer.UpdateLoad 000")

	serverType, serverTag, num := args.ServerType, args.ServerTag, args.Num

	if servers, ok := server.Servers[serverType]; ok {
		server, ok2 := servers[serverTag]
		if ok2 {
			server.Num = num
			fmt.Println("LoadbalanceServer.UpdateLoad 333: ", serverType, serverTag, num)
		}
	}

	return nil
}

func (server *LoadbalanceServer) GetServerAddr(args *LBArgs, reply *([]LBRsp)) error {
	mtx.Lock()
	defer mtx.Unlock()

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

func NewLoadbalanceServer(addr string) *LoadbalanceServer {
	server := &LoadbalanceServer{
		Addr:    addr,
		Servers: make(map[string]map[string]*ServerInfo),
	}

	rpc.Register(server)
	rpc.HandleHTTP()

	listener, e := net.Listen("tcp", server.Addr)
	if e != nil {
		zed.ZLog("NewLoadbalanceServer Listen error: %v", e)
		return nil
	}

	zed.NewCoroutine(func() {
		http.Serve(listener, nil)
	})

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
