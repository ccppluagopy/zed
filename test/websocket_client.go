package wsclient

import(
	"encoding/binary"
	"github.com/ccppluagopy/zed"
	"golang.org/x/net/websocket"
	"io"
	"time"
)

type WSClient struct {
	RWMutex
	ID string
	Conn *websocket.Conn
	Url string
	Origin string
	Over bool
	Running bool
	ConnTimes int
	chSend chan *WSAsyncMsg
	handlerMap map[uint32]func(*WSMsg)
	onConnect func()
}

type WSMsg struct {
	Len uint32
	Cmd uint32
	Data []byte
	Client *WSClient
}

type WSAsyncMsg struct {
	msg *WSMsg
	cb func()
}

const (
	PACK_HEAD_LEN = 8
)

func (wsclient *WSClient) RecvMsg() *WSMsg() {
	return wsclient.RecvMsg()
}

func (wsclient *WSClient) HandleMsg(msg *WSMsg) {
	if handler, ok := wsclient.handlerMap[msg.Cmd]; ok{
		handler(msg)
	} else {
		
	}
}

func (wsclient *WSClient) StartReader() {
	zed.NewCoroutine(func() {
		for !wsclient.Over && wsclient.Running {
			msg := wsclient.RecvMsg()
			if msg == nil {
				wsclient.Restart()
				break
			}
		}
		wsclient.HandleMsg(msg)
	})
}

func (wsclient *WSClient) StartWriter() {
	zed.NewCoroutine(func() {
		if wsclient.chSend != nil {
			for {
				if asyncMsg, ok := <-wsclient.chSend; ok{
					func() {
						defer func(){recover()}()
						wsclient.SendMsg(asyncMsg.msg)
						if asyncMsg.cb != nil {
							zed.NewCoroutine(func() {
								asyncMsg.cb()
							})
						}
					}()
				}
			}
		}
	})
}

func (wsclient *WSClient) Start() {
	zed.NewCoroutine(func(){
		wsclient.Lock()
		defer wsclient.Unlock()
		if wsclient.Over && !wsclient.Running {
			var err error = nil
			wsclient.Conn, err = websocket.Dial(wsclient.Url, "", wsclient.Origin)
			
			if err != nil {
				wsclient.ConnTimes++
				zed.ZLog("WSClient Start Faile %d\n", wsclient.ConnTimes)
				if wsclient.Over {
					wsclient.Restart()
					return
				}
			}

			wsclient.Conn.PayloadType = websocket.BinaryFrame
			wsclient.Over = false
			wsclient.Running = true
			wsclient.chSend = make(chan *WSAsyncMsg, 100)
			wsclient.StartReader()
			wsclient.StartWriter()

			if wsclient.onConnect != nil {
				wsclient.onConnect()
			}
		}
	})
}

func (wsclient *WSClient) SendMsg(msg *WSMsg){
	wsclient.Lock()
	defer wsclient.Unlock()

ErrExit:
	wsclient.Restart()
	return
}

func (wsclient *WSClient) SendMsgAsync(msg *WSMsg, arg ...interface{}) {
	wsclient.Lock()
	defer wsclient.Unlock()

	if !wsclient.Over && wsclient.Running {
		asyncmsg := &WSAsyncMsg{
			msg: msg,
			cb: nil,
		}
		if len(argv) > 0 {
			if cb, ok := (argv[0]).(func()); ok{
				asyncmsg.cb = cb
			}
		}
		if wsclient.chSend != nil {
			wsclient.chSend <- asyncmsg
		}
	}
}

func (wsclient *WSClient) Restart() {
	zed.LogStackInfo()
	zed.NewCoroutine(func(){
		wsclient.Lock()
		defer wsclient.Unlock()
		if !wsclient.Over {
			zed.Println("------ Restart()", wsclient.Over, wsclient.Running)
			wsclient.Over = true
			if wsclient.Running {
				wsclient.Running = false
				wsclient.Conn.Close()
				close(wsclient.chSend)
				wsclient.chSend = nil
			}
		}
	
		wsclient.Start()
	})
}

func (wsclient *WSClient) Stop() {
	zed.NewCoroutine(func() {
		wsclient.Lock()
		defer wsclient.Unlock()

		if !wsclient.Over {
			wsclient.Over = true
			if wsclient.Running {
				wsclient.Running = false
				wsclient.Conn.Close()
				close(wsclient.chSend)
				wsclient.chSend = nil
			}
		}
	})
}

func (wsclient *WSClient) AddMsgHandler(cmd uint32, handler func(*WSMsg)){
	wsclient.Lock()
	defer wsclient.Unlock()
	wsclient.handlerMap[cmd] = func(msg *WSMsg) {
		defer zed.HandlePanic(true)
		handler(msg)
	}
}

func NewWSClient(id string, url string, origin string, onConnect func()) *WSClient {
	return &WSClient{
		ID: id,
		Conn: nil,
		Url: url,
		Origin: origin,
		Over: true,
		Running: false,
		ConnTimes: 0,
		chSend: nil,
		handlerMap: make(map[uint32]func(*WSMsg)),
		onConnect: onConnect,
	}
}












