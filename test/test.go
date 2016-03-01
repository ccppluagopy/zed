package test

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"time"
)

func EventMgrExample() {
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

func LoggerExample() {
	const (
		TagNull = iota
		Tag1
		Tag2
		Tag3
		TagMax
	)
	var LogTags = map[int]string{
		TagNull: "",
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

	zed.StartLogger(LogConf, true, TagMax, LogTags, 3, 3, 3, 19)
	for i := 0; i < 5; i++ {
		zed.LogError(Tag1, i, "log test %d", i)
	}

}

func TimerMgrExample() {
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

func TimerWheelExample() {
	timerWheel := Timer.NewTimerWheel(int64(tickTime), int64(time.Second), 2)

	cb1 := func() {
		fmt.Println("cb1")
	}
	cb2 := func() {
		fmt.Println("cb2")
	}

	timerWheel.NewTimer("cb1", cb1, true)
	time.Sleep(time.Second * 1)
	timerWheel.NewTimer("cb2", cb2, true)
}
