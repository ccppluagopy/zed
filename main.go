package main

import (
	"github.com/ccppluagopy/zed/test"
	"time"
)

func main() {

	test.EventMgrExample()
	test.LoggerExample()
	test.TimerMgrExample()

	for {
		time.Sleep(time.Hour)
	}
}
