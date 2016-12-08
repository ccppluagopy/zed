package observer

import (
	"fmt"
	"testing"
	"time"
	//"github.com/ccppluagopy/zed"
)

func TestObserverServer(t *testing.T) {
	//timewheel = zed.NewTimerWheel(10*time.Second, 3)
	addr := "127.0.0.1:3333"
	go NewOBServer("server").Start(addr)
	time.Sleep(2 * time.Second)

	obsc := NewOBClient(addr, time.Second*1000)
	if obsc == nil {
		fmt.Println("obsc is nil.....")
		return
	}
	event := "asdfja"
	obsc.Regist(event, []byte("regist test"))
	obsc.NewListener("xx", event, func(e interface{}, args []interface{}) {
		fmt.Println("--- event 001 : ", e, args[0])
	})
	fmt.Println("xxxxxxxxxx 111")
	obsc.Publish(event, []byte("punlish test"))
	fmt.Println("xxxxxxxxxx 222")
	obsc.Unregist(event, []byte("unregist test event"))
	fmt.Println("xxxxxxxxxx 333")
	time.Sleep(10 * time.Second)
}

// func TestObserverClient(t *testing.T) {
// 	timewheel = zed.NewTimerWheel(10*time.Second, 3)
//
// 	obsc := NewOBClient("localhost:9999")
// 	if obsc == nil {
// 		fmt.Println("obsc is nil.....")
// 		return
// 	}
// 	obsc.RegistReq("test", []byte("test"))
// 	obsc.PublicshReq("test", []byte("punlish test"))
// 	obsc.UnRegistReq("test", []byte("unregist test event"))
//
// 	time.Sleep(10 * time.Second)
// }
