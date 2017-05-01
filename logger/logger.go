package logger

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	LOG_NONE = iota
	LogCmd
	LogFile
	LogCmdFile
	LOG_MAX
)

const (
	LOG_LEVEL_NONE = iota
	LOG_LEVEL_INFO
	LOG_LEVEL_WARN
	LOG_LEVEL_ERROR
	LOG_LEVEL_ACTION
	LOG_LEVEL_MAX
)

const (
	TAG_NULL             = "--"
	LOG_FILE_NAME_FORMAT = "20060102-15"
	LOG_STR_FORMAT       = "2006-01-02 15:04:05.000"
	logsep               = ""
)

var (
	logmtx                      = sync.Mutex{}
	logdir                      = "./logs/"
	currlogdir                  = ""
	logfile       *os.File      = nil
	filewriter    *bufio.Writer = nil
	logfilename                 = ""
	logfilesubnum               = 0
	logfilesize                 = 0

	maxfilesize   = (1024 * 1024 * 32)
	loginfotype   = LogCmd
	logwarntype   = LogCmd
	logerrortype  = LogCmd
	logactiontype = LogCmd

	logdebug     = true
	loglevel     = LOG_LEVEL_INFO
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
		"Info":   LogCmd,
		"Warn":   LogCmd,
		"Error":  LogCmd,
		"Action": LogCmd,
	}

	logtimer *time.Timer = nil
	//chsynclogfile chan struct{} = nil
)

func NewFile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		Println("zlog NewFile Error: %s, %s", path, err.Error())
		return nil, err
	}
	if filewriter == nil {
		filewriter = bufio.NewWriter(file)
	} else {
		filewriter.Reset(file)
	}

	return file, err
}

func checkFile() bool {
	var err error = nil
	currfilename := Sprintf("%s-%d", time.Now().Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
	if logfilename != currfilename {
		logfilesubnum = 0
		logfilename = Sprintf("%s-%d", time.Now().Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
		if logfile != nil {
			logfile.Close()
		}
		logfile, err = NewFile(currlogdir + logfilename)
		if err != nil {
			Println("zlog checkFile Failed")
			return false
		} else {
			logfilesize = 0
			return true
		}
	}

	if logfilesize > maxfilesize {
		if logfilename == currfilename {
			logfilesubnum++
			logfilename = Sprintf("%s-%d", time.Now().Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
		} else {
			logfilesubnum = 0
			logfilename = Sprintf("%s-%d", time.Now().Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
		}
		if logfile != nil {
			logfile.Close()
		}
		logfile, err = NewFile(currlogdir + logfilename)
		if err != nil {
			Println("zlog checkFile Failed")
			return false
		}
		logfilesize = 0
	}
	return true
}

func MakeDir(path string) error {
	err := os.MkdirAll(path, 0777)
	return err
}

func SetLogDir(dir string) {
	if strings.HasSuffix(dir, "/") || strings.HasSuffix(dir, "\\") {
		logdir = dir
	} else {
		logdir = dir + "/"
	}
}

func SetLogLevel(level int) {
	if level > LOG_LEVEL_NONE && level < LOG_LEVEL_MAX {
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
	currlogdir = logdir + time.Now().Format("20060102-150405/")
	err := MakeDir(currlogdir)
	if err != nil {
		Printf("zlog initLogDirAndFile Error: %s-%v\n", currlogdir, err)
		return false
	}
	return true
}

func writetofile(str string) {
	logmtx.Lock()
	defer logmtx.Unlock()

	checkFile()

	n, err := filewriter.WriteString(str)
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

	filewriter.Flush()
	//logfile.Sync()
}

func LogInfo(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_INFO >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
			s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), " [", tagstr, "] [  INFO] ", Sprintf(format, v...), "\n"}, logsep)
			switch loginfotype {
			case LogFile:
				writetofile(s)
				break
			case LogCmd:
				Printf(s)
				break
			case LogCmdFile:
				Printf(s)
				writetofile(s)
				break
			default:
				break
			}
		}
	}
}

func LogWarn(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_WARN >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
			s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), " [", tagstr, "] [  WARN] ", Sprintf(format, v...), "\n"}, logsep)
			switch logwarntype {
			case LogFile:
				writetofile(s)
				break
			case LogCmd:
				Printf(s)
				break
			case LogCmdFile:
				Printf(s)
				writetofile(s)
				break
			default:
				break
			}
		}
	}
}

func LogError(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_ERROR >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
			s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), " [", tagstr, "] [ ERROR] ", Sprintf(format, v...), "\n"}, logsep)
			switch logerrortype {
			case LogFile:
				writetofile(s)
				break
			case LogCmd:
				Printf(s)
				break
			case LogCmdFile:
				Printf(s)
				writetofile(s)
				break
			default:
				break
			}
		}
	}
}

func LogAction(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_ACTION >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL) {
			s := strings.Join([]string{time.Now().Format(LOG_STR_FORMAT), " [", tagstr, "] [ACTION] ", Sprintf(format, v...), "\n"}, logsep)
			switch logactiontype {
			case LogFile:
				writetofile(s)
				break
			case LogCmd:
				Printf(s)
				break
			case LogCmdFile:
				Printf(s)
				writetofile(s)
				break
			default:
				break
			}
		}
	}
}

func Save() {
	syncLogFile()
}

func startSync() {
	logmtx.Lock()
	defer logmtx.Unlock()

	if logfile == nil { //|| chsynclogfile != nil {
		return
	}

	//chsynclogfile = make(chan struct{})

	go func() {
		defer func() {
			recover()
		}()

		logtimer = time.NewTimer(syncinterval)

		for {
			//select {
			//case _, ok := <-logtimer.C:
			_, ok := <-logtimer.C
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
			logtimer.Reset(syncinterval)
		}
	}()
}

func StartLogger(conf map[string]int, tags map[int]string) {
	Println("===========================================")
	Println("Logger Start")
	if conf == nil {
		conf = logconf
	}
	if conf != nil {
		Println("logconf:")
		if lt, ok := conf["Info"]; !ok || lt <= LOG_NONE || lt >= LOG_MAX {
			Println("StartLogger Error: Invalid Info LogType")
			return
		} else {
			str := "LogCmd"
			if lt == LogFile {
				str = "LogFile"
			} else if lt == LogCmdFile {
				str = "LogCmdFile"
			}
			loginfotype = lt
			Println("	", "Info	", str)
		}
		if lt, ok := conf["Warn"]; !ok || lt <= LOG_NONE || lt >= LOG_MAX {
			Println("StartLogger Error: Invalid Warn LogType")
			return
		} else {
			str := "LogCmd"
			if lt == LogFile {
				str = "LogFile"
			} else if lt == LogCmdFile {
				str = "LogCmdFile"
			}
			logwarntype = lt
			Println("	", "Warn	", str)
		}
		if lt, ok := conf["Error"]; !ok || lt <= LOG_NONE || lt >= LOG_MAX {
			Println("StartLogger Error: Invalid Error LogType")
			return
		} else {
			str := "LogCmd"
			if lt == LogFile {
				str = "LogFile"
			} else if lt == LogCmdFile {
				str = "LogCmdFile"
			}
			logerrortype = lt
			Println("	", "Error	", str)
		}
		if lt, ok := conf["Action"]; !ok || lt <= LOG_NONE || lt >= LOG_MAX {
			Println("StartLogger Error: Invalid Action LogType")
			return
		} else {
			str := "LogCmd"
			if lt == LogFile {
				str = "LogFile"
			} else if lt == LogCmdFile {
				str = "LogCmdFile"
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
					for i := 0; i < maxtaglen; i++ {
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

	startSync()

	Println("===========================================")
}

func StopLogger() {
	if logtimer != nil {
		logtimer.Stop()
		logtimer = nil
	}
	/*if chsynclogfile != nil {
		close(chsynclogfile)
		chsynclogfile = nil
	}*/
	Println("Logger Stop!")
}
