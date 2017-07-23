package main

import (
	//"fmt"
	"github.com/ccppluagopy/zed"
	"time"
)

type Delegate struct {
	zed.DefaultTCDelegate
}

func (dele *Delegate) HandleMsg(msg zed.NetMsgDef) {
	time.Sleep(time.Second)
	msg.GetClient().SendMsg(msg)
}

func main() {
	zed.Println("----------- 000")
	dele := Delegate{}
	dele.Init()
	zed.Println("----------- 111")
	dele.SetShowClientData(true)

	client := zed.NewTcpClient(&dele, "127.0.0.1:3333", 1, true, func(c *zed.TcpClient) {
		c.SendMsg(zed.NewNetMsg(333, []byte("hello world")))
	})
	zed.Println("----------- 222")
	client.Connect()
	zed.Println("----------- 333")
	for {
		time.Sleep(time.Hour)
	}
}
