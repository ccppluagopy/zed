package zed

/*import (
	"reflect"
)*/

type ClosureCB func()

type TimerCallBack func()

//type DBErrorHandler func()

type EventHandler func(event interface{}, args []interface{})

type MsgHandler func(msg *NetMsg) bool

type ClientCloseCB func(client *TcpClient)

type MongoActionCB func(mongo *MongoMgr) bool

type MysqlActionCB func(mysql *MysqlMgr) bool

func NewCoroutine(cb ClosureCB) {
	go func() {
		defer PanicHandle(true)
		cb()
	}()
}

/*
func NewCoroutine(cb interface{}, args ...interface{}) {
	f := reflect.ValueOf(cb)
	if f.Kind() == reflect.Func {
		go func() {
			defer PanicHandle(true)
			n := len(args)
			if n > 0 {
				refargs := make([]reflect.Value, n)
				for i := 0; i < n; i++ {
					refargs[i] = reflect.ValueOf(args[i])
				}
				f.Call(refargs)

			} else {
				f.Call(nil)
			}
		}()
	} else {
		ZLog("NewCoroutine Error: cb is not function")
		Println("NewCoroutine Error: cb is not function")
	}
}
*/
