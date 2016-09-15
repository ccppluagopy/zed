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
	infoEnabled   = true
	warnEnabled   = true
	errorEnabled  = true
	actionEnabled = true

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

	//zlogfile *logfile = nil

	logSep = ""

	Printf  = fmt.Printf
	Println = fmt.Println

	isLoggerStarted = false
)

func (task *logtask) start(taskType string, logType int) {
	task.running = true
	task.chMsg = make(chan *string, 100)
	task.logType = logType
	if logType == LogFile {
		task.ticker = time.NewTicker(time.Second * LOG_FILE_SYNC_INTERNAL)
		task.logFile = CreateLogFile(taskType)
		if task.logFile.NewFile() {
			NewCoroutine(func() {
				var (
					s  *string
					ok = false
				)
				for {
					select {
					case s, ok = <-task.chMsg:
						if !ok {
							//task.stop()
							return
						}
						task.logFile.Write(s)
					case <-task.ticker.C:
						task.logFile.Save()
					}
				}
			})
		} else {

		}
	} else {
		NewCoroutine(func() {
			var (
				s  *string
				ok = false
			)
			for {
				select {
				case s, ok = <-task.chMsg:
					if !ok {
						//task.stop()
						return
					}
					Printf(*s)
				}
			}
		})
	}

}

func (task *logtask) stop() {
	close(task.chMsg)
	for msg := range task.chMsg {
		if task.logType == LogFile {
			task.logFile.Write(msg)
		} else {
			Printf(*msg)
		}
	}

	if task.logType == LogFile {
		task.ticker.Stop()
		task.logFile.Close()
	}

	task.running = false
	ZLog("logtask stopped, taskType: %d, idx: %d", task.logType, task.idx)
}

func ZLog(format string, v ...interface{}) {
	/*if zlogfile != nil {
		s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[ZLog][zed] ", fmt.Sprintf(format, v...), "\n"}, logSep)
		zlogfile.Write(&s)
		zlogfile.Save()
	}*/
	s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[ZLog][zed] ", fmt.Sprintf(format, v...), "\n"}, logSep)
	Printf(s)
}

func LogInfo(tag int, loggerIdx int, format string, v ...interface{}) {
	if infoEnabled {
		if debug {
			if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Info][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				infoCount++
				loggerIdx = loggerIdx % infoLoggerNum
				if arrTaskInfo[loggerIdx].running {
					arrTaskInfo[loggerIdx].chMsg <- &s
				} else {
					ZLog("Error when LogInfo, Task runnign: fasle")
				}
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Info][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
					infoCount++
					loggerIdx = loggerIdx % infoLoggerNum
					if arrTaskInfo[loggerIdx].running {
						arrTaskInfo[loggerIdx].chMsg <- &s
					} else {
						ZLog("Error when LogInfo, Task runnign: fasle")
					}
				}
			}
		}
	}
}

func LogWarn(tag int, loggerIdx int, format string, v ...interface{}) {
	//if warnEnabled && (tags[tag] == true) {
	if warnEnabled {
		if debug {
			if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Warn][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				warnCount++
				loggerIdx = loggerIdx % warnLoggerNum
				if arrTaskWarn[loggerIdx].running {
					arrTaskWarn[loggerIdx].chMsg <- &s
				} else {
					ZLog("Error when LogWarn, Task runnign: fasle")
				}
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Warn][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
					warnCount++
					loggerIdx = loggerIdx % warnLoggerNum
					if arrTaskWarn[loggerIdx].running {
						arrTaskWarn[loggerIdx].chMsg <- &s
					} else {
						ZLog("Error when LogWarn, Task runnign: fasle")
					}
				}
			}
		}
	}
}

func LogError(tag int, loggerIdx int, format string, v ...interface{}) {
	if errorEnabled {
		if debug {
			if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Error][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				errorCount++
				loggerIdx = loggerIdx % errorLoggerNum
				if arrTaskError[loggerIdx].running {
					arrTaskError[loggerIdx].chMsg <- &s
				} else {
					ZLog("Error when LogError, Task runnign: fasle")
				}
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Error][", "] ", tagstr, fmt.Sprintf(format, v...), "\n"}, logSep)
					errorCount++
					loggerIdx = loggerIdx % errorLoggerNum
					if arrTaskError[loggerIdx].running {
						arrTaskError[loggerIdx].chMsg <- &s
					} else {
						ZLog("Error when LogError, Task runnign: fasle")
					}
				}
			}
		}
	}
}

func LogAction(tag int, loggerIdx int, format string, v ...interface{}) {
	if actionEnabled {
		if debug {
			if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Action][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				errorCount++
				loggerIdx = loggerIdx % actionLoggerNum
				if arrTaskAction[loggerIdx].running {
					arrTaskAction[loggerIdx].chMsg <- &s
				} else {
					ZLog("Error when LogAction, Task runnign: fasle")
				}
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Action][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
					errorCount++
					loggerIdx = loggerIdx % actionLoggerNum
					if arrTaskAction[loggerIdx].running {
						arrTaskAction[loggerIdx].chMsg <- &s
					} else {
						ZLog("Error when LogAction, Task runnign: fasle")
					}
				}
			}
		}
	}
}

//func StartLogger(logconf map[string]int, isDebug bool, maxTag int, logtags map[int]string, infoLogNum int, warnLogNum int, errorLogNum int, actionLogNum int) {
func StartLogger(logconf map[string]int, logtags map[int]string, isDebug bool) {
	if logconf != nil {
		logConf = logconf
		if n, ok := logconf["InfoCorNum"]; ok {
			infoLoggerNum = n
		}
		if n, ok := logconf["WarnCorNum"]; ok {
			warnLoggerNum = n
		}
		if n, ok := logconf["ErrorCorNum"]; ok {
			errorLoggerNum = n
		}
		if n, ok := logconf["ActionCorNum"]; ok {
			actionLoggerNum = n
		}
		infoEnabled = (infoLoggerNum > 0)
		warnEnabled = (warnLoggerNum > 0)
		errorEnabled = (errorLoggerNum > 0)
		actionEnabled = (actionLoggerNum > 0)
	}
	if logtags != nil {
		tags = logtags
		for i := 0; ; i++ {
			if _, ok := logtags[i]; ok {
				maxTagNum = i
			} else {
				break
			}
		}
	}

	debug = isDebug
	//maxTagNum = maxTag

	//tags[LOG_IDX] = LOG_TAG

	var i int
	arrTaskInfo = make([]*logtask, infoLoggerNum)
	for i = 0; i < infoLoggerNum; i++ {
		arrTaskInfo[i] = &logtask{
			idx:     i,
			logFile: nil,
		}
	}
	for i = 0; i < infoLoggerNum; i++ {
		arrTaskInfo[i].start("Info", logConf["Info"])
	}

	arrTaskWarn = make([]*logtask, warnLoggerNum)
	for i = 0; i < warnLoggerNum; i++ {
		arrTaskWarn[i] = &logtask{
			idx:     i,
			logFile: nil,
		}
	}
	for i = 0; i < warnLoggerNum; i++ {
		arrTaskWarn[i].start("Warn", logConf["Warn"])
	}

	arrTaskError = make([]*logtask, errorLoggerNum)
	for i = 0; i < errorLoggerNum; i++ {
		arrTaskError[i] = &logtask{
			idx:     i,
			logFile: nil,
		}
	}
	for i = 0; i < errorLoggerNum; i++ {
		arrTaskError[i].start("Error", logConf["Error"])
	}

	arrTaskAction = make([]*logtask, actionLoggerNum)
	for i = 0; i < actionLoggerNum; i++ {
		arrTaskAction[i] = &logtask{
			idx:     i,
			logFile: nil,
		}
	}
	for i = 0; i < actionLoggerNum; i++ {
		arrTaskAction[i].start("Action", logConf["Action"])
	}

	isLoggerStarted = true

	Println("======================================================================")
	Println("StartLogger")
	Println("logConf:")
	for k, v := range logConf {
		str := "LogCmd"
		if v == LogFile {
			str = "LogFile"
		}
		Println("	", k, str)
	}
	Println(" - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -")
	Println("tags:")
	for i := 0; ; i++ {
		if tag, ok := tags[i]; ok {
			if tag == "" {
				tag = "		not set, log of this tag will not be recorded!"
			}
			Println("	", i, tag)
		} else {
			break
		}
	}
	Println("======================================================================")
}

func StopLogger() {
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
	ZLog("[ShutDown] Logger Stopped!")
	/*if zlogfile != nil {
		zlogfile.Close()
	}*/
	/*	return
		}*/
}
