package zed

import (
	//"github.com/ccppluagopy/zed"
	//"reflect"
	"time"
)

//import ()

type LogConfig interface {
	GetDebug() bool
	GetMaskSignal() bool
	GetWorkDir() string
	GetLogDir() string
	GetLogLevel() int
	GetLogConf() map[string]int
	GetLogTags() map[int]string
	GetAliveLogTime() int
	GetAliveLogStr() string
}

type DefaultLogConfig struct {
	Debug      bool
	MaskSignal bool

	WorkDir      string
	LogDir       string
	LogLevel     int
	LogConf      map[string]int
	LogTags      map[int]string
	AliveLogTime int
	AliveLogStr  string
}

func (conf *DefaultLogConfig) GetDebug() bool {
	return conf.Debug
}
func (conf *DefaultLogConfig) GetMaskSignal() bool {
	return conf.MaskSignal
}
func (conf *DefaultLogConfig) GetWorkDir() string {
	return conf.WorkDir
}
func (conf *DefaultLogConfig) GetLogDir() string {
	return conf.LogDir
}
func (conf *DefaultLogConfig) GetLogLevel() int {
	return conf.LogLevel
}
func (conf *DefaultLogConfig) GetLogConf() map[string]int {
	return conf.LogConf
}
func (conf *DefaultLogConfig) GetLogTags() map[int]string {
	return conf.LogTags
}
func (conf *DefaultLogConfig) GetAliveLogTime() int {
	return conf.AliveLogTime
}
func (conf *DefaultLogConfig) GetAliveLogStr() string {
	return conf.AliveLogStr
}

func InitLog(logconf LogConfig) {
	//logconf, ok := conf.(*LogConfig)
	//if ok {
	ZLog("InitLog()")
	/*	sendInValue := reflect.ValueOf(pnew).Elem()
		dbValue := reflect.ValueOf(*pold)
		typeof := dbValue.Type()
		for i := 0; i < sendInValue.NumField(); i++ {
			if sendInValue.Field(i).String() != dbValue.Field(i).String() {
				fieldName := strings.ToLower(string(typeof.Field(i).Name))
				changeMap[fieldName] = sendInValue.Field(i).Interface()
			}
		}
	*/
	Init(logconf.GetWorkDir(), logconf.GetLogDir())
	SetLogLevel(logconf.GetLogLevel())
	StartLogger(logconf.GetLogConf(), logconf.GetLogTags(), logconf.GetDebug())
	StartAliveLog(logconf.GetAliveLogTime(), logconf.GetAliveLogStr())
	/*} else {
		zed.ZLog("InitLog Error: Invalid config.")
	}*/
}

func StartAliveLog(internal int, logstr string) {
	ZLog("StartAliveLog: %ds - %s", internal, logstr)
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(internal))
			LogAction(0, 0, logstr)
			LogActionSave(0)
		}
	}()
}
