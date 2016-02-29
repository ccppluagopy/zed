package test

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"time"
)

func EventMgrExample() {
	mgr := zed.NewEventMgr("haha")
	if emgr, ok := GetEventMgrByTag("haha"); ok {

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
		Tag1 = iota
		Tag2
		Tag3
		TagMax
	)
	var LogTags = map[int]string{
		Tag1: "Tag1",
		Tag2: "Tag2",
		Tag3: "Tag3",
	}
	var LogConf = map[string]int{
		"Info":  LogFile,
		"Warn":  LogFile,
		"Error": LogCmd,
	}

	//zed.StartLogger(isDebug bool, maxTag int, logtags map[int]string, infoLogNum int, warnLogNum int, errorLogNum int) {
	zed.StartLogger(LogConf, true, TagMax, LogTags, 3, 3, 3)
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
