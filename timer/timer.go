package timer

import (
	"container/heap"
	"fmt"
	//"github.com/ccppluagopy/zed"
	"sync"
	"time"
)

const (
	TIME_FOREVER = time.Duration(1<<63 - 1)
)

var (
	t0 = time.Now()
)

type TimeItem struct {
	Index int
	//GroupIdx int
	Expire   time.Time
	Callback func()
}

type Timers []*TimeItem

func (tm Timers) Len() int { return len(tm) }

func (tm Timers) Less(i, j int) bool {
	return tm[i].Expire.Before(tm[j].Expire)
}

func (tm *Timers) Swap(i, j int) {
	t := *tm
	t[i], t[j] = t[j], t[i]
	t[i].Index, t[j].Index = i, j
}

func (tm *Timers) Push(x interface{}) {
	n := len(*tm)
	item := x.(*TimeItem)
	item.Index = n
	*tm = append(*tm, item)
}

func (tm *Timers) Remove(i int) {
	n := tm.Len() - 1
	if n != i {
		(*tm).Swap(i, n)
		(*tm)[n] = nil
		*tm = (*tm)[:n]
		heap.Fix(tm, i)
	} else {
		(*tm)[n] = nil
		*tm = (*tm)[:n]
	}

}

func (tm *Timers) Pop() interface{} {
	old := *tm
	n := len(old)
	if n > 0 {
		tm.Swap(0, n-1)
		item := old[n-1]
		item.Index = -1 // for safety
		*tm = old[:n-1]
		heap.Fix(tm, 0)
		return item
	} else {
		return nil
	}
}

func (tm *Timers) Head() *TimeItem {
	t := *tm
	n := t.Len()
	if n > 0 {
		return t[0]
	}
	return nil
}

type Timer struct {
	sync.Mutex
	timers Timers
	signal *time.Timer
}

func (tm *Timer) NewItem(timeout time.Duration, cb func()) *TimeItem {
	tm.Lock()
	defer tm.Unlock()

	item := &TimeItem{
		Index:    len(tm.timers),
		Expire:   time.Now().Add(timeout),
		Callback: cb,
	}
	tm.timers = append(tm.timers, item)
	heap.Fix(&(tm.timers), item.Index)
	if head := tm.timers.Head(); head == item {
		tm.signal.Reset(head.Expire.Sub(time.Now()))
	}

	return item
}

func (tm *Timer) Schedule(delay time.Duration, internal time.Duration, cb func()) *TimeItem {
	tm.Lock()
	defer tm.Unlock()

	var (
		item *TimeItem
		now  = time.Now()
	)

	item = &TimeItem{
		Index:  len(tm.timers),
		Expire: now.Add(delay),
		Callback: func() {
			now = time.Now()
			item.Index = len(tm.timers)
			item.Expire = now.Add(internal)
			tm.timers = append(tm.timers, item)
			heap.Fix(&(tm.timers), item.Index)

			cb()

			if head := tm.timers.Head(); head == item {
				tm.signal.Reset(head.Expire.Sub(now))
			}
		},
	}

	tm.timers = append(tm.timers, item)
	heap.Fix(&(tm.timers), item.Index)
	if head := tm.timers.Head(); head == item {
		tm.signal.Reset(head.Expire.Sub(now))
	}

	return item
}

func (tm *Timer) DeleteItem(item *TimeItem) {
	tm.Lock()
	defer tm.Unlock()
	n := tm.timers.Len()
	if n == 0 {
		fmt.Println("Timer DeleteItem Error: Timer Size Is 0!")
		return
	}
	if item.Index > 0 && item.Index < n {
		if item != tm.timers[item.Index] {
			fmt.Println("Timer DeleteItem Error: Invalid Item!")
			return
		}
		tm.timers.Remove(item.Index)
	} else if item.Index == 0 {
		if item != tm.timers[item.Index] {
			fmt.Println("Timer DeleteItem Error: Invalid Item!")
			return
		}
		tm.timers.Remove(item.Index)
		if head := tm.timers.Head(); head != nil && head != item {
			tm.signal.Reset(head.Expire.Sub(time.Now()))
		}
	} else {
		fmt.Println("Timer DeleteItem Error: Invalid Index: %d", item.Index)
	}
}

func (tm *Timer) DeleteItemInCall(item *TimeItem) {
	n := tm.timers.Len()
	if n == 0 {
		fmt.Println("Timer DeleteItem Error: Timer Size Is 0!")
		return
	}
	if item.Index > 0 && item.Index < n {
		if item != tm.timers[item.Index] {
			fmt.Println("Timer DeleteItem Error: Invalid Item!")
			return
		}
		tm.timers.Remove(item.Index)
	} else if item.Index == 0 {
		if item != tm.timers[item.Index] {
			fmt.Println("Timer DeleteItem Error: Invalid Item!")
			return
		}
		tm.timers.Remove(item.Index)
		if head := tm.timers.Head(); head != nil && head != item {
			tm.signal.Reset(head.Expire.Sub(time.Now()))
		}
	} else {
		fmt.Println("Timer DeleteItem Error: Invalid Index: %d", item.Index)
	}
}

func (tm *Timer) ResetItem(item *TimeItem, delay time.Duration) {
	tm.Lock()
	defer tm.Unlock()

	n := tm.timers.Len()
	if n == 0 {
		fmt.Println("Timer ResetItem Error: Timer Size Is 0!")
		return
	}
	if item.Index < n {
		if item != tm.timers[item.Index] {
			fmt.Println("Timer ResetItem Error: Invalid Item!")
			return
		}
		item.Expire = time.Now().Add(delay)
		heap.Fix(&(tm.timers), item.Index)
	} else {
		fmt.Println("Timer ResetItem Error: Invalid Item!")
	}
}

func (tm *Timer) Size() int {
	tm.Lock()
	defer tm.Unlock()
	return len(tm.timers)
}

func (tm *Timer) Stop() {
	tm.Lock()
	defer tm.Unlock()
	tm.signal.Stop()
}

func NewTimer() *Timer {
	tm := &Timer{
		signal: time.NewTimer(TIME_FOREVER),
		timers: []*TimeItem{},
	}

	once := func() {
		defer func() {
			recover()
		}()
		tm.Lock()
		if item := tm.timers.Pop(); item != nil {
			if head := tm.timers.Head(); head != nil {
				tm.signal.Reset(head.Expire.Sub(time.Now()))
			}
		} else {
			tm.signal.Reset(TIME_FOREVER)
		}
		tm.Unlock()

		item.(*TimeItem).Callback()
	}

	go func() {
		ok := false
		for {
			if _, ok = <-tm.signal.C; !ok {
				break
			}
			once()
		}
	}()

	return tm
}
