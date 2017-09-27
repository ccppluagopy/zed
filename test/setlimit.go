package main

import (
	"fmt"
	"syscall"
)

func main() {
	var rLimit syscall.Rlimit
	rLimit.Max = 65535 //100000
	rLimit.Cur = 65535 //100000
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		fmt.Printf("setUlimit(ulimit -n 65535) failed: %v\n", err)
	} else {
		fmt.Printf("setUlimit(ulimit -n 65535) success\n")
	}

	rLimit.Max = 0
	rLimit.Cur = 0
	rLimit.Max -= 1
	rLimit.Cur -= 1
	if err := syscall.Setrlimit(syscall.RLIMIT_CORE, &rLimit); err != nil {
		fmt.Printf("setUlimit(ulimit -c unlimited) failed: %v\n", err)
	} else {
		fmt.Printf("setUlimit(ulimit -c unlimited) success\n")
	}
}
