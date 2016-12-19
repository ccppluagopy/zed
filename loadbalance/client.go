package loadbalance

import (
	//"fmt"
	//"net"
	//"net/http"
	"github.com/ccppluagopy/zed"
	//"net"
	//"net/http"
	"net/rpc"
	//"sync"
	//"time"
)

type LoadbalanceClient struct {
	Client *rpc.Client
}

func NewLoadbalanceClient(addr string) *LoadbalanceClient {
	var err error
	client := LoadbalanceClient{}
	client.Client, err = rpc.DialHTTP("tcp", addr)

	if err != nil {
		zed.ZLog("NewLoadbalanceClient Error: %v", err)
	}

	return &client
}

func (client *LoadbalanceClient) AddServer(serverType string, serverTag string, addr string) {
	args := &LBArgs{
		ServerType: serverType,
		ServerTag:  serverTag,
		Addr:       addr,
	}

	err := client.Client.Call("LoadbalanceServer.AddServer", args, nil)
	if err != nil {
		zed.ZLog("LoadbalanceClient AddServer Error: %v", err)
	}
}

func (client *LoadbalanceClient) Increament(serverType string, serverTag string, num int) {
	args := &LBArgs{
		ServerType: serverType,
		ServerTag:  serverTag,
		Num:        num,
	}

	err := client.Client.Call("LoadbalanceServer.Increament", args, nil)
	if err != nil {
		zed.ZLog("LoadbalanceClient Increament Error: %v", err)
	}
}

func (client *LoadbalanceClient) UpdateLoad(serverType string, serverTag string, num int) {
	args := &LBArgs{
		ServerType: serverType,
		ServerTag:  serverTag,
		Num:        num,
	}

	err := client.Client.Call("LoadbalanceServer.UpdateLoad", args, nil)
	if err != nil {
		zed.ZLog("LoadbalanceClient UpdateLoad Error: %v", err)
	}
}

func (client *LoadbalanceClient) DeleteServer(serverType string, serverTag string) {
	args := &LBArgs{
		ServerType: serverType,
		ServerTag:  serverTag,
	}

	err := client.Client.Call("LoadbalanceServer.DeleteServer", args, nil)
	if err != nil {
		zed.ZLog("LoadbalanceClient DeleteServer Error: %v", err)
	}
}

func (client *LoadbalanceClient) GetServerAddr(serverType string) *ServerInfo {
	args := &LBArgs{
		ServerType: serverType,
	}

	reply := []LBRsp{}
	err := client.Client.Call("LoadbalanceServer.GetServerAddr", args, &reply)
	if err != nil {
		zed.ZLog("LoadbalanceClient GetServerAddr Error:", err)
	}

	if len(reply) == 1 {
		//zed.ZLog("LoadbalanceClient GetServerAddr:", reply[0].Info.Addr, reply[0].Info.Num)
		return &(reply[0].Info)
	}

	return nil
}

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
