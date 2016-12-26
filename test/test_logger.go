package main

import (
	"github.com/ccppluagopy/zed"
)

func main() {
	workDir := "./"
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

	zed.Init(workDir, logDir)
	zed.StartLogger(LogConf, LogTags, true)

	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			zed.LogInfo(Tag1, i, "test log i: %d", i)
		} else {
			zed.LogInfo(Tag2, i, "test log i: %d", i)
		}
	}

	zed.StopLogger()
}
