package zed

import (
	"fmt"
	"strings"
	"time"
)

type logtask struct {
	idx     int
	chMsg   chan *string
	ticker  *time.Ticker
	logType int
	logFile *logfile
	running bool
}

var (
	tags = map[int]string{
		LOG_IDX: LOG_TAG,
	}
	logConf = map[string]int{
		"Info":   LogFile,
		"Warn":   LogFile,
		"Error":  LogCmd,
		"Action": LogFile,
	}

	debug         = true
	maxTagNum     = 0
	infoEnabled   = false
	warnEnabled   = false
	errorEnabled  = false
	actionEnabled = false

	infoCount   = 0
	warnCount   = 0
	errorCount  = 0
	actionCount = 0

	infoLoggerNum int = 1
	arrTaskInfo   []*logtask

	warnLoggerNum int = 1
	arrTaskWarn   []*logtask

	errorLoggerNum int = 1
	arrTaskError   []*logtask

	actionLoggerNum int = 1
	arrTaskAction   []*logtask

	logSep = ""

	Printf  = fmt.Printf
	Println = fmt.Println
)

func (task *logtask) start(taskType string, logType int) {
	task.running = true
	task.chMsg = make(chan *string, 100)
	task.logType = logType
	if logType == LogFile {
		task.ticker = time.NewTicker(time.Second * LOG_FILE_SYNC_INTERNAL)
		task.logFile = CreateLogFile(taskType)
		if task.logFile.NewFile() {
			go func() {
				var s *string
				for {
					select {
					case s = <-task.chMsg:
						if s == nil {
							task.running = false
							return
						}
						task.logFile.Write(s)
					case <-task.ticker.C:
						task.logFile.Save()
					}
				}
			}()
		} else {
			ZLog("logtask start failed, taskType: %s, idx: %d", taskType, task.idx)
		}
	} else {
		go func() {
			var s *string
			for {
				select {
				case s = <-task.chMsg:
					if s == nil {
						task.running = false
						return
					}
					Printf(*s)
				}
			}
		}()
	}

}

func (task *logtask) stop() {
	for msg := range task.chMsg {
		task.logFile.Write(msg)
	}
	if task.logType == LogFile {
		task.ticker.Stop()
		task.logFile.Close()
	}
	close(task.chMsg)
}

func ZLog(format string, v ...interface{}) {
	//fmt.Printf(format+"\n", v...)
	fmt.Printf(strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[ZLog][zed] ", fmt.Sprintf(format, v...), "\n"}, logSep))
}

func LogInfo(tag int, loggerIdx int, format string, v ...interface{}) {
	if infoEnabled {
		if debug {
			if tagstr, ok := tags[tag]; ok {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Info][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				infoCount++
				loggerIdx = loggerIdx % infoLoggerNum
				arrTaskInfo[loggerIdx].chMsg <- &s
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Info][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
					infoCount++
					Println("111 logtask idx: ", loggerIdx, infoLoggerNum)
					loggerIdx = loggerIdx % infoLoggerNum
					arrTaskInfo[loggerIdx].chMsg <- &s
					Println("222 logtask idx: ", loggerIdx, infoLoggerNum)
				}
			}
		}
	}
}

func LogWarn(tag int, loggerIdx int, format string, v ...interface{}) {
	//if warnEnabled && (tags[tag] == true) {
	if warnEnabled {
		if debug {
			if tagstr, ok := tags[tag]; ok {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Warn][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				warnCount++
				loggerIdx = loggerIdx % warnLoggerNum
				arrTaskWarn[loggerIdx].chMsg <- &s
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Warn][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
					warnCount++
					loggerIdx = loggerIdx % warnLoggerNum
					arrTaskWarn[loggerIdx].chMsg <- &s
				}
			}
		}
	}
}

func LogError(tag int, loggerIdx int, format string, v ...interface{}) {
	if errorEnabled {
		if debug {
			if tagstr, ok := tags[tag]; ok {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Error][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				errorCount++
				loggerIdx = loggerIdx % errorLoggerNum
				arrTaskError[loggerIdx].chMsg <- &s
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Error][", "] ", tagstr, fmt.Sprintf(format, v...), "\n"}, logSep)
					errorCount++
					loggerIdx = loggerIdx % errorLoggerNum
					arrTaskError[loggerIdx].chMsg <- &s
				}
			}
		}
	}
}

func LogAction(tag int, loggerIdx int, format string, v ...interface{}) {
	if actionEnabled {
		if debug {
			if tagstr, ok := tags[tag]; ok {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Action][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				errorCount++
				loggerIdx = loggerIdx % actionLoggerNum
				arrTaskAction[loggerIdx].chMsg <- &s
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Action][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
					errorCount++
					loggerIdx = loggerIdx % actionLoggerNum
					arrTaskAction[loggerIdx].chMsg <- &s
				}
			}
		}
	}
}

func StartLogger(logconf map[string]int, isDebug bool, maxTag int, logtags map[int]string, infoLogNum int, warnLogNum int, errorLogNum int, actionLogNum int) {
	logConf = logconf

	debug = isDebug
	maxTagNum = maxTag
	tags = logtags
	tags[LOG_IDX] = LOG_TAG

	infoLoggerNum = infoLogNum
	warnLoggerNum = warnLogNum
	errorLoggerNum = errorLogNum
	actionLoggerNum = actionLogNum

	infoEnabled = (infoLogNum > 0)
	warnEnabled = (warnLogNum > 0)
	errorEnabled = (errorLogNum > 0)
	actionEnabled = (actionLoggerNum > 0)

	var i int
	arrTaskInfo = make([]*logtask, infoLoggerNum)
	for i = 0; i < infoLoggerNum; i++ {
		arrTaskInfo[i] = &logtask{
			idx:     i,
			logFile: nil,
		}
		arrTaskInfo[i].start("Info", logConf["Info"])
	}

	arrTaskWarn = make([]*logtask, warnLoggerNum)
	for i = 0; i < warnLoggerNum; i++ {
		arrTaskWarn[i] = &logtask{
			idx:     i,
			logFile: nil,
		}
		arrTaskWarn[i].start("Warn", logConf["Warn"])
	}

	arrTaskError = make([]*logtask, errorLoggerNum)
	for i = 0; i < errorLoggerNum; i++ {
		arrTaskError[i] = &logtask{
			idx:     i,
			logFile: nil,
		}
		arrTaskError[i].start("Error", logConf["Error"])
	}

	arrTaskAction = make([]*logtask, actionLoggerNum)
	for i = 0; i < actionLoggerNum; i++ {
		arrTaskAction[i] = &logtask{
			idx:     i,
			logFile: nil,
		}
		arrTaskAction[i].start("Action", logConf["Action"])
	}
}

func StopLogger() {
	/*for {
	REP:*/
	for _, task := range arrTaskInfo {
		task.stop()
	}

	for _, task := range arrTaskWarn {
		task.stop()
	}

	for _, task := range arrTaskError {
		task.stop()
	}

	/*	time.Sleep(time.Second / 10)

		for _, task := range arrTaskInfo {
			if task.running {
				goto REP
			}
		}

		for _, task := range arrTaskWarn {
			if task.running {
				goto REP
			}
		}

		for _, task := range arrTaskError {
			if task.running {
				goto REP
			}
		}
	*/
	Println("[ShutDown] Logger Stop!")
	/*	return
		}*/
}
