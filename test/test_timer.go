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

	fmt.Println("******************")
	fmt.Println("Delete 333333: ", item3.Index)
	tm.DeleteItem(item3)
	fmt.Println("Size: ", tm.Size())

	fmt.Println("******************")
	fmt.Println("Delete 555555: ", item5.Index)
	tm.DeleteItem(item5)
	fmt.Println("Size: ", tm.Size())

	fmt.Println("******************")
	fmt.Println("Delete 111111: ", item1.Index)
	tm.DeleteItem(item1)
	fmt.Println("Size: ", tm.Size())

	fmt.Println("******************")
	fmt.Println("Delete 101010: ", item10.Index)
	tm.DeleteItem(item10)
	fmt.Println("Size: ", tm.Size())

	fmt.Println("******************")
	fmt.Println("Delete 888888: ", item8.Index)
	tm.DeleteItem(item8)
	fmt.Println("Size: ", tm.Size())
	fmt.Println("******************")
	/*fmt.Println(item3)
	fmt.Println(item5)
	fmt.Println(item8)*/
	time.Sleep(time.Second * 12)

	fmt.Println("Size: ", tm.Size())
	time.Sleep(time.Hour)
}
