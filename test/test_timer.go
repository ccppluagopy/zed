package main

import (
	"fmt"
	"github.com/ccppluagopy/zed/timer"
	"time"
)

func main() {
	t0 := time.Now()

	tm := timer.NewTimer()
	var (
		item1  *timer.TimeItem
		item3  *timer.TimeItem
		item5  *timer.TimeItem
		item8  *timer.TimeItem
		item10 *timer.TimeItem
	)

	for i := 0; i < 5; i++ {
		n := i*2 + 1
		str := fmt.Sprintf("%02d - ", n)
		item := tm.NewItem(time.Second*time.Duration(n), func() {
			fmt.Println(str, time.Since(t0).Seconds())
		})
		if n == 1 {
			item1 = item
		}
		if n == 3 {
			item3 = item
		}
		if n == 5 {
			item5 = item
		}
		if n == 8 {
			item8 = item
		}
		if n == 10 {
			item10 = item
		}
	}

	for i := 0; i < 5; i++ {
		n := (i + 1) * 2
		str := fmt.Sprintf("%02d - ", n)
		item := tm.NewItem(time.Second*time.Duration(n), func() {
			fmt.Println(str, time.Since(t0).Seconds())
		})
		if n == 3 {
			item3 = item
		}
		if n == 5 {
			item5 = item
		}
		if n == 8 {
			item8 = item
		}
		if n == 10 {
			item10 = item
		}
	}

	fmt.Println("000 Size: ", tm.Size())

	tm.DeleteItem(item3)
	tm.DeleteItem(item5)
	tm.DeleteItem(item1)
	tm.DeleteItem(item1)
	tm.DeleteItem(item10)
	tm.DeleteItem(item8)
	fmt.Println("111 Size: ", tm.Size())

	time.Sleep(time.Second * 10)

	fmt.Println("222 Size: ", tm.Size())

	n := 0
	var scheduleItem *timer.TimeItem
	scheduleItem = tm.Schedule(time.Second*3, time.Second, func() {
		n++
		fmt.Println("Schedule: ", n, "pass:", time.Since(t0).Seconds())
		if n >= 5 {
			tm.DeleteItemInCall(scheduleItem)
		}

	})

	fmt.Println("333 Size: ", tm.Size())

	time.Sleep(time.Second * 10)

	fmt.Println("444 Size: ", tm.Size())

	fmt.Println("Over!")
}
