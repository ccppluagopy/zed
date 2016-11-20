package main

import (
	"fmt"
	"github.com/ccppluagopy/zed/zsync"
	"time"
)

func main() {
	//zed.NewTcpServer("xx").Start("192.168.1.15:2222")
	go func() {
		time.Sleep(time.Second)
		mt := zsync.RWMutex{}
		zsync.SetDebug(true)
		fmt.Println("R 111")
		mt.Lock()
		fmt.Println("R 222")
		mt.RLock()
		fmt.Println("R 333")
	}()

	mt := zsync.Mutex{}
	zsync.SetDebug(true)
	fmt.Println("111")
	mt.Lock()
	fmt.Println("222")
	mt.Lock()
	fmt.Println("333")
}
