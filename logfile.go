package zed

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	logsubdir    = ""
	logdirInited = false
	mutex        sync.Mutex
	last         = time.Now().Unix()
	logfiletag   = -1

	MAX_LOG_FILE_SIZE = (1024 * 1024)
)

func SetMaxLogFileSize(size int) {
	MAX_LOG_FILE_SIZE = size
}

type logfile struct {
	tag  string
	name string
	file *os.File
	size int
}

func (logf *logfile) NewFile() bool {
	mutex.Lock()
	defer mutex.Unlock()

	logf.Close()

	/*if !logdirInited {
		MakeNewLogDir()
	}*/

	var err error

	now := time.Now().Unix()
	if now > last {
		last = now
		logfiletag = 0
	} else {
		logfiletag = logfiletag + 1
	}

	subdir := time.Now().Format("20060102/")
	//subdir := time.Now().Format("150405/")
	if logsubdir != subdir {
		logsubdir = subdir
		err = os.Mkdir(logdir+logsubdir, 0777)
		if err != nil {
			ZLog("Error when Make Log Sub Dir: %s: %v", logdir+logsubdir, err)
			return false
		} else {
			ZLog("Make Log Sub Dir: %s Success", logdir+logsubdir)
		}
	}

	//s := time.Unix(last, 0).Format("150405/")
	logf.name = logdir + logsubdir + time.Unix(now, 0).Format("150405") + fmt.Sprintf("-%d-", logfiletag) + logf.tag
	logf.file, err = os.OpenFile(logf.name, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		logf.file = nil
		ZLog("Error when Create logfile: %s: %v", logf.name, err)
		return false
	} else {
		logf.size = 0
		ZLog("Create logfile: %s: Success", logf.name)
	}

	return true
}

func (logf *logfile) Write(s *string) {
	if logf.file == nil {
		ZLog("Error when logfile %s Write, err: file is nil, str: %s", logf.name, *s)
		return
	}
	nLen := len(*s)

	if logf.size+nLen >= MAX_LOG_FILE_SIZE {
		logf.NewFile()
	}

	//subdir := time.Now().Format("150405/")
	subdir := time.Now().Format("20060102/")
	if logsubdir != subdir {
		logf.NewFile()
	}

	nWrite, err := logf.file.WriteString(*s)
	if err != nil || nWrite != nLen {
		ZLog("Error when logfile %s Write, write len: %d err: %v", logf.name, err)
	} else {
		logf.size = logf.size + nLen
	}
}

func (logf *logfile) Save() {
	if logf.file != nil {
		if err := logf.file.Sync(); err != nil {
			ZLog("Error when logfile %s Save(): %v", logf.name, err)
		}
	}
}

func (logf *logfile) Close() {
	if logf.file != nil {
		if err := logf.file.Close(); err != nil {
			ZLog("Error when logfile %s Close(): %v", logf.name, err)
		}
	}
}

func MakeDir(path string) error {
	err := os.Mkdir(path, 0777)
	return err
}

func MakeNewLogDir() bool {
	logdir = worklogdir + time.Now().Format("20060102-150405") + "/"
	err := MakeDir(logdir)
	if err != nil {
		ZLog("Error when MakeNewLogDir: %s: %v", logdir, err)
		return false
	}
	ZLog("MakeNewLogDir %s  %s Success", worklogdir, logdir)

	logdirInited = true
	return true
}

func CreateLogFile(taskType string) *logfile {
	return &logfile{
		tag:  taskType,
		name: "",
		file: nil,
		size: 0,
	}
}
