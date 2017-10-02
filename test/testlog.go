package main

import (
	"github.com/ccppluagopy/zed/logger"
	"time"
)

const (
	LOG_TAG_ZED = iota
	LOG_TAG_1
	LOG_TAG_2
	LOG_TAG_3
)

var LogConf = map[string]int{
	"Debug":  logger.LogCmd | logger.LogFile,
	"Info":   logger.LogFile,
	"Warn":   logger.LogCmd | logger.LogFile,
	"Error":  logger.LogCmd | logger.LogFile,
	"Action": logger.LogCmd | logger.LogFile,
}

var LogTags = map[int]string{
	LOG_TAG_ZED: "zed",
	LOG_TAG_1:   "LOG_TAG_1",
	LOG_TAG_2:   "LOG_TAG_2",
	LOG_TAG_3:   "LOG_TAG_3",
}

func main() {
	logger.SetLogDir("./logs", false)
	logger.StartLogger(LogConf, LogTags)

	go func() {
		idx := 0
		for {
			idx++
			logger.Info(LOG_TAG_1, "test %d", idx)
		}
	}()

	go func() {
		idx := 10000000
		for {
			idx++
			logger.Info(LOG_TAG_2, "test %d", idx)
		}
	}()

	go func() {
		idx := 20000000
		for {
			idx++
			logger.Info(LOG_TAG_3, "test %d", idx)
		}
	}()

	for {
		time.Sleep(time.Hour)
	}

}
