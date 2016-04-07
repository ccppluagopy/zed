package zed

import (
	"sync"
	"time"
)

var (
	mutex sync.Mutex
	last  = time.Now().Unix()
)

type logfile struct {
	name string
	file *os.File
	size uint32
}

func (logf *logfile) NewFile() {
	mutex.Lock()
	defer mutex.Unlock()

	var err error

	now := time.Now().Unix()
	if now != last {
		last = now
	} else {
		last = last + 1
	}

	logf.name = time.Unix(last, 0).Format("20060102-150405")

	logf.file, err = os.OpenFile(logf.name, os.O_APPEND, 0666)
	if err != nil {
		logf.file = nil
		ZLog("Error when Create logfile: %s: %v.", logf.name, err)
	}
}

func (logf *logfile) Write(s *string) {
	if logf.file == nil {
		ZLog("Error when logfile %s Write, err: file is nil.", logf.name)
		return
	}
	nLen := len(*s)
	nWrite, err := logf.file.WriteString(*s)
	if err != nil || nWrite != nLen {
		ZLog("Error when logfile %s Write, write len: %d err: %v.", logf.name, n, err)
	} else {
		logf.size = logf.size + nLen

		if logf.size >= MAX_LOG_FILE_SIZE {
			logf.Close()
			logf.NewFile()
		}
	}

}

func (logf *logfile) Save() {
	if logf.file != nil {
		if err := logf.file.Sync(); err != nil {
			ZLog("Error when logfile %s Save(): %v.", logf.name, err)
		}
	}
}

func (logf *logfile) Close() {
	if logf.file != nil {
		var err error
		if err = logf.file.Sync(); err != nil {
			ZLog("Error when logfile %s Sync() before Close(): %v.", logf.name, err)
		}

		if err = logf.file.Close(); err != nil {
			ZLog("Error when logfile %s Close(): %v.", logf.name, err)
		}
	}
}

func NewLogFile() *logfile {
	return &logfile{
		name: "",
		file: nil,
		size: 0,
	}
}
