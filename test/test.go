package test

import (
	"encoding/binary"
	"fmt"
	"github.com/ccppluagopy/zed"
	"github.com/ccppluagopy/zed/mutex"
	"io"
	"net"
	"time"
)

func TestEventMgr() {
	mgr := zed.NewEventMgr("haha")
	if emgr, ok := zed.GetEventMgrByTag("haha"); ok {

		emgr.NewListener("listener_001", 3, func(e interface{}, args []interface{}) {
			fmt.Println("--- event 001 : ", e, args[0])
		})
		emgr.NewListener("listener_002", 3, func(e interface{}, args []interface{}) {
			fmt.Println("--- event 002 : ", e, args[0], args[1])
		})
		emgr.NewListener("listener_003", 3, func(e interface{}, args []interface{}) {
			fmt.Println("--- event 003 : ", e, args[0], args[1], args[2], args[3])
		})
		emgr.NewListener("listener_004", 3, func(e interface{}, args []interface{}) {
			fmt.Println("--- event 004 : ", e, args[0], args[1])
		})

		emgr.Dispatch(3, "ss", 10000)
		mgr.Dispatch(3, "yy", 10000, "xx")
	}
}

func TestLogger() {
	zed.Init("./", "./logs/", true, nil, nil)
	//zed.MakeNewLogDir("./")
	const (
		TagZed = iota
		TagInit
		TagDB
		TagLogin
		TagRoom
		TagRoomMgr
		TagMax
	)

	var LogTags = map[int]string{
		TagZed:     "--zed", /*'--'开头则关闭*/
		TagInit:    "Init",
		TagDB:      "DB",
		TagLogin:   "--Login",
		TagRoom:    "Room",
		TagRoomMgr: "RoomMgr",
	}

	var LogConf = map[string]int{
		"Info":         zed.LogCmd,
		"Warn":         zed.LogCmd,
		"Error":        zed.LogCmd,
		"Action":       zed.LogCmd,
		"InfoCorNum":   0,
		"WarnCorNum":   1,
		"ErrorCorNum":  1,
		"ActionCorNum": 1,
	}

	zed.StartLogger(LogConf, LogTags, true)
	for i := 0; i < 5; i++ {
		zed.LogError(TagDB, i, "log test %d", i)
	}

}

func TestTimerMgr() {
	timerMgr := zed.NewTimerMgr(int64(3))

	cb1 := func() {
		fmt.Println("cb1")
	}
	cb2 := func() {
		fmt.Println("cb2")
	}

	timerMgr.NewTimer("cb1", int64(time.Second), int64(time.Second), cb1, true)
	timerMgr.NewTimer("cb2", int64(time.Second), int64(time.Second*2), cb2, true)
}

func TestTimerWheel() {
	timerWheel := zed.NewTimerWheel(int64(5), int64(time.Second), 2)

	cb1 := func() {
		fmt.Println("cb1")
	}
	cb2 := func() {
		fmt.Println("cb2")
	}

	timerWheel.NewTimer("cb1", 0, cb1, true)
	time.Sleep(time.Second * 1)
	timerWheel.NewTimer("cb2", 0, cb2, true)
}

func TestTcpServer(addr string) {
	zed.HandleSignal(true)
	server := zed.NewTcpServer("testserver")

	/*go func() {
		time.Sleep(time.Second * 3)
		server.Stop()
	}()*/

	server.Start(addr)
}

func TestEchoClientForTcpServer(addr string, clientNum int) {
	zed.HandleSignal(true)

	var (
		ROBOT_NUM = clientNum
		buf       = make([]byte, len("hello world")+8)
		buf2      = make([]byte, len(buf))
		conns     = make([]net.Conn, ROBOT_NUM)
	)

	checkError := func(err error) {
		if err != nil {
			fmt.Println(err)
		}
	}

	robot := func(idx int, conn net.Conn) {
		n := 0
		var err error
		var nwrite = 0
		for {
			n = n + 1
			nwrite, err = conn.Write(buf)
			if err == nil {
				fmt.Println(fmt.Sprintf("Client %d Send Msg %d: %s", idx, nwrite, string(buf[8:])))
			} else {
				checkError(err)
				break
			}
			nwrite, err = io.ReadFull(conn, buf2)
			if err == nil {
				fmt.Println(fmt.Sprintf("Client %d Recv Msg %d: %s", idx, nwrite, string(buf2[8:])))
			} else {
				checkError(err)
				break
			}
			//time.Sleep(time.Second)
		}
	}

	binary.LittleEndian.PutUint32(buf[0:4], uint32(len(buf)-8))
	binary.LittleEndian.PutUint32(buf[4:8], 899)
	copy(buf[8:], fmt.Sprintf("hello world"))

	for i := 0; i < ROBOT_NUM; i++ {
		conn, err := net.Dial("tcp", addr)
		checkError(err)
		conns[i] = conn

	}

	for i := 0; i < ROBOT_NUM; i++ {
		idx := i
		zed.NewCoroutine(func() {
			go robot(idx, conns[idx])
		})
	}

	var str string
	fmt.Scanf("%s", &str)
}

func TestMutex() {
	const (
		key  = "key"
		addr = "127.0.0.1:33333"
	)

	mutex1 := func() {
		client := mutex.NewMutexClient("mutex1", addr)
		for {
			client.Lock("")
			client.Lock(key)
			time.Sleep(time.Second * 1)
			client.UnLock(key)
		}
	}

	mutex2 := func() {
		time.Sleep(time.Second)
		client := mutex.NewMutexClient("mutex2", addr)
		n := 0
		for {
			client.Lock(key)
			client.UnLock(key)
			n = n + 1
			zed.Println("mutex2 action .......", n)
		}
	}

	mutex3 := func() {
		time.Sleep(time.Second * 1)
		client := mutex.NewMutexClient("mutex3", addr)
		n := 0
		for {
			client.Lock(key)
			client.UnLock(key)
			n = n + 1
			zed.Println("mutex3 action .......", n)
			time.Sleep(time.Second * 1)
		}
	}

	mutex.NewMutexServer("test", addr)
	time.Sleep(time.Second)

	go mutex1()
	go mutex2()
	go mutex3()

	time.Sleep(time.Hour)
}

func TestBase() {
	addr := "127.0.0.1:9999"
	TestLogger()
	TestEventMgr()

	zed.NewCoroutine(func() {
		time.Sleep(time.Second)
		TestEchoClientForTcpServer(addr, 10)
	})

	zed.HandleSignal(true)
	server := zed.NewTcpServer("testserver")

	zed.NewCoroutine(func() {
		time.Sleep(time.Second)
		server.Stop()
		zed.StopLogger()
	})

	server.Start(addr)
	//TestTcpServer(addr)
}
