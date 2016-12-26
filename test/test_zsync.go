package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"time"
)

func main() {

	mt := zed.Mutex{}
	zed.SetMutexDebug(true, time.Second*3)
	fmt.Println("111")
	mt.Lock()
	fmt.Println("222")
	mt.Lock()
	fmt.Println("333")
	time.Sleep(time.Second * 5)
}
