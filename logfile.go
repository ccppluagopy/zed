package zed

import (
	//"fmt"
	"os"
	"sync"
	"time"
)

var (
	logdir       = "./"
	logdirInited = false
	mutex        sync.Mutex
	last         = time.Now().Unix()
)

type logfile struct {
	tag  string
	name string
	file *os.File
	size int
}

func (logf *logfile) NewFile() bool {
	mutex.Lock()
	defer mutex.Unlock()

	if !logdirInited {
		MakeNewLogDir(logdir)
	}

	var err error

	now := time.Now().Unix()
	if now > last {
		last = now
	} else {
		last = last + 1
	}

	logf.name = logdir + time.Unix(last, 0).Format("20060102-150405") + "-" + logf.tag

	logf.file, err = os.OpenFile(logf.name, os.O_CREATE, 0666)
	if err != nil {
		logf.file = nil
		LogError(LOG_IDX, LOG_IDX, "Error when Create logfile: %s %s: %v.", logdir, logf.name, err)
		return false
	} else {
		logf.size = 0
		//LogInfo(LOG_IDX, LOG_IDX, "Create logfile: %s: Success.  %d", logf.name, last)
	}

	return true
}

func (logf *logfile) Write(s *string) {
	if logf.file == nil {
		LogError(LOG_IDX, LOG_IDX, "Error when logfile %s Write, err: file is nil.", logf.name)
		return
	}
	nLen := len(*s)

	if logf.size+nLen >= MAX_LOG_FILE_SIZE {
		logf.Close()
		logf.NewFile()
	}

	nWrite, err := logf.file.WriteString(*s)
	if err != nil || nWrite != nLen {
		LogError(LOG_IDX, LOG_IDX, "Error when logfile %s Write, write len: %d err: %v.", logf.name, err)
	} else {
		logf.size = logf.size + nLen
	}

}

/*func (logf *logfile) Write(s *string) {
	if logf.file == nil {
		LogError(LOG_IDX, LOG_IDX, "Error when logfile %s Write, err: file is nil.", logf.name)
		return
	}
	nLen := len(*s)
	nWrite, err := logf.file.WriteString(*s)
	if err != nil || nWrite != nLen {
		LogError(LOG_IDX, LOG_IDX, "Error when logfile %s Write, write len: %d err: %v.", logf.name, err)
	} else {
		logf.size = logf.size + nLen

		if logf.size >= MAX_LOG_FILE_SIZE {
			logf.Close()
			logf.NewFile()
		}
	}

}*/

func (logf *logfile) Save() {
	if logf.file != nil {
		if err := logf.file.Sync(); err != nil {
			LogError(LOG_IDX, LOG_IDX, "Error when logfile %s Save(): %v.", logf.name, err)
		}
	}
}

func (logf *logfile) Close() {
	if logf.file != nil {
		var err error
		/*if err = logf.file.Sync(); err != nil {
			LogError(LOG_IDX, LOG_IDX, "Error when logfile %s Sync() before Close(): %v.", logf.name, err)
		}
		*/
		if err = logf.file.Close(); err != nil {
			LogError(LOG_IDX, LOG_IDX, "Error when logfile %s Close(): %v.", logf.name, err)
		}
	}
}

func MakeNewLogDir(parentDir string) {
	logdir = parentDir + time.Now().Format("20060102-150405") + "/"
	err := os.Mkdir(logdir, 0777)
	if err != nil {
		ZLog("Error when MakeNewLogDir: %s: %v.", logdir, err)
	} else {
		ZLog("MakeNewLogDir: %s Success", logdir)
		logdirInited = true
	}
}

func CreateLogFile(taskType string) *logfile {
	return &logfile{
		tag:  taskType,
		name: "",
		file: nil,
		size: 0,
	}
}
