package main

import (
	"github.com/ccppluagopy/zed"
)

func main() {
	server := zed.NewTcpServer(conf.ServerName)
	server.SetDelegate(&ServerDelegate{})
	server.SetDataInSupervisor(module.DataRecvSupervisor)
	server.SetDataOutSupervisor(module.DataSendSupervisor)
	server.SetShowClientData(true)
	clients := make(map[*zed.TcpClient]*zed.TcpClient)
	server.SetNewConnCB(func(c *zed.TcpClient) {
		//clients[c] = c
		c.Stop()
	})
	server.SetServerStopCB(func() {
		/*for c, _ := range clients {
			c.Stop()
		}*/
	})

}
