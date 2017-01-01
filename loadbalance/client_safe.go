package loadbalance

/*
import (
	"github.com/ccppluagopy/zed"
	"net/rpc"
	"sync"
)

type LoadbalanceClient struct {
	sync.Mutex
	Addr   string
	Client *rpc.Client
}

func NewLoadbalanceClient(addr string) *LoadbalanceClient {
	client := LoadbalanceClient{
		Addr: addr,
	}

	if err := client.Start(); err != nil {
		zed.ZLog("NewLoadbalanceClient Error: %v", err)
		return nil
	}

	return &client
}

func (client *LoadbalanceClient) Start() error {
	var err error
	client.Client, err = rpc.DialHTTP("tcp", client.Addr)
	return err
}

func (client *LoadbalanceClient) AddServer(serverType string, serverTag string, addr string) error {
	client.Lock()
	defer client.Unlock()

	redo := false

REDO:

	args := &LBArgs{
		ServerType: serverType,
		ServerTag:  serverTag,
		Addr:       addr,
	}

	err := client.Client.Call("LoadbalanceServer.AddServer", args, nil)
	if err != nil {
		if !redo {
			redo = true
			zed.ZLog("LoadbalanceClient AddServer Warn: %v", err)
			err = client.Start()
			if err != nil {
				zed.ZLog("LoadbalanceClient AddServer Error: %v", err)
				return err
			} else {
				goto REDO
			}
		} else {
			zed.ZLog("LoadbalanceClient AddServer Error: %v", err)
			return err
		}

	}

	return nil
}

func (client *LoadbalanceClient) Increament(serverType string, serverTag string, num int) error {
	client.Lock()
	defer client.Unlock()

	redo := false

REDO:

	args := &LBArgs{
		ServerType: serverType,
		ServerTag:  serverTag,
		Num:        num,
	}

	err := client.Client.Call("LoadbalanceServer.Increament", args, nil)
	if err != nil {
		if !redo {
			redo = true
			zed.ZLog("LoadbalanceClient Increament Warn: %v", err)
			err = client.Start()
			if err != nil {
				zed.ZLog("LoadbalanceClient Increament Error: %v", err)
				return err
			} else {
				goto REDO
			}
		} else {
			zed.ZLog("LoadbalanceClient Increament Error: %v", err)
			return err
		}

	}

	return err
}

func (client *LoadbalanceClient) UpdateLoad(serverType string, serverTag string, num int) error {
	client.Lock()
	defer client.Unlock()

	redo := false

REDO:

	args := &LBArgs{
		ServerType: serverType,
		ServerTag:  serverTag,
		Num:        num,
	}

	err := client.Client.Call("LoadbalanceServer.UpdateLoad", args, nil)
	if err != nil {
		if !redo {
			redo = true
			zed.ZLog("LoadbalanceClient UpdateLoad Warn: %v", err)
			err = client.Start()
			if err != nil {
				zed.ZLog("LoadbalanceClient UpdateLoad Error: %v", err)
				return err
			} else {
				goto REDO
			}
		} else {
			zed.ZLog("LoadbalanceClient UpdateLoad Error: %v", err)
			return err
		}

	}

	return err
}

func (client *LoadbalanceClient) DeleteServer(serverType string, serverTag string) error {
	client.Lock()
	defer client.Unlock()

	redo := false

REDO:

	args := &LBArgs{
		ServerType: serverType,
		ServerTag:  serverTag,
	}

	err := client.Client.Call("LoadbalanceServer.DeleteServer", args, nil)
	if err != nil {
		if !redo {
			redo = true
			zed.ZLog("LoadbalanceClient DeleteServer Warn: %v", err)
			err = client.Start()
			if err != nil {
				zed.ZLog("LoadbalanceClient DeleteServer Error: %v", err)
				return err
			} else {
				goto REDO
			}
		} else {
			zed.ZLog("LoadbalanceClient DeleteServer Error: %v", err)
			return err
		}

	}

	return err
}

func (client *LoadbalanceClient) GetMinLoadServerInfoByType(serverType string) (*ServerInfo, error) {
	client.Lock()
	defer client.Unlock()

	redo := false

REDO:

	args := &LBArgs{
		ServerType: serverType,
	}

	reply := LBRsp{}
	err := client.Client.Call("LoadbalanceServer.GetMinLoadServerInfoByType", args, &reply)
	if err != nil {
		if !redo {
			redo = true
			zed.ZLog("LoadbalanceClient GetMinLoadServerInfoByType Warn: %v", err)
			err = client.Start()
			if err != nil {
				zed.ZLog("LoadbalanceClient GetMinLoadServerInfoByType Error: %v", err)
				return nil, err
			} else {
				goto REDO
			}
		} else {
			zed.ZLog("LoadbalanceClient GetMinLoadServerInfoByType Error: %v", err)
			return nil, err
		}

	}

	if len(reply.Infos) == 1 {
		return &(reply.Infos[0]), nil
	}

	return nil, err
}

func (client *LoadbalanceClient) GetServersInfoByType(serverType string) ([]ServerInfo, error) {
	client.Lock()
	defer client.Unlock()

	redo := false

REDO:

	args := &LBArgs{
		ServerType: serverType,
	}

	reply := LBRsp{}
	err := client.Client.Call("LoadbalanceServer.GetServersInfoByType", args, &reply)
	if err != nil {
		if !redo {
			redo = true
			zed.ZLog("LoadbalanceClient GetServersInfoByType Warn: %v", err)
			err = client.Start()
			if err != nil {
				zed.ZLog("LoadbalanceClient GetServersInfoByType Error: %v", err)
				return nil, err
			} else {
				goto REDO
			}
		} else {
			zed.ZLog("LoadbalanceClient GetServersInfoByType Error: %v", err)
			return nil, err
		}

	}

	return reply.Infos, err
}

func (client *LoadbalanceClient) Stop() {
	client.Lock()
	defer client.Unlock()

	client.Client.Close()
}
*/
/*
func main() {
	time.Sleep(time.Second * 3)
	client, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	if err != nil {
		fmt.Println("dialing:", err)
	}

	args := &LBArgs{
		ServerType: "test",
		ServerTag:  "tag",
		Addr:       "127.0.0.1:aaaa",
		Num:        1,
	}

	reply := []commrpc.LBRsp{}

	for i := 0; i < 10000; i++ {
		err = client.Call("LoadbalanceServer.Increament", args, &reply)
		if err != nil {
			fmt.Println("arith error:", err)
		}
	}
	fmt.Println("time used: ", time.Since(t1).Seconds())

	err = client.Call("Online.GetServerAddr", args, &reply)
	if err != nil {
		fmt.Println("arith error:", err)
	}
	fmt.Println(222, reply)

	for i := 0; i < len(reply); i++ {
		fmt.Println("i,N:", i, reply[i].Info.Addr, reply[i].Info.Num)
	}

}
*/
