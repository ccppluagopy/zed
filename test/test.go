package test

import (
	"encoding/binary"
	"fmt"
	"github.com/ccppluagopy/zed"
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
	zed.MakeNewLogDir("./")
	const (
		TagNull = iota
		Tag1
		Tag2
		Tag3
		TagMax
	)
	var LogTags = map[int]string{
		TagNull: "zed",
		Tag1:    "Tag1",
		Tag2:    "Tag2",
		Tag3:    "Tag3",
	}
	var LogConf = map[string]int{
		"Info":   zed.LogFile,
		"Warn":   zed.LogFile,
		"Error":  zed.LogCmd,
		"Action": zed.LogFile,
	}

	zed.StartLogger(LogConf, true, TagMax, LogTags, 3, 3, 3, 3)
	for i := 0; i < 5; i++ {
		zed.LogError(Tag1, i, "log test %d", i)
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
	zed.NewTcpServer(10, 20).Start(addr)
}

func TestEchoClientForTcpServer(addr string, clientNum int) {
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
		for {
			n = n + 1
			conn.Write(buf)
			_, err := io.ReadFull(conn, buf2)
			if err == nil {
				fmt.Println(fmt.Sprintf("Client %d Recv Msg %d: %s", idx, n, string(buf2[8:])))
			} else {
				checkError(err)
				break
			}
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
		go robot(i+1, conns[i])
	}

	var str string
	fmt.Scanf("%s", &str)
}

func TestBase() {
	addr := "127.0.0.1:9999"
	TestLogger()
	TestEventMgr()

	go func() {
		time.Sleep(time.Second)
		TestEchoClientForTcpServer(addr, 10)
	}()

	TestTcpServer(addr)
}
