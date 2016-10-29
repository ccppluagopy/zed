package zed

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type logtask struct {
	sync.Mutex
	idx     int
	chMsg   chan *string
	ticker  *time.Ticker
	logType int
	logFile *logfile
	running bool
}

const (
	LOG_LEVEL_INFO = iota
	LOG_LEVEL_WARN
	LOG_LEVEL_ERROR
	LOG_LEVEL_ACTION
	LOG_LEVEL_MAX
)

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
	Sprintf = fmt.Sprintf
	Println = fmt.Println

	isLoggerStarted = false

	loglevel = LOG_LEVEL_INFO
)

func SetLogLevel(l int) {
	if l >= LOG_LEVEL_INFO && l < LOG_LEVEL_MAX {
		loglevel = l
	} else {
		ZLog("SetLogLevel Error: Invalid Level.")
	}
}

func (task *logtask) start(taskType string, logType int) {
	task.running = true
	//task.chMsg = make(chan *string, 100)
	task.logType = logType
	if logType == LogFile {
		task.ticker = time.NewTicker(time.Second * LOG_FILE_SYNC_INTERNAL)
		task.logFile = CreateLogFile(taskType)
		if task.logFile.NewFile() {
			NewCoroutine(func() {
				/*var (
					s  *string
					ok = false
				)*/
				for {
					select {
					/*
						case s, ok = <-task.chMsg:
							if !ok {
								return
							}
							task.logFile.Write(s)
					*/
					case <-task.ticker.C:
						task.save()
					}
				}
			})
		} else {

		}
	}
	/* else {
		NewCoroutine(func() {
			var (
				s  *string
				ok = false
			)
			for {
				select {
				case s, ok = <-task.chMsg:
					if !ok {
						return
					}
					Printf(*s)
				}
			}
		})
	}*/
}

func (task *logtask) stop() {
	task.Lock()
	defer task.Unlock()

	/*close(task.chMsg)
	for msg := range task.chMsg {
		if task.logType == LogFile {
			task.logFile.Write(msg)
		} else {
			Printf(*msg)
		}
	}*/

	if task.logType == LogFile {
		task.ticker.Stop()
		task.logFile.Close()
	}

	task.running = false
	ZLog("logtask stopped, taskType: %d, idx: %d", task.logType, task.idx)
}

func (task *logtask) save() {
	task.Lock()
	defer task.Unlock()
	task.logFile.Save()
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

func LogInfoSave(tag int, loggerIdx int) {
	loggerIdx = loggerIdx % infoLoggerNum

	arrTaskInfo[loggerIdx].Lock()
	defer arrTaskInfo[loggerIdx].Unlock()

	arrTaskInfo[loggerIdx].logFile.Save()
}

func LogInfo(tag int, loggerIdx int, format string, v ...interface{}) {
	if infoEnabled && loglevel >= LOG_LEVEL_INFO {
		loggerIdx = loggerIdx % infoLoggerNum

		arrTaskInfo[loggerIdx].Lock()
		defer arrTaskInfo[loggerIdx].Unlock()

		if debug {
			if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Info][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				infoCount++

				if arrTaskInfo[loggerIdx].running {
					//arrTaskInfo[loggerIdx].chMsg <- &s
					if arrTaskInfo[loggerIdx].logType == LogFile {
						arrTaskInfo[loggerIdx].logFile.Write(&s)
					} else {
						Printf(s)
					}
				} else {
					ZLog("Error when LogInfo, Task runnign: fasle")
				}
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Info][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
					infoCount++

					if arrTaskInfo[loggerIdx].running {
						//arrTaskInfo[loggerIdx].chMsg <- &s
						if arrTaskInfo[loggerIdx].logType == LogFile {
							arrTaskInfo[loggerIdx].logFile.Write(&s)
						} else {
							Printf(s)
						}
					} else {
						ZLog("Error when LogInfo, Task runnign: fasle")
					}
				}
			}
		}
	}
}

func LogWarnSave(tag int, loggerIdx int) {
	loggerIdx = loggerIdx % warnLoggerNum

	arrTaskWarn[loggerIdx].Lock()
	defer arrTaskWarn[loggerIdx].Unlock()

	arrTaskWarn[loggerIdx].logFile.Save()
}

func LogWarn(tag int, loggerIdx int, format string, v ...interface{}) {
	//if warnEnabled && (tags[tag] == true) {
	if warnEnabled && loglevel >= LOG_LEVEL_WARN {
		loggerIdx = loggerIdx % warnLoggerNum

		arrTaskWarn[loggerIdx].Lock()
		defer arrTaskWarn[loggerIdx].Unlock()

		if debug {
			if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Warn][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				warnCount++

				if arrTaskWarn[loggerIdx].running {
					//arrTaskWarn[loggerIdx].chMsg <- &s
					if arrTaskWarn[loggerIdx].logType == LogFile {
						arrTaskWarn[loggerIdx].logFile.Write(&s)
					} else {
						Printf(s)
					}
				} else {
					ZLog("Error when LogWarn, Task runnign: fasle")
				}
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Warn][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
					warnCount++

					if arrTaskWarn[loggerIdx].running {
						//arrTaskWarn[loggerIdx].chMsg <- &s
						if arrTaskWarn[loggerIdx].logType == LogFile {
							arrTaskWarn[loggerIdx].logFile.Write(&s)
						} else {
							Printf(s)
						}
					} else {
						ZLog("Error when LogWarn, Task runnign: fasle")
					}
				}
			}
		}
	}
}

func LogErrorSave(tag int, loggerIdx int) {
	loggerIdx = loggerIdx % errorLoggerNum

	arrTaskError[loggerIdx].Lock()
	defer arrTaskError[loggerIdx].Unlock()

	arrTaskError[loggerIdx].logFile.Save()
}

func LogError(tag int, loggerIdx int, format string, v ...interface{}) {
	if errorEnabled && loglevel >= LOG_LEVEL_ERROR {
		loggerIdx = loggerIdx % errorLoggerNum

		arrTaskError[loggerIdx].Lock()
		defer arrTaskError[loggerIdx].Unlock()

		if debug {
			if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Error][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				errorCount++

				if arrTaskError[loggerIdx].running {
					//arrTaskError[loggerIdx].chMsg <- &s
					if arrTaskError[loggerIdx].logType == LogFile {
						arrTaskError[loggerIdx].logFile.Write(&s)
					} else {
						Printf(s)
					}
				} else {
					ZLog("Error when LogError, Task runnign: fasle")
				}
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().Format("20060102-150405")), "[Error][", "] ", tagstr, fmt.Sprintf(format, v...), "\n"}, logSep)
					errorCount++

					if arrTaskError[loggerIdx].running {
						//arrTaskError[loggerIdx].chMsg <- &s
						if arrTaskError[loggerIdx].logType == LogFile {
							arrTaskError[loggerIdx].logFile.Write(&s)
						} else {
							Printf(s)
						}
					} else {
						ZLog("Error when LogError, Task runnign: fasle")
					}
				}
			}
		}
	}
}

func LogActionSave(tag int, loggerIdx int) {
	loggerIdx = loggerIdx % errorLoggerNum

	arrTaskAction[loggerIdx].Lock()
	defer arrTaskAction[loggerIdx].Unlock()

	arrTaskAction[loggerIdx].logFile.Save()
}

func LogAction(tag int, loggerIdx int, format string, v ...interface{}) {
	if actionEnabled && loglevel >= LOG_LEVEL_ACTION {
		loggerIdx = loggerIdx % actionLoggerNum

		arrTaskAction[loggerIdx].Lock()
		defer arrTaskAction[loggerIdx].Unlock()

		if debug {
			if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
				s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Action][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
				errorCount++

				if arrTaskAction[loggerIdx].running {
					//arrTaskAction[loggerIdx].chMsg <- &s
					if arrTaskError[loggerIdx].logType == LogFile {
						arrTaskError[loggerIdx].logFile.Write(&s)
					} else {
						Printf(s)
					}
				} else {
					ZLog("Error when LogAction, Task runnign: fasle")
				}
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
					s := strings.Join([]string{fmt.Sprintf("[%s]", time.Now().String()), "[Action][", tagstr, "] ", fmt.Sprintf(format, v...), "\n"}, logSep)
					errorCount++

					if arrTaskAction[loggerIdx].running {
						//arrTaskAction[loggerIdx].chMsg <- &s
						if arrTaskError[loggerIdx].logType == LogFile {
							arrTaskAction[loggerIdx].logFile.Write(&s)
						} else {
							Printf(s)
						}
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
