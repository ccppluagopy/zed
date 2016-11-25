package zed

import (
	//"encoding/binary"
	"fmt"
	//"io"

	"net"
	//"sync"
	//"time"
)

type AsyncMsg struct {
	msg *NetMsg
	cb  func()
}

func (client *TcpClient) Info() string {
	return fmt.Sprintf("Client(ID: %v-Addr: %s)", client.ID, client.Addr)
}

func (client *TcpClient) AddCloseCB(key interface{}, cb ClientCloseCB) {
	client.Lock()
	defer client.Unlock()
	if client.running {
		client.closeCB[key] = cb
	}
}

func (client *TcpClient) GetConn() *net.TCPConn {
	return client.conn
}

func (client *TcpClient) RemoveCloseCB(key interface{}) {
	client.Lock()
	defer client.Unlock()
	if client.running {
		delete(client.closeCB, key)
	}
}

func (client *TcpClient) IsRunning() bool {
	client.Lock()
	defer client.Unlock()
	running := client.running
	return running
}

func (client *TcpClient) Stop() {
	LogStackInfo()
	NewCoroutine(func() {
		client.Lock()
		defer client.Unlock()

		if client.running {

			if client.parent.showClientData {
				ZLog("[Stop_0] %s %v", client.Info(), client.running)
			}
			client.running = false

			client.conn.Close()
			//client.conn.SetLinger(0)

			if client.parent.showClientData {
				ZLog("[Stop_1] %s chSend: %v %v", client.Info(), client.chSend, client.running)
			}

			if client.chSend != nil {
				close(client.chSend)
				client.chSend = nil
			}

			if client.parent.showClientData {
				ZLog("[Stop_2] %s %v", client.Info(), client.running)
			}

			if len(client.closeCB) > 0 {
				NewCoroutine(func() {
					for _, cb := range client.closeCB {
						cb(client)
					}

					if client.parent.showClientData {
						ZLog("[Stop_5] %s %v", client.Info(), client.running)
					}
				})
			}

			if client.parent.showClientData {
				ZLog("[Stop_3] %s %v", client.Info(), client.running)
			}

			for key, _ := range client.closeCB {
				delete(client.closeCB, key)
			}

			if client.parent.showClientData {
				ZLog("[Stop_4] %s %v", client.Info(), client.running)
			}

		}
	})

}

func (client *TcpClient) writer() {
	parent := client.parent
	if client.chSend != nil {
		for {
			if asyncMsg, ok := <-client.chSend; ok {
				if !func() bool {
					client.Lock()
					defer client.Unlock()
					defer func() {
						if err := PanicHandle(true); err != nil {
							client.Stop()
							return
						}
					}()
					parent.SendMsg(client, asyncMsg.msg)
					if asyncMsg.cb != nil {
						NewCoroutine(func() {
							asyncMsg.cb()
						})
					}
					return true
				}() {
					return
				}
			} else {
				break
			}
		}
	}
}

func (client *TcpClient) SendMsg(msg *NetMsg) {
	ZLog("[Send_0] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
	client.Lock()
	defer client.Unlock()
	defer func() {
		if err := PanicHandle(true); err != nil {
			client.conn.Close()
			return
		}
	}()

	ZLog("[Send_1] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
	client.parent.SendMsg(client, msg)

	//client.SendMsgAsync(msg)
}

func (client *TcpClient) SendMsgAsync(msg *NetMsg, argv ...interface{}) {
	ZLog("[SendAsync_00] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)

	client.Lock()
	defer client.Unlock()
	if client.running {
		ZLog("[SendAsync_01] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
		asyncmsg := &AsyncMsg{
			msg: msg,
			cb:  nil,
		}

		if len(argv) > 0 {
			if cb, ok := (argv[0]).(func()); ok {
				asyncmsg.cb = cb
			}
		}
		if client.chSend != nil {
			Println("aaaaaaa", client.Info(), msg.Cmd, msg.Len, client.chSend)
			client.chSend <- asyncmsg
			Println("bbbbbbb", client.Info(), msg.Cmd, msg.Len, client.chSend)
		}
	}
	ZLog("[SendAsync_1] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
}

func (client *TcpClient) reader() {
	var (
		/*head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error*/
		msg    *NetMsg
		parent = client.parent
	)

	for {

		msg = parent.RecvMsg(client)
		if msg == nil {
			goto Exit
		}

		parent.HandleMsg(msg)
	}

Exit:
	client.Stop()

}

func (client *TcpClient) StartReader() {
	NewCoroutine(func() {
		client.reader()
	})
}

func (client *TcpClient) StartWriter() {
	NewCoroutine(func() {
		client.writer()
	})
}

func (client *TcpClient) start() bool {
	if err := client.conn.SetKeepAlive(true); err != nil {
		if client.parent.showClientData {
			ZLog("%s SetKeepAlive Err: %v.", client.Info())
		}
		return false
	}

	if err := client.conn.SetKeepAlivePeriod(client.parent.aliveTime); err != nil {
		if client.parent.showClientData {
			ZLog("%s SetKeepAlivePeriod Err: %v.", client.Info(), err)
		}
		return false
	}

	if err := (*client.conn).SetReadBuffer(client.parent.recvBufLen); err != nil {
		if client.parent.showClientData {
			ZLog("%s SetReadBuffer Err: %v.", client.Info(), err)
		}
		return false
	}
	if err := (*client.conn).SetWriteBuffer(client.parent.sendBufLen); err != nil {
		if client.parent.showClientData {
			ZLog("%s SetWriteBuffer Err: %v.", client.Info(), err)
		}
		return false
	}

	/*NewCoroutine(func() {
		client.writer()
	})*/
	client.StartWriter()
	client.StartReader()

	if client.parent.showClientData {
		ZLog("New Client Start %s", client.Info())
	}

	return true
}

func newTcpClient(parent *TcpServer, conn *net.TCPConn) *TcpClient {
	client := &TcpClient{
		conn:    conn,
		parent:  parent,
		ID:      NullID,
		Idx:     parent.ClientNum,
		Addr:    conn.RemoteAddr().String(),
		closeCB: make(map[interface{}]ClientCloseCB),
		chSend:  make(chan *AsyncMsg, 100),

		//Data:    nil,
		Valid:   false,
		running: true,
	}

	return client
}
