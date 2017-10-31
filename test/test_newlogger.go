package main

import (
	log "github.com/ccppluagopy/zed/logger"
)

func main() {
	//log.SetLogDir("./logs/", true)
	log.SetLogType(log.LogCmd | log.LogFile)
	log.Info(" %d %s", 111, "hahaha")
}
