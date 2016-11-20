package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"sync"
	"time"
)

func main() {
	var (
		internal   time.Duration = time.Second
		period     int64         = 5
		timerWheel               = zed.NewTimerWheel(internal, period)
		sum        int64         = 0
	)

	//tt := time.Now()
	cb1 := func(t *zed.WTimer) {
		//fmt.Println("cb1", time.Since(tt).Seconds())
		fmt.Println("cb1", time.Now().Unix())
	}
	cb2 := func(t *zed.WTimer) {
		//fmt.Println("cb2", time.Since(tt).Seconds())
		fmt.Println("cb2", time.Now().Unix())
	}
	cb3 := func(t *zed.WTimer) {
		//fmt.Println("cb3", time.Since(tt).Seconds())
		fmt.Println("cb3", time.Now().Unix())
	}

	t1 := timerWheel.NewTimer("cb1", internal*1, cb1, internal*time.Duration(period+2))
	t2 := timerWheel.NewTimer("cb2", internal*2, cb2, internal)
	t3 := timerWheel.NewTimer("cb3", internal*3, cb3, internal)

	fmt.Println("test 111: ", time.Now().Unix())

	time.Sleep(time.Second * 10)
	timerWheel.DeleteTimer(t1)

	time.Sleep(time.Second * 3)
	timerWheel.DeleteTimer(t2)
	timerWheel.DeleteTimer(t3)

	fmt.Println("test end 222: ", time.Now().Unix())
	TEST_NUM := 100000
	wg := sync.WaitGroup{}
	wg.Add(TEST_NUM)
	tbegin := time.Now()

	timers := []*zed.WTimer{}
	for i := 0; i < TEST_NUM; i++ {
		flag := true
		var n int64 = int64(i)
		var timer *zed.WTimer
		tb := time.Now()
		//timer = timerWheel.NewTimer(n, time.Second*time.Duration(n%period), func() {

		timer = timerWheel.NewTimer(n, time.Second, func(t *zed.WTimer) {
			if flag {
				sum += n
				flag = !flag
				wg.Done()
				if timeout := time.Since(tb).Seconds(); timeout > float64(internal) {
					fmt.Println("timeout :", n, timeout)
				}
			} else {

				//fmt.Println("n: ", n, timer)
				/*timer.Stop()
				wg.Done()*/
			}

			//}, internal*time.Duration(n+1%period))
		}, time.Second)
		timers = append(timers, timer)
	}

	fmt.Println("Test NewTimer", TEST_NUM, "times use seconds:", time.Since(tbegin).Seconds())
	wg.Wait()
	fmt.Println("sum: ", sum, (TEST_NUM-1)*TEST_NUM/2)
	tbegin = time.Now()

	for i := 0; i < len(timers); i++ {
		timerWheel.DeleteTimer(timers[i])
	}
	//wg.Wait()
	fmt.Println("Test DeleteTimer", TEST_NUM, "times use seconds:", time.Since(tbegin).Seconds())
	time.Sleep(time.Hour)
}
