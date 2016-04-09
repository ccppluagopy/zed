package zed

type ClosureCB func()

type TimerCallBack func()

type DBErrorHandler func()

type EventHandler func(event interface{}, args []interface{})

type MsgHandler func(msg *NetMsg) bool

type ClientCloseCB func(client *TcpClient)

func NewCoroutine(cb ClosureCB) {
	go func() {
		defer PanicHandle(true)
		cb()
	}()
}
