package logger

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	//LOG_NONE = iota
	LogCmd  = 0x1 << 0
	LogFile = 0x1 << 1
	LogUser = 0x1 << 3
	//LOG_MAX
)

const (
	LOG_LEVEL_DEBUG = iota
	LOG_LEVEL_INFO
	LOG_LEVEL_WARN
	LOG_LEVEL_ERROR
	LOG_LEVEL_ACTION
	LOG_LEVEL_NONE
	LOG_LEVEL_MAX
)

const (
	TAG_NULL               = "--"
	LOG_DIR_NAME_FORMAT    = "20060102-150405/"
	LOG_SUBDIR_NAME_FORMAT = "20060102/"
	LOG_FILE_NAME_FORMAT   = "20060102-15"
	LOG_SUFIX              = ".log"
	/*LOG_DIR_NAME_FORMAT    = "20060102-150405/"
	LOG_SUBDIR_NAME_FORMAT = "20060102-150405/"
	LOG_FILE_NAME_FORMAT   = "20060102-150405"*/
	LOG_STR_FORMAT = "2006-01-02 15:04:05.000"
	logsep         = ""
	callerdepth    = 1
)

var (
	logmtx                      = sync.Mutex{}
	logdir                      = "./logs/"
	logdirinited                = false
	currlogdir                  = ""
	logfile       *os.File      = nil
	filewriter    *bufio.Writer = nil
	logfilename                 = ""
	logfilesubnum               = 0
	logfilesize                 = 0

	maxfilesize   = (1024 * 1024 * 10)
	logdebugtype  = LogCmd
	loginfotype   = LogCmd
	logwarntype   = LogCmd
	logerrortype  = LogCmd
	logactiontype = LogCmd

	//logdebug     = true
	loglevel     = LOG_LEVEL_DEBUG
	syncinterval = time.Second * 30

	Printf  = fmt.Printf
	Sprintf = fmt.Sprintf
	Println = fmt.Println

	LOG_IDX = 0
	LOG_TAG = "zlog"

	logtags = map[int]string{
		LOG_IDX: LOG_TAG,
	}

	logconf = map[string]int{
		"Debug":  LogCmd,
		"Info":   LogCmd,
		"Warn":   LogCmd,
		"Error":  LogCmd,
		"Action": LogCmd,
	}

	logticker    *time.Ticker = nil
	enablebufio               = false
	enableSubdir              = false
	//chsynclogfile chan struct{} = nil
	userlogger = func(str string) {}
	inittime   = time.Now()

	initsync = false
)

func newFile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		Printf("zlog newFile Error: %s, %s\n", path, err.Error())
		return nil, err
	}

	if enablebufio {
		if filewriter == nil {
			filewriter = bufio.NewWriter(file)
		} else {
			filewriter.Reset(file)
		}
	}

	return file, err
}

func checkFile() bool {
	var (
		err error = nil
		now       = time.Now()
	)

	currfilename := Sprintf("%s-%d", now.Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
	if logfilename != currfilename {
		logfilesubnum = 0
		logfilename = Sprintf("%s-%d", now.Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
		if logfile != nil {
			logfile.Close()
		}
		if enableSubdir {
			logdir := logdir + inittime.Format(LOG_DIR_NAME_FORMAT) + now.Format(LOG_SUBDIR_NAME_FORMAT)
			if logdir != currlogdir {
				currlogdir = logdir
				if err := MakeDir(currlogdir); err != nil {
					Println("zlog checkFile Failed")
					return false
				}
			}
		}
		logfile, err = newFile(currlogdir + logfilename + LOG_SUFIX)
		if err != nil {
			Println("zlog checkFile Failed")
			return false
		} else {
			logfilesize = 0
			/*if logticker != nil {
				logticker.Reset(syncinterval)
			}*/
			return true
		}
	}

	if logfilesize > maxfilesize {
		if logfilename == currfilename {
			logfilesubnum++
			logfilename = Sprintf("%s-%d", now.Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
		} else {
			logfilesubnum = 0
			logfilename = Sprintf("%s-%d", now.Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
		}
		if logfile != nil {
			logfile.Close()
		}
		if enableSubdir {
			logdir := logdir + inittime.Format(LOG_DIR_NAME_FORMAT) + now.Format(LOG_SUBDIR_NAME_FORMAT)
			if logdir != currlogdir {
				currlogdir = logdir
				if err := MakeDir(currlogdir); err != nil {
					Println("zlog checkFile Failed")
					return false
				}
			}
		}
		logfile, err = newFile(currlogdir + logfilename + LOG_SUFIX)
		if err != nil {
			Println("zlog checkFile Failed")
			return false
		}
		logfilesize = 0
		/*if logticker != nil {
			logticker.Reset(syncinterval)
		}*/
	}

	if !initsync {
		startSync()
	}

	return true
}

func MakeDir(path string) error {
	err := os.MkdirAll(path, 0777)
	return err
}

func SetLogDir(dir string, args ...interface{}) {
	if strings.HasSuffix(dir, "/") || strings.HasSuffix(dir, "\\") {
		logdir = dir
	} else {
		logdir = dir + "/"
	}
	if len(args) > 0 {
		if es, ok := args[0].(bool); ok {
			enableSubdir = es
		}
	}
}

func SetLogUser(w func(str string)) {
	userlogger = w
}

func SetMaxFileSize(size int) {
	if size > 0 {
		maxfilesize = size
	} else {
		Printf("zlog SetMaxFileSize Error: size(%d) <= 0\n", size)
	}

}
func SetLogLevel(level int) {
	if level >= 0 && level < LOG_LEVEL_MAX {
		loglevel = level
	} else {
		Printf("zlog SetLogLevel Error: Invalid Level - %d\n", level)
	}
}

func SetMaxLogFileSize(size int) {
	if size > 0 {
		maxfilesize = size
	} else {
		Printf("zlog SetMaxLogFileSize Error: Invalid size - %d\n", size)
	}
}

func setSyncLogFileInterval(interval time.Duration) {
	syncinterval = interval
}

func initLogDirAndFile() bool {
	if !logdirinited {
		logdirinited = true
		//currlogdir = logdir + time.Now().Format("20060102-150405/")
		inittime := time.Now()
		currlogdir = logdir + inittime.Format(LOG_DIR_NAME_FORMAT)
		if enableSubdir {
			currlogdir += inittime.Format(LOG_SUBDIR_NAME_FORMAT)
		}
		err := MakeDir(currlogdir)
		if err != nil {
			Printf("zlog initLogDirAndFile Error: %s-%v\n", currlogdir, err)
			return false
		}
	}
	return true
}

func writetofile(str string) {
	logmtx.Lock()
	defer logmtx.Unlock()

	checkFile()
	//return
	var (
		n   int
		err error
	)
	if enablebufio {
		n, err = filewriter.WriteString(str)
	} else {
		n, err = logfile.WriteString(str)
	}
	//Println(n, err, str)
	if err != nil || n != len(str) {
		Printf("zlog writetofile Failed: %d/%d wrote, Error: %v", n, len(str), err)
	} else {
		logfilesize += len(str)
	}
}

func syncLogFile() {
	logmtx.Lock()
	defer logmtx.Unlock()

	if enablebufio {
		if filewriter != nil {
			filewriter.Flush()
		}
	} else {
		if logfile != nil {
			logfile.Sync()
		}
	}
	//logfile.Sync()
}

/*func Debug(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_DEBUG >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
			s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), " [", tagstr, "] [  INFO] ", Sprintf(format, v...), "\n"}, logsep)
			if logdebugtype&LogFile != 0 {
				writetofile(s)
			}
			if logdebugtype&LogCmd != 0 {
				Printf(s)
			}
			if loginfotype&LogUser != 0 {
				userlogger(s)
			}
		}
	}
}

func Info(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_INFO >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
			s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), " [", tagstr, "] [  INFO] ", Sprintf(format, v...), "\n"}, logsep)
			if loginfotype&LogFile != 0 {
				writetofile(s)
			}
			if loginfotype&LogCmd != 0 {
				Printf(s)
			}
			if loginfotype&LogUser != 0 {
				userlogger(s)
			}
		}
	}
}

func Warn(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_WARN >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
			s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), " [", tagstr, "] [  WARN] ", Sprintf(format, v...), "\n"}, logsep)
			if logwarntype&LogFile != 0 {
				writetofile(s)
			}
			if logwarntype&LogCmd != 0 {
				Printf(s)
			}
			if loginfotype&LogUser != 0 {
				userlogger(s)
			}
		}
	}
}

func Error(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_ERROR >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
			s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), " [", tagstr, "] [ ERROR] ", Sprintf(format, v...), "\n"}, logsep)
			if logerrortype&LogFile != 0 {
				writetofile(s)
			}
			if logerrortype&LogCmd != 0 {
				Printf(s)
			}
			if loginfotype&LogUser != 0 {
				userlogger(s)
			}
		}
	}
}

func Action(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_ACTION >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
			s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), " [", tagstr, "] [ACTION] ", Sprintf(format, v...), "\n"}, logsep)
			if logactiontype&LogFile != 0 {
				writetofile(s)
			}
			if logactiontype&LogCmd != 0 {
				Printf(s)
			}
			if loginfotype&LogUser != 0 {
				userlogger(s)
			}
		}
	}
}*/

func Debug(format string, v ...interface{}) {
	if LOG_LEVEL_DEBUG >= loglevel {
		_, file, line, ok := runtime.Caller(callerdepth)
		if !ok {
			file = "???"
			line = -1
		} else {
			pos := strings.LastIndex(file, "/")
			if pos >= 0 {
				file = file[pos+1:]
			}
		}
		s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), Sprintf(" [ Debug] [%s:%d]", file, line), Sprintf(format, v...), "\n"}, logsep)
		if logdebugtype&LogFile != 0 {
			writetofile(s)
		}
		if logdebugtype&LogCmd != 0 {
			Printf(s)
		}
		if logdebugtype&LogUser != 0 {
			userlogger(s)
		}
	}
}

func Info(format string, v ...interface{}) {
	if LOG_LEVEL_INFO >= loglevel {
		_, file, line, ok := runtime.Caller(callerdepth)
		if !ok {
			file = "???"
		} else {
			pos := strings.LastIndex(file, "/")
			if pos >= 0 {
				file = file[pos+1:]
			}
		}
		s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), Sprintf(" [ Debug] [%s:%d]", file, line), Sprintf(format, v...), "\n"}, logsep)
		if loginfotype&LogFile != 0 {
			writetofile(s)
		}
		if loginfotype&LogCmd != 0 {
			Printf(s)
		}
		if loginfotype&LogUser != 0 {
			userlogger(s)
		}
	}
}

func Warn(format string, v ...interface{}) {
	if LOG_LEVEL_WARN >= loglevel {
		_, file, line, ok := runtime.Caller(callerdepth)
		if !ok {
			file = "???"
		} else {
			pos := strings.LastIndex(file, "/")
			if pos >= 0 {
				file = file[pos+1:]
			}
		}
		s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), Sprintf(" [ Debug] [%s:%d]", file, line), Sprintf(format, v...), "\n"}, logsep)
		if logwarntype&LogFile != 0 {
			writetofile(s)
		}
		if logwarntype&LogCmd != 0 {
			Printf(s)
		}
		if logwarntype&LogUser != 0 {
			userlogger(s)
		}
	}
}

func Action(format string, v ...interface{}) {
	if LOG_LEVEL_ACTION >= loglevel {
		_, file, line, ok := runtime.Caller(callerdepth)
		if !ok {
			file = "???"
		} else {
			pos := strings.LastIndex(file, "/")
			if pos >= 0 {
				file = file[pos+1:]
			}
		}
		s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), Sprintf(" [ Debug] [%s:%d]", file, line), Sprintf(format, v...), "\n"}, logsep)
		if logactiontype&LogFile != 0 {
			writetofile(s)
		}
		if logactiontype&LogCmd != 0 {
			Printf(s)
		}
		if logactiontype&LogUser != 0 {
			userlogger(s)
		}
	}
}

func SetLogDebugType(t int) {
	logdebugtype = t
	if t&LogFile == LogFile {
		initLogDirAndFile()
	}
}

func SetLogInfoType(t int) {
	loginfotype = t
	if t&LogFile == LogFile {
		initLogDirAndFile()
	}
}

func SetLogWarnType(t int) {
	logwarntype = t
	if t&LogFile == LogFile {
		initLogDirAndFile()
	}
}

func SetLogErrorType(t int) {
	logerrortype = t
	if t&LogFile == LogFile {
		initLogDirAndFile()
	}
}

func SetLogType(t int) {
	logdebugtype = t
	loginfotype = t
	logwarntype = t
	logerrortype = t
	if t&LogFile == LogFile {
		initLogDirAndFile()
	}
}

func Save() {
	syncLogFile()
}

func startSync() {
	if !initsync {
		initsync = true
		go func() {
			defer func() {
				recover()
			}()

			logticker = time.NewTicker(syncinterval)

			for {
				//select {
				//case _, ok := <-logticker.C:
				_, ok := <-logticker.C
				syncLogFile()
				if !ok {
					return
				}
				/*case _, ok := <-chsynclogfile:
					syncLogFile()
					if !ok {
						return
					}
				}*/
				//logticker.Reset(syncinterval)
			}
		}()
	}
}

func StartLogger(conf map[string]int, tags map[int]string, args ...interface{}) {
	Println("===========================================")
	Println("Logger Start")
	if len(args) == 1 {
		if enable, ok := args[0].(bool); ok {
			enablebufio = enable
		}
	}
	if conf == nil {
		conf = logconf
	}
	if conf != nil {
		Println("logconf:")
		if lt, ok := conf["Debug"]; ok {
			str := ""
			if lt&LogCmd != 0 {
				str += "Cmd "
			}
			if lt&LogFile != 0 {
				str += "File"
			}
			logdebugtype = lt
			Println("	", "Debug	", str)
		}
		if lt, ok := conf["Info"]; ok {
			str := ""
			if lt&LogCmd != 0 {
				str += "Cmd "
			}
			if lt&LogFile != 0 {
				str += "File"
			}
			loginfotype = lt
			Println("	", "Info	", str)
		}
		if lt, ok := conf["Warn"]; ok {
			str := ""
			if lt&LogCmd != 0 {
				str += "Cmd "
			}
			if lt&LogFile != 0 {
				str += "File"
			}
			logwarntype = lt
			Println("	", "Warn	", str)
		}
		if lt, ok := conf["Error"]; ok {
			str := ""
			if lt&LogCmd != 0 {
				str += "Cmd "
			}
			if lt&LogFile != 0 {
				str += "File"
			}
			logerrortype = lt
			Println("	", "Error	", str)
		}
		if lt, ok := conf["Action"]; ok {
			str := ""
			if lt&LogCmd != 0 {
				str += "Cmd "
			}
			if lt&LogFile != 0 {
				str += "File"
			}
			logactiontype = lt
			Println("	", "Action	", str)
		}
		logconf = conf
	}
	if tags != nil && len(tags) > 0 {
		maxtaglen := 0
		for _, v := range tags {
			if v[:len(TAG_NULL)] != TAG_NULL {
				if len(v) > maxtaglen {
					maxtaglen = len(v)
				}
			}
		}
		maxtaglen = maxtaglen + 1
		for k, v := range tags {
			if v[:len(TAG_NULL)] != TAG_NULL {
				if len(v) < maxtaglen {
					newv := v
					for i := 0; i < maxtaglen-len(v); i++ {
						newv = " " + newv
					}
					tags[k] = newv
				}
			}
		}
		logtags = tags
	}
	Println(" - - - - - - - - - - - - - - - - - - - - - - - - - -")
	Println("logtags:")
	for i := 0; ; i++ {
		if tag, ok := logtags[i]; ok {
			Println("	", i, tag)
		} else {
			break
		}
	}
	Println(" - - - - - - - - - - - - - - - - - - - - - - - - - -")

	if !initLogDirAndFile() {
		Println("StartLogger Failed")
	} else {
		Printf("initLogDirAndFile %s Success\n", currlogdir)
	}

	checkFile()

	Println("===========================================")
}

func StopLogger() {
	if logticker != nil {
		logticker.Stop()
		logticker = nil
	}
	/*if chsynclogfile != nil {
		close(chsynclogfile)
		chsynclogfile = nil
	}*/
	syncLogFile()
	Println("Logger Stop!")
}
