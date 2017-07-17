package zed

import (
	//"encoding/binary"
	"fmt"
	//"io"
	"net"
	"runtime"
	"time"
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
	//LogStackInfo()
	//NewCoroutine(func() {
	time.AfterFunc(1, func() {
		defer HandlePanic(true)
		client.Lock()
		defer client.Unlock()
		showClientData := client.parent.ShowClientData()

		if client.running {

			if showClientData {
				ZLog("[Stop_0] %s %v", client.Info(), client.running)
			}
			client.running = false

			client.conn.Close()
			//client.conn.SetLinger(0)

			if showClientData {
				ZLog("[Stop_1] %s chSend: %v %v", client.Info(), client.chSend, client.running)
			}

			if client.chSend != nil {
				close(client.chSend)
				for _ = range client.chSend {

				}
				client.chSend = nil
			}

			if showClientData {
				ZLog("[Stop_2] %s %v", client.Info(), client.running)
			}

			if len(client.closeCB) > 0 {
				NewCoroutine(func() {
					for _, cb := range client.closeCB {
						cb(client)
					}

					if showClientData {
						ZLog("[Stop_5] %s %v", client.Info(), client.running)
					}
				})
			}

			if showClientData {
				ZLog("[Stop_3] %s %v", client.Info(), client.running)
			}

			for key, _ := range client.closeCB {
				delete(client.closeCB, key)
			}

			if showClientData {
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
						if err := HandlePanic(true); err != nil {
							client.Stop()
							return
						}
					}()
					time.AfterFunc(1, func() {
						parent.SendMsg(client, asyncMsg.msg)
						if asyncMsg.cb != nil {
							defer HandlePanic(true)
							asyncMsg.cb()
						}
					})
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
	if client.parent.ShowClientData() {
		ZLog("[Send_0] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
	}
	client.Lock()
	defer client.Unlock()
	defer func() {
		if err := HandlePanic(true); err != nil {
			client.conn.Close()
			return
		}
	}()
	if client.parent.ShowClientData() {
		ZLog("[Send_1] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
	}
	client.parent.SendMsg(client, msg)

	//client.SendMsgAsync(msg)
}

func (client *TcpClient) SendMsgAsync(msg *NetMsg, argv ...interface{}) bool {
	if client.parent.ShowClientData() {
		ZLog("[SendAsync_00] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
	}

	client.Lock()
	defer client.Unlock()
	if client.running {
		if client.parent.ShowClientData() {
			ZLog("[SendAsync_01] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
		}
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
			//Println("aaaaaaa", client.Info(), msg.Cmd, msg.Len, client.chSend)
			select {
			case client.chSend <- asyncmsg:
				break
			case <-time.After(time.Second*2):
				return false
			}
			//Println("bbbbbbb", client.Info(), msg.Cmd, msg.Len, client.chSend)
		}
	}
	if client.parent.ShowClientData() {
		ZLog("[SendAsync_02] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
	}
	return true
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

func (client *TcpClient) Start() bool {
	return client.start()
}

func (client *TcpClient) start() bool {
	showClientData := client.parent.ShowClientData()
	if err := client.conn.SetKeepAlive(true); err != nil {
		if showClientData {
			ZLog("%s SetKeepAlive Err: %v.", client.Info())
		}
		return false
	}

	if err := client.conn.SetKeepAlivePeriod(client.parent.AliveTime()); err != nil {
		if showClientData {
			ZLog("%s SetKeepAlivePeriod Err: %v.", client.Info(), err)
		}
		return false
	}

	if err := (*client.conn).SetReadBuffer(client.parent.RecvBufLen()); err != nil {
		if showClientData {
			ZLog("%s SetReadBuffer Err: %v.", client.Info(), err)
		}
		return false
	}
	if err := (*client.conn).SetWriteBuffer(client.parent.SendBufLen()); err != nil {
		if showClientData {
			ZLog("%s SetWriteBuffer Err: %v.", client.Info(), err)
		}
		return false
	}
	if err := (*client.conn).SetNoDelay(client.parent.NoDelay()); err != nil {
		if showClientData {
			ZLog("%s SetNoDelay Err: %v.", client.Info(), err)
		}
		return false
	}

	/*NewCoroutine(func() {
		client.writer()
	})*/
	client.StartWriter()
	client.StartReader()

	if showClientData {
		ZLog("New Client Start %s", client.Info())
	}

	return true
}

func newTcpClient(parent ZTcpClientDelegate, conn *net.TCPConn, idx int) *TcpClient {
	client := &TcpClient{
		conn:    conn,
		parent:  parent,
		ID:      NullID,
		Idx:     idx,
		Addr:    conn.RemoteAddr().String(),
		closeCB: make(map[interface{}]ClientCloseCB),
		chSend:  make(chan *AsyncMsg, 100),

		//Data:    nil,
		Valid:   false,
		running: true,
	}
	
	if runtime.GOOS != "windows" && conn != nil {
		file, _ := conn.File()
		client.Idx = int(file.Fd())
	}

	return client
}

func NewTcpClient(dele ZTcpClientDelegate, serveraddr string, idx int) *TcpClient {
	tcpAddr, err := net.ResolveTCPAddr("tcp", serveraddr)
	if err != nil {
		ZLog("NewTcpClient ResolveTCPAddr Failed, err: %s", err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	//conn, err := net.DialTimeout("tcp", serveraddr, 3000000000)

	if err != nil {
		ZLog("NewTcpClient DialTCP(%s) Failed, err: %s", serveraddr, err)
		return nil
	}

	//dele.Init()

	client := newTcpClient(dele, conn, idx)

	if client != nil && client.start() {
		return client
	}

	return nil
}

func Ping(addr string) bool {
	client := NewTcpClient(&DefaultTCDelegate{}, addr, 0)
	if client != nil {
		client.Stop()
		return true
	}
	return false
}
