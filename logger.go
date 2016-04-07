package zed

import (
	"fmt"
	"strings"
	"time"
)

type logtask struct {
	running  bool
	chMsg    chan *string
	taskType string
	ticker   *time.Ticker
	logFile  *logfile
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
	infoOutput    = LogCmd
	warnOutput    = LogCmd
	errorOutput   = LogCmd
	debug         bool
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

	logSep = ""

	Printf  = fmt.Printf
	Println = fmt.Println
	ZLog    = func(format string, v ...interface{}) {
		fmt.Printf(format+"\n", v...)
	}
)

func (task *logtask) start(taskType string) {
	task.running = true
	task.taskType = taskType
	task.chMsg = make(chan *string, 100)
	task.logFile = NewLogFile()

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
				Println(*s)
			case <-task.ticker.C:
				task.logFile.Save()
			}
		}
	}()
	Printf("Logger Start Log %s Task\n", taskType)
}

func (task *logtask) stop() {
	for _, msg := range task.chMsg {
		task.logFile.Write(msg)
	}
	task.ticker.Stop()
	task.logFile.Close()
	close(task.chMsg)
}

func LogInfo(tag int, loggerIdx int, format string, v ...interface{}) {
	if infoEnabled {
		if debug {
			if tagstr, ok := tags[tag]; ok {
				s := strings.Join([]string{"[Info] [", tagstr, fmt.Sprintf("] [%s] ", time.Now().String()), fmt.Sprintf(format, v...)}, logSep)
				infoCount++
				loggerIdx = loggerIdx % infoLoggerNum
				arrTaskInfo[loggerIdx].chMsg <- &s
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok {
					s := strings.Join([]string{"[Info] [", tagstr, fmt.Sprintf("] [%s] ", time.Now().String()), fmt.Sprintf(format, v...)}, logSep)
					infoCount++
					loggerIdx = loggerIdx % infoLoggerNum
					arrTaskInfo[loggerIdx].chMsg <- &s
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
				s := strings.Join([]string{"[Warn] [", tagstr, fmt.Sprintf("] [%s] ", time.Now().String()), fmt.Sprintf(format, v...)}, logSep)
				warnCount++
				loggerIdx = loggerIdx % warnLoggerNum
				arrTaskWarn[loggerIdx].chMsg <- &s
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok {
					s := strings.Join([]string{"[Warn] [", tagstr, fmt.Sprintf("] [%s] ", time.Now().String()), fmt.Sprintf(format, v...)}, logSep)
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
				s := strings.Join([]string{"[Error] [", tagstr, fmt.Sprintf("] [%s] ", time.Now().String()), fmt.Sprintf(format, v...)}, logSep)
				errorCount++
				loggerIdx = loggerIdx % errorLoggerNum
				arrTaskError[loggerIdx].chMsg <- &s
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok {
					s := strings.Join([]string{"[Error] [", tagstr, fmt.Sprintf("] [%s] ", time.Now().String()), fmt.Sprintf(format, v...)}, logSep)
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
				s := strings.Join([]string{"[Action] [", tagstr, fmt.Sprintf("] [%s] ", time.Now().String()), fmt.Sprintf(format, v...)}, logSep)
				errorCount++
				loggerIdx = loggerIdx % actionLoggerNum
				arrTaskAction[loggerIdx].chMsg <- &s
			}
		} else {
			if tag < maxTagNum {
				if tagstr, ok := tags[tag]; ok {
					s := strings.Join([]string{"[Action] [", tagstr, fmt.Sprintf("] [%s] ", time.Now().String()), fmt.Sprintf(format, v...)}, logSep)
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
	infoOutput = logConf["Info"]
	warnOutput = logConf["Warn"]
	errorOutput = logConf["Error"]

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
			logFile: nil,
			ticker:  time.NewTicker(time.Second * LOG_FILE_SYNC_INTERNAL),
		}
		arrTaskInfo[i].start("Info")
	}

	arrTaskWarn = make([]*logtask, warnLoggerNum)
	for i = 0; i < warnLoggerNum; i++ {
		arrTaskWarn[i] = &logtask{
			logFile: nil,
			ticker:  time.NewTicker(time.Second * LOG_FILE_SYNC_INTERNAL),
		}
		arrTaskWarn[i].start("Warn")
	}

	arrTaskError = make([]*logtask, errorLoggerNum)
	for i = 0; i < errorLoggerNum; i++ {
		arrTaskError[i] = &logtask{
			logFile: nil,
			ticker:  time.NewTicker(time.Second * LOG_FILE_SYNC_INTERNAL),
		}
		arrTaskError[i].start("Error")
	}

	arrTaskAction = make([]*logtask, actionLoggerNum)
	for i = 0; i < actionLoggerNum; i++ {
		arrTaskAction[i] = &logtask{
			logFile: nil,
			ticker:  time.NewTicker(time.Second * LOG_FILE_SYNC_INTERNAL),
		}
		arrTaskAction[i].start("Action")
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
