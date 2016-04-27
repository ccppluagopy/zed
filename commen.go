package zed

import (
	"os"
	"os/exec"
	//"reflect"
	"fmt"
	"runtime"
	"strings"
)

type ClosureCB func()

type TimerCallBack func()

//type DBErrorHandler func()

type EventHandler func(event interface{}, args []interface{})

type MsgHandler func(msg *NetMsg) bool

type ClientCloseCB func(client *TcpClient)

type MysqlActionCB func(mysql *MysqlMgr) bool

func NewCoroutine(cb ClosureCB) {
	go func() {
		defer PanicHandle(true)
		cb()
	}()
}

/*
func NewCoroutine(cb interface{}, args ...interface{}) {
	f := reflect.ValueOf(cb)
	if f.Kind() == reflect.Func {
		go func() {
			defer PanicHandle(true)
			n := len(args)
			if n > 0 {
				refargs := make([]reflect.Value, n)
				for i := 0; i < n; i++ {
					refargs[i] = reflect.ValueOf(args[i])
				}
				f.Call(refargs)

			} else {
				f.Call(nil)
			}
		}()
	} else {
		ZLog("NewCoroutine Error: cb is not function")
		Println("NewCoroutine Error: cb is not function")
	}
}
*/

func GetProcName() string {
	name := os.Args[0]
	if strings.Contains(name, "/") {
		substrs := strings.Split(name, "/")
		name = substrs[len(substrs)-1]
	} else if strings.Contains(name, "\\") {
		substrs := strings.Split(name, "\\")
		name = substrs[len(substrs)-1]
	}

	return name
}

func IsProcExisting(name string) bool {
	//runtime.GOOS, runtime.GOARCH
	pid := fmt.Sprintf("%d", os.Getpid())
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	} else {
		cmd = exec.Command("ps", "-a")
	}

	buf, err := cmd.Output()
	if err != nil {
		return true
	}

	list := string(buf)
	lines := strings.Split(list, "\n")
	for _, line := range lines {
		if strings.Contains(line, name) && !strings.Contains(line, pid) {
			return true
		}
	}

	return false
}

func CheckSingleProc() {
	if IsProcExisting(GetProcName()) {
		os.Exit(0)
	}
}
