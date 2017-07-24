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
	msg NetMsgDef
	cb  func()
}

func (client *TcpClient) Info() string {
	return fmt.Sprintf("Client(Idx: %v, Addr: %s <-> %s)", client.Idx, client.conn.LocalAddr().String(), client.Addr)
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
				ZLog("[Stop] %s %v", client.Info(), client.running)
			}
			client.running = false

			client.conn.Close()
			//client.conn.SetLinger(0)

			if client.chSend != nil {
				close(client.chSend)
				for _ = range client.chSend {

				}
				client.chSend = nil
			}

			if len(client.closeCB) > 0 {
				//NewCoroutine(func() {
				for _, cb := range client.closeCB {
					time.AfterFunc(1, func() {
						cb(client)
					})
				}
			}

			/*for key, _ := range client.closeCB {
				delete(client.closeCB, key)
			}*/
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

func (client *TcpClient) SendMsg(msg NetMsgDef) {
	client.Lock()
	defer client.Unlock()
	defer func() {
		if err := HandlePanic(true); err != nil {
			client.conn.Close()
			return
		}
	}()

	client.parent.SendMsg(client, msg)
	/*if client.parent.ShowClientData() {
		ZLog("[Send] %s Cmd: %d Len: %d", client.Info(), msg.Cmd, msg.Len)
	}*/
	//client.SendMsgAsync(msg)
}

func (client *TcpClient) SendMsgAsync(msg NetMsgDef, argv ...interface{}) bool {
	client.Lock()
	defer client.Unlock()
	if client.running {

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
				if client.parent.ShowClientData() {
					ZLog("[SendAsync][Info] %s Cmd: %d Len: %d Success", client.Info(), msg.GetCmd(), msg.GetMsgLen())
				}
				break
			case <-time.After(time.Second * 2):
				if client.parent.ShowClientData() {
					ZLog("[SendAsync][Error] %s Cmd: %d Len: %d Timeout", client.Info(), msg.GetCmd(), msg.GetMsgLen())
				}
				return false
			}
			//Println("bbbbbbb", client.Info(), msg.Cmd, msg.Len, client.chSend)
		}
	}

	return true
}

func (client *TcpClient) reader() {
	var (
		/*head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error*/
		parent = client.parent
	)

	for {
		msgdef := parent.RecvMsg(client)
		if msgdef == nil {
			goto Exit
		}
		parent.HandleMsg(msgdef)
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

func (client *TcpClient) Connect() {
	client.Lock()
	defer client.Unlock()
	tcpAddr, err := net.ResolveTCPAddr("tcp", client.Addr)
	if err != nil {
		ZLog("TcpClient Connect ResolveTCPAddr Failed, err: %s, Addr: %s", err, client.Addr)
		goto ErrExit
	}
	client.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		ZLog("TcpClient Connect DialTCP(%s) Failed, err: %s", client.Addr, err)
		goto ErrExit
	}
	if !client.start() {
		goto ErrExit
	}

	client.running = true
	if client.EnableReconnect {
		Async(func() {
			HandlePanic(true)
			client.AddCloseCB("__auto__reconnect__", func(c *TcpClient) {
				c.Connect()
			})
		})
	}
	if client.onConnected != nil {
		Async(func() {
			HandlePanic(true)
			client.onConnected(client)
		})
	}
	return
ErrExit:
	if client.EnableReconnect {
		Async(func() {
			time.Sleep(time.Second)
			client.Connect()
		})
	}
}

func newTcpClient(parent ZTcpClientDelegate, conn *net.TCPConn, idx int) *TcpClient {
	client := &TcpClient{
		conn:   conn,
		parent: parent,
		//ID:     NullID,
		Idx:    idx,
		//Addr:    conn.RemoteAddr().String(),
		closeCB: make(map[interface{}]ClientCloseCB),
		chSend:  make(chan *AsyncMsg, 100),

		//Data:    nil,
		Valid:           false,
		running:         true,
		EnableReconnect: false,
		onConnected:     nil,
	}

	if conn != nil {
		client.Addr = conn.RemoteAddr().String()
	}
	if runtime.GOOS != "windows" && conn != nil {
		file, _ := conn.File()
		client.Idx = int(file.Fd())
	}

	return client
}

func NewTcpClient(dele ZTcpClientDelegate, serveraddr string, idx int, reconn bool, onconnected func(*TcpClient)) *TcpClient {
	dele.Init()
	client := newTcpClient(dele, nil, idx)
	client.Addr = serveraddr
	client.EnableReconnect = reconn
	client.onConnected = onconnected

	return client
}

func Ping(addr string) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return false
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	defer conn.Close()
	if err != nil {
		return false
	}

	return true
}
