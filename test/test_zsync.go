package main

import (
	"fmt"
	//"github.com/ccppluagopy/zed"
	"github.com/ccppluagopy/zed/zsync"
	"time"
)

type xx struct {
}

var (
	xxmap = map[int]*xx{}
)

func main() {
	/*tw := zed.NewTimerWheel(time.Second, 5)
	tw.NewTimer("1", time.Second, func(t *zed.WTimer) {
		fmt.Println("xx 111")
	}, 0)
	tw.NewTimer("1", time.Second, func(t *zed.WTimer) {
		fmt.Println("xx 222")
	}, 0)
	*/

	/*go func() {
		time.Sleep(time.Second)
		mt := zsync.RWMutex{}
		zsync.SetDebug(true)
		fmt.Println("R 111")
		mt.Lock()
		fmt.Println("R 222")
		mt.RLock()
		fmt.Println("R 333")
	}()*/
	a, _ := xxmap[3]
	fmt.Println("xxmap[3]:", a)
	mt := zsync.Mutex{}
	zsync.SetDebug(true)
	fmt.Println("111")
	mt.Lock()
	fmt.Println("222")
	mt.Unlock()
	fmt.Println("333")
	mt.Lock()
	fmt.Println("444")
	mt.Unlock()
	fmt.Println("555")
	mt.Lock()
	fmt.Println("666")
	mt.Lock()
	fmt.Println("777")
	time.Sleep(time.Hour)

}
