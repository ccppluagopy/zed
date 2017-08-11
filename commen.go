package zed

import (
	"os"
	"os/exec"
	//"reflect"
	"fmt"
	//_ "github.com/go-sql-driver/mysql"
	"net"
	"runtime"
	"strings"
	//"time"
)

func Async(cb func(), args ...interface{}) {
	//var t *time.Timer
	//t = time.AfterFunc(1, func() {
	go func() {
		if len(args) == 1 {
			if panichandler, ok := args[0].(func()); ok {
				defer panichandler()
			} else {
				defer HandlePanic(true)
			}
		} else {
			defer HandlePanic(true)
		}
		cb()
	}()
		//t.Stop()
	//})
}

func NewCoroutine(cb ClosureCB) {
	go func() {
		defer HandlePanic(true)
		cb()
	}()
}

/*
func NewCoroutine(cb interface{}, args ...interface{}) {
	f := reflect.ValueOf(cb)
	if f.Kind() == reflect.Func {
		go func() {
			defer HandlePanic(true)
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

func GetLocalAddr() ([]string, error) {
	var ret []string

	addrs, err := net.InterfaceAddrs()

	if err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ret = append(ret, ipnet.IP.String())
				}
			}
		}
	}

	return ret, err
}

/*func GetPublicAddr() {
	conn, err := net.Dial("udp", "www.baidu.com:80")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer conn.Close()
	fmt.Println(conn.LocalAddr().String())
}*/
