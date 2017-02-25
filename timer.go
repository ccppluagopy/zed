package zed

import (
	"container/heap"
	//"fmt"
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
	//zed.Println("Less i j:", i, j)
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
		*tm = (*tm)[:n]
		heap.Fix(tm, i)
	} else {
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
	//zed.Println("=== 111 Index:", item.Index)
	heap.Fix(&(tm.timers), item.Index)
	//zed.Println("=== 222 Index:", item.Index)
	if head := tm.timers.Head(); head == item {
		//zed.Println("=== 333 Index:", head.Index, head.Expire.Sub(time.Now()))
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
	//zed.Println("=== 111 Index:", item.Index)
	heap.Fix(&(tm.timers), item.Index)
	//zed.Println("=== 222 Index:", item.Index)
	if head := tm.timers.Head(); head == item {
		//zed.Println("=== 333 Index:", head.Index, head.Expire.Sub(time.Now()))
		tm.signal.Reset(head.Expire.Sub(now))
	}

	return item
}

func (tm *Timer) DeleteItem(item *TimeItem) {
	tm.Lock()
	defer tm.Unlock()
	//zed.Println("DeleteItem: ", item.Index, item.Expire.Sub(t0))
	n := tm.timers.Len()
	if n == 0 {
		ZLog("Timer DeleteItem Error: Timer Size Is 0!")
	}
	if item.Index > 0 && item.Index < n {
		if item != tm.timers[item.Index] {
			zed.ZLog("Timer DeleteItem Error: Invalid Item!")
		}
		tm.timers.Remove(item.Index)
	} else if item.Index == 0 {
		if item != tm.timers[item.Index] {
			zed.ZLog("Timer DeleteItem Error: Invalid Item!")
		}
		tm.timers.Remove(item.Index)
		if head := tm.timers.Head(); head != nil && head != item {
			//zed.Println("=== 333 Index:", head.Index, head.Expire.Sub(time.Now()))
			tm.signal.Reset(head.Expire.Sub(time.Now()))
		}
	} else {
		ZLog("Timer DeleteItem Error: Invalid Index: %d", item.Index)
	}
}

func (tm *Timer) DeleteItemInCall(item *TimeItem) {
	//zed.Println("DeleteItem: ", item.Index, item.Expire.Sub(t0))
	n := tm.timers.Len()
	if n == 0 {
		ZLog("Timer DeleteItem Error: Timer Size Is 0!")
	}
	if item.Index > 0 && item.Index < n {
		if item != tm.timers[item.Index] {
			zed.ZLog("Timer DeleteItem Error: Invalid Item!")
		}
		tm.timers.Remove(item.Index)
	} else if item.Index == 0 {
		if item != tm.timers[item.Index] {
			zed.ZLog("Timer DeleteItem Error: Invalid Item!")
		}
		tm.timers.Remove(item.Index)
		if head := tm.timers.Head(); head != nil && head != item {
			//zed.Println("=== 333 Index:", head.Index, head.Expire.Sub(time.Now()))
			tm.signal.Reset(head.Expire.Sub(time.Now()))
		}
	} else {
		ZLog("Timer DeleteItem Error: Invalid Index: %d", item.Index)
	}
}

/*
func (tm *Timer) NewGroupItem(timeout time.Duration, cb func(), gidx int) *TimeItem {
	tm.Lock()
	defer tm.Unlock()

	item := &TimeItem{
		Index:    tm.timers.Len(),
		GroupIdx: gidx,
		Expire:   time.Now().Add(timeout),
		Callback: cb,
	}
	tm.timers = append(tm.timers, item)
	heap.Fix(&(tm.timers), item.Index)

	return item
}
*/

func (tm *Timer) Size() int {
	tm.Lock()
	defer tm.Unlock()
	return len(tm.timers)
}

func NewTimer() *Timer {
	tm := &Timer{
		signal: time.NewTimer(TIME_FOREVER),
		timers: []*TimeItem{},
	}

	once := func() {
		tm.Lock()
		defer func() {
			tm.Unlock()
			PanicHandle(true)
		}()

		if item := tm.timers.Pop(); item != nil {
			item.(*TimeItem).Callback()

			if head := tm.timers.Head(); head != nil {
				tm.signal.Reset(head.Expire.Sub(time.Now()))
			}
		} else {
			tm.signal.Reset(TIME_FOREVER)
		}
	}

	go func() {
		for {
			<-tm.signal.C
			once()
		}
	}()

	return tm
}
