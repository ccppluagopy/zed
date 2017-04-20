package logger 

import(
"fmt"
"os"
"strings"
"sync"
"time"
)

const(
	LOG_NONE = iota
	LogCmd
	LogFile
	LogCmdFile
	LOG_MAX
)

const(
	LOG_LEVEL_NONE = iota
	LOG_LEVEL_INFO
	LOG_LEVEL_WARN
	LOG_LEVEL_ERROR
	LOG_LEVEL_ACTION
	LOG_LEVEL_MAX
)

const(
	TAG_NULL = "--"
	LOG_FILE_NAME_FORMAT = "20060102-15"
	logsep = ""
)

var(
	logmtx = sync.Mutex{}
	logdir = "./logs/"
	currlogdir = ""
	logfile *os.File = nil
	logfilename = ""
	logfilesubnum = 0
	logfilesize = 0

	maxfilesize = (1000*1000*100)
	loginfotype = LogCmd
	logwarntype = LogCmd
	logerrortype = LogCmd
	logactiontype = LogCmd

	logdebug = true
	loglevel = LOG_LEVEL_INFO
	syncinterval = time.Second * 30

	Printf = fmt.Printf
	Sprintf = fmt.Sprintf
	Println = fmt.Println

	LOG_IDX = 0
	LOG_TAG = "zlog"

	logtags = map[int]string{
		LOG_IDX: LOG_TAG,
	}

	logconf = map[string]int{	
		"Info": LogCmd,
		"Warn": LogCmd,
		"Error": LogCmd,
		"Action": LogCmd,
	}

	logtimer *time.Timer = nil
	chsynclogfile chan struct{} = nil
)

func NewFile(path string) (*os.File, error){
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		Println("zlog NewFile Error: %s, %s", path, err.Error())
		return nil err
	}
	return file, err
}

func checkFile() bool {
	var err error = nil
	currfilename := Sprintf("%s-%d", time.Now().Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
	if logfilename != currfilename {
		logfilesubnum = 0
		logfilename = Sprintf("%s-%d", time.Now().Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
		if logfile != nil{
			logfile.Close()
		}
		logfile, err = NewFile(currlogdir + logfilename)
		if err != nil {
			Println("zlog checkFile Failed")
			return false
		}else{
			logfilesize = 0
			return true
		}
	}

	if logfilesize > maxfilesize {
		if logfilename == currfilename {
			logfilesubnum++
			logfilename = Sprintf("%s-%d", time.Now().Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
		}else{
			logfilesubnum = 0
			logfilename = Sprintf("%s-%d", time.Now().Format(LOG_FILE_NAME_FORMAT), logfilesubnum)
		}
		if logfile != nil{
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
	}else{
		logdir = dir + "/"
	}
}

func SetLogLevel(level int) {
	if level > LOG_LEVEL_NONE && level < LOGLEVEL_MAX {
		loglevel = level
	}else{
		Printf("zlog SetLogLevel Error: Invalid Level - %d\n", level)
	}
}

func SetMaxLogFileSize(size int) {
	if size > 0 {
		maxfilesize = size
	}else{
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

func writetofile(str string){
	logmtx.Lock()
	defer logmtx.Unlock()

	checkFile()
	n, err := logfile.WriteString(str)
	if err != nil || n != len(str) {
		Printf("zlog writetofile Failed: %d/%d wrote, Error: %v", n, len(str), err)
	}else{
		logfilesize += len(str)
	}
}

func syncLogFile() {
	logmtx.Lock()
	defer logmtx.Unlock()

	logfile.Sync()
}

func LogInfo(tag int, format string, v ...interface{}) {
	if LOG_LEVEL_INFO >= loglevel {
		if tagstr, ok := logtags[tag]; ok && (tagstr[0:len(TAG_NULL)] != TAG_NULL){
			s:= strings.Join([]string{time.Now().Format("2006-01-02 15:04:05.000"), " [", tagstr, "] [  INFO] ", Sprintf(format, v...), "\n", logsep)
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















