package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	lb "github.com/ccppluagopy/zed/loadbalance"
	"time"
)

func client(n int, stype string, stag string, addr string, addr2 string) {
	go func() {
		time.Sleep(time.Second)

		server := zed.NewTcpServer(addr2)
		go server.Start(addr2)

		client := lb.NewLoadbalanceClient(addr)
		client.AddServer(stype, stag, addr2)
		client.UpdateLoad(stype, stag, n)
		time.Sleep(time.Second * 2)
		ret, err := client.GetMinLoadServerInfoByType(stype)
		fmt.Println(stype, stag, addr, " -- over ---", ret.Addr, ret.Num, err)
	}()
}

func timeout(addr string) {
	client := lb.NewLoadbalanceClient(addr)
	for i := 1; i < 1000; i++ {
		fmt.Println(" -- xxx 000 ---", 5*i)
		time.Sleep(time.Minute * 5 * time.Duration(i))
		fmt.Println(" -- xxx 111 ---", 5*i)
		//client.GetMinLoadServerInfoByType("test server")
		ret, err := client.GetMinLoadServerInfoByType("test server")
		fmt.Println(" -- xxx 222 ---", 5*i, "minutes:", ret.Addr, ret.Num, err)
	}
}
func main() {
	addr := "127.0.0.1:8888"
	server := lb.NewLoadbalanceServer(addr, time.Second*10)

	addr1 := "127.0.0.1:10001"
	addr2 := "127.0.0.1:10002"
	addr3 := "127.0.0.1:10003"
	client(1, "test server", "test server 1", addr, addr1)
	client(2, "test server", "test server 2", addr, addr2)
	client(3, "test server", "test server 3", addr, addr3)
	go timeout(addr)
	lb.StartLBServer(server)
	time.Sleep(time.Second * 3)
	fmt.Println("Over!")
}
