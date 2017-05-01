package main

import (
	zed "github.com/ccppluagopy/zed/logger"
	"time"
)

func main() {
	//workDir := "./"
	logDir := "./logs/"

	const (
		TagZed = iota
		Tag1
		Tag2
	)

	var LogTags = map[int]string{
		TagZed: "--zed", /*'--'开头则关闭*/
		Tag1:   "Tag1",
		Tag2:   "Tag2",
	}

	var LogConf = map[string]int{
		"Info":         zed.LogFile,
		"Warn":         zed.LogCmd,
		"Error":        zed.LogCmd,
		"Action":       zed.LogCmd,
		"InfoCorNum":   2,
		"WarnCorNum":   3,
		"ErrorCorNum":  4,
		"ActionCorNum": 5,
	}

	zed.SetLogDir(logDir)
	zed.StartLogger(LogConf, LogTags)
	test := func() {
		i := 0
		for {
			time.Sleep(time.Second)
			if i%2 == 0 {
				zed.LogInfo(Tag1, "test log i: %d", i)
			} else {
				zed.LogInfo(Tag2, "test log i: %d", i)
			}
			i++
		}
	}

	go test()
	go test()
	test()
	zed.StopLogger()
}
