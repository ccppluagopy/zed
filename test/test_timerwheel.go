package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"sync"
	"time"
)

func test1() {
	var (
		internal   time.Duration = time.Second
		period     int64         = 5
		timerWheel               = zed.NewTimerWheel(internal, period)
		sum        int64         = 0
	)

	fmt.Println(".................... :::: ", time.Minute/time.Second)
	tb1 := time.Now()
	cb1 := func(t *zed.WTimer) {
		fmt.Println("cb1", time.Since(tb1).Seconds())
		tb1 = time.Now()
	}

	tb2 := time.Now()
	cb2 := func(t *zed.WTimer) {
		fmt.Println("cb2", time.Since(tb2).Seconds())
		tb2 = time.Now()
	}

	tb3 := time.Now()
	cb3 := func(t *zed.WTimer) {
		fmt.Println("cb3", time.Since(tb3).Seconds())
		tb3 = time.Now()
	}

	t1 := timerWheel.NewTimer("cb1", internal*3, cb1, internal*9)
	t2 := timerWheel.NewTimer("cb2", internal*2, cb2, 0)
	t3 := timerWheel.NewTimer("cb3", internal*3, cb3, 0)

	fmt.Println("test 111: ", time.Now().Unix())
	//fmt.Println("test aaa: ", t1.WheelIdx, t2.WheelIdx, t3.WheelIdx)

	time.Sleep(time.Second * 100000)
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
	fmt.Println("Test DeleteTimer", TEST_NUM, "times use seconds:", time.Since(tbegin).Seconds())
	time.Sleep(time.Hour)
}

func test2() {
	/*var (
		internal   time.Duration = time.Second
		period     int64         = 10
		timerWheel               = zed.NewTimerWheel(internal, period)
	)
	zed.Println("test2:")
	go func() {
		time.Sleep(internal * 1)
		var i time.Duration = 0
		for {
			time.Sleep(100000000)
			i++
			t1 := time.Now()
			delay := internal * time.Duration(i%15)
			currWheel := timerWheel.CurrWheel
			var delayIdx int64 = 0
			tt := timerWheel.NewTimer(100+i, delay, func(t *zed.WTimer) {
				sub := (time.Duration(time.Since(t1).Nanoseconds()) - delay)
				if sub < -time.Second || sub > time.Second {
					zed.Println("Error 111, delay, sub:", delay, sub, currWheel, delayIdx)
					t.Stop()
				}
				t1 = time.Now()
			}, delay)
			if tt != nil {
				delayIdx = tt.WheelIdx
			}
		}
	}()*/

	/*go func() {
		time.Sleep(internal * 3)
		var i time.Duration = 0
		for {
			time.Sleep(3)
			i++
			t1 := time.Now()
			delay := internal * time.Duration(i%20)
			currWheel := timerWheel.CurrWheel
			var delayIdx int64 = 0
			tt := timerWheel.NewTimer(-i, delay, func(t *zed.WTimer) {
				sub := (time.Duration(time.Since(t1).Nanoseconds()) - delay)
				if sub < -internal || sub > internal {
					zed.Println("Error 222, delay, sub:", delay, sub, currWheel, delayIdx)
					t.Stop()
				}
				t1 = time.Now()
				zed.Println("222, delay, sub:", delay, sub, currWheel, delayIdx)
			}, delay)
			delayIdx = tt.WheelIdx
		}
	}()*/
	/*go func() {
		time.Sleep(internal * 4)
		for i := 0; i < 10; i++ {
			time.Sleep(258)
			timerWheel.NewTimer(1000+i, internal*time.Duration(i), func(t *zed.WTimer) {

			}, internal)
		}
	}()*/
	time.Sleep(time.Hour)
}

func main() {
	test1()
}
