package autoreconnclient

import(
	"encoding/binary"
	"github.com/ccppluagopy/zed"
	"io"
	"net"
	"sync"
	"time"
)

const(
	MAX_PACK_LEN = 1024*1024
	SEND_BLOCK_TIME = time.Second * 5
)

type AutoReconnectClient struct{
	sync.Mutex
	ID string
	Conn interface{}
	ServerAddr string
	Running bool
	ConnTimes int
	chSend chan *AsyncMsg
	handlerMap map[uint32]func(interface{})
	onConnected func()
}

type Msg struct{
	Cmd uint32
	Len uint32
	Seq uint32
	Session uint16
	Ver uint16
	Data []byte
	Client *AutoReconnectClient
}

type AsyncMsg struct{
	msg interface{}
	cb func()
}

const(
	PACK_HEAD_LEN = 16
)

func (arcclient *AutoReconnectClient) RecvMsg() interface{} {
	var(
		head = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err error
		msg *Msg
	)

	conn, ok := arcclient.Conn.(io.Reader)
	if !ok {
		goto Exit
	}
	readLen, err = io.ReadFull(conn, head)
	if err != nil || readLen < PACK_HEAD_LEN {
		goto Exit
	}

	msg = &Msg{
		Cmd: binary.BigEndian.Uint32(head[0:]),
		Len: binary.BigEndian.Uint32(head[4:]) - PACK_HEAD_LEN,
		Seq: binary.BigEndian.Uint32(head[8:]),
		Session: binary.BigEndian.Uint32(head[12:]),
		Ver: binary.BigEndian.Uint32(head[14:]),
		Data: nil,
		Client: arcclient,
	}

	if msg.Len > MAX_PACK_LEN {
		goto Exit
	}
	if msg.Len > 0 {
		msg.Data = append(head, make([]byte, msg.Len)...)
		readLen, err := io.ReadFull(conn, msg.Data[PACK_HEAD_LEN:])
		if err != nil || readLen != int(msg.Len) {
			goto Exit
		}
	}

	return msg

Exit:
	return nil
}

func (arcclient *AutoReconnectClient) HandleMsg(imsg interface{}) {
	defer func(){ recover() }
	if msg, ok := imsg.(*Msg); ok {
		if handler, ok := arcclient.handlerMap[msg.Cmd]; ok {
			handler(msg)
		}else{
			
		}
	}
}

func (arcclient *AutoReconnectClient) StartReader() {
	zed.NewCoroutine(func() {
		for arcclient.Running {
			msg := arcclient.RecvMsg()
			if msg == nil {
				arcclient.Restart()
				break
			}
			arcclient.HandleMsg(msg)
		}
	})
}

func (arcclient *AutoReconnectClient) StartWriter() {
	zed.NewCoroutine(func() {
		if arcclient.chSend != nil {
			for {
				if asyncMsg, ok := <-arcclient.chSend; ok {
					func() {
						defer func() { recover() }()
						arcclient.SendMsg(asyncMsg.msg)
						if asyncMsg.cb != nil {
							time.AfterFunc(func(){
								defer func() { recover() }()
								asyncMsg.cb()
							})
						}
					}()
				}else{
					break
				}
			}
		}
	})
}							

func (arcclient *AutoReconnectClient) Dial() (interface{}, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", arcclient.ServerAddr)
	if err != nil {
		return nil, err
	}

	return net.DialTCP("tcp", nil, tcpAddr)
}

func (arcclient *AutoReconnectClient) Start() {
	zed.NewCoroutine(func() {
		arcclient.Lock()
		defer arcclient.Unlock()

		if !arcclient.Running {
			var err error = nil
			arcclient.Conn, err = arcclient.Dial()
			if err != nil {
				arcclient.ConnTimes++
				arcclient.Restart()
				return
			}
			arcclient.Running = true
			arcclient.chSend = make(chan *AsyncMsg, 100)
			arcclient.StartReader()
			arcclient.StartWriter()
			if arcclient.onConnected != nil {
				arcclient.onConnected()
			}
		}
	})
}

func (arcclient *AutoReconnectClient) SendMsg(imsg interface{}) {
	arcclient.Lock()
	defer arcclient.Unlock()
	if arcclient.Running {
		var(
			writeLen = 0
			buf []byte
			err error
		}
		if msg, ok := imsg.(*Msg); ok {
			if msg.Data != nil {
				msg.Len = uint32(len(msg.Data))
			}else{
				msg.Len = 0
			}
			if msg.Len + PACK_HEAD_LEN > MAX_PACK_LEN {
				goto Exit
			}
			conn, ok := arcclient.Conn.(*net.TCPConn)
			if !ok {
				goto Exit
			}
			if err := conn.SetWriteDeadline(time.Now().Add(SEND_BLOCK_TIME); err != nil {
				goto Exit
			}
			buf = make([]byte, PACK_HEAD_LEN+msg.Len)
			binary.BigEndian.PutUint32(buf[0:], uint32(msg.Cmd))
			binary.BigEndian.PutUint32(buf[4:], uint32(msg.Len+PACK_HEAD_LEN))
			binary.BigEndian.PutUint32(buf[8:], uint32(msg.Seq))
			binary.BigEndian.PutUint32(buf[12:], uint32(msg.Session))
			binary.BigEndian.PutUint32(buf[14:], uint32(msg.Ver))
			if msg.Len > 0 {
				copy(buf[PACK_HEAD_LEN:], msg.Data)
			}

			writeLen, err = conn.Write(buf)
			if err == nil && writeLen == len(buf) {
				return
			}
		}
	}else{
		return
	}
Exit:
	arcclient.Restart()
}

func (arcclient *AutoReconnectClient) SendMsgAsync(msg interface{})
	arcclient.Lock()
	defer arcclient.Unlock()
	if arcclient.Running {
		asyncmsg := &AsyncMsg{
			msg: msg,
			cb: nil,
		}

		if len(argv) > 0 {
			if cb, ok := (argv[0]).(func()); ok {
				asyncmsg.cb = cb
			}
		}
		if arcclient.chSend != nil {
			arcclient.chSend <- asyncmsg
		}
	}
}

func (arcclient *AutoReconnectClient) Restart() {
	zed.NewCoroutine(func() {
		arcclient.Lock()
		defer arcclient.Unlock()
		time.Sleep(time.Second)
		if arcclient.Running {
			arcclient.Running = false
			if conn, ok := arcclient.Conn.(*net.TCPConn); ok {
				conn.Close()
			}
			close(arcclient.chSend)
			arcclient.chSend = nil
		}
		arcclient.Start()
	})
}

func (arcclient *AutoReconnectClient) Stop() {
	zed.NewCoroutine(func() {
		arcclient.Lock()
		defer arcclient.Unlock()

		if arcclient.Running {
			arcclient.Running = false
			if conn, ok := arcclient.Conn.(*net.TCPConn); ok {
				conn.Close()
			}
			close(arcclient.chSend)
			arcclient.chSend = nil
		}
	})
}

func (arcclient *AutoReconnectClient) AddMsgHandler(cmd uint32, handler func(interface{})) {
	arcclient.Lock()
	defer arcclient.Unlock()
	if _, ok := arcclient.handlerMap[cmd]; ok {
		zed.Println("AddMsgHandler Error: ", cmd)
		return
	}
	arcclient.handlerMap[cmd] = handler
}
	
func (arcclient *AutoReconnectClient) RemoveMsgHandler(cmd uint32) {
	arcclient.Lock()
	defer arcclient.Unlock()
	delete(arcclient.handlerMap, cmd)
}

func NewAutoReconnectClient(id string, addr string, onConnect func()) *AutoReconnectClient {
	c := &AutoReconnectClient{
		ID: id,
		Conn: nil,
		ServerAddr: addr
		Running: false,
		ConnTimes: 0,
		chSend: nil,
		handlerMap: make(map[uint32]func(interface{})),
		onConnected: onConnect,
	}
	return c
}
