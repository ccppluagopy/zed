package timer
 


package main

import (
	"fmt"
	"net"
	"sync"
	"test"
	"time"
)

const (
	timerFormat      = "2006-01-02 15:04:05"
	infiniteDuration = time.Duration(1<<63 - 1)
)

var (
	Debug          = false
	timerLazyDelay = 300 * time.Millisecond
)

type TimerData struct {
	Key    string
	expire time.Time
	fn     func()
	index  int
	next   *TimerData
}

func (td *TimerData) Delay() time.Duration {
	return td.expire.Sub(time.Now())
}

func (td *TimerData) ExpireString() string {
	return td.expire.Format(timerFormat)
}

type Timer struct {
	lock   sync.Mutex
	free   *TimerData
	timers []*TimerData
	signal *time.Timer
}

// A heap must be initialized before any of the heap operations
// can be used. Init is idempotent with respect to the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// Its complexity is O(n) where n = h.Len().
//
func NewTimer(num int) (t *Timer) {
	t = new(Timer)
	t.init(num)
	return t
}

// Init init the timer.
func (t *Timer) Init(num int) {
	t.init(num)
}

func (t *Timer) init(num int) {
	t.signal = time.NewTimer(infiniteDuration)
	t.timers = make([]*TimerData, 0, num)
	tds := make([]TimerData, num)
	t.free = &(tds[0])
	td := t.free
	for i := 1; i < num; i++ {
		td.next = &(tds[i])
		td = td.next
	}
	go t.start()
}

// get get a free timer data.
func (t *Timer) get() (td *TimerData) {
	var (
		i   int
		num = len(t.timers)
		tds []TimerData
	)
	if td = t.free; td == nil {
		tds = make([]TimerData, num)
		t.free = &(tds[0])
		td = t.free
		for i = 1; i < num; i++ {
			td.next = &(tds[i])
			td = td.next
		}
		td = t.free
	}
	t.free = td.next
	return
}

// put put back a timer data.
func (t *Timer) put(td *TimerData) {
	td.next = t.free
	t.free = td
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len().
func (t *Timer) Add(expire time.Duration, fn func()) (td *TimerData) {
	t.lock.Lock()
	td = t.get()
	td.expire = time.Now().Add(expire)
	td.fn = fn
	t.add(td)
	t.lock.Unlock()
	return
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len().
func (t *Timer) add(td *TimerData) {
	var d time.Duration
	td.index = len(t.timers)
	// add to the minheap last node
	t.timers = append(t.timers, td)
	t.up(td.index)
	if td.index == 0 {
		// if first node, signal start goroutine
		d = td.Delay()
		t.signal.Reset(d)
		if Debug {
			fmt.Println("timer: add reset delay %d ms", int64(d)/int64(time.Millisecond))
		}
	}
	if Debug {
		fmt.Println("timer: push item key: %s, expire: %s, index: %d", td.Key, td.ExpireString(), td.index)
	}
	return
}

// Del removes the element at index i from the heap.
// The complexity is O(log(n)) where n = h.Len().
func (t *Timer) Del(td *TimerData) {
	t.lock.Lock()
	if t.del(td) {
		t.put(td)
	}
	t.lock.Unlock()
	return
}

func (t *Timer) del(td *TimerData) bool {
	var (
		i    = td.index
		last = len(t.timers) - 1
	)
	if i < 0 || i > last || t.timers[i] != td {
		// already remove, usually by expire
		if Debug {
			fmt.Printf("timer del i: %d, last: %d, %p,%p\n", i, last, t.timers[i], td)
		}
		return false
	}
	if i != last {
		t.swap(i, last)
		t.down(i, last)
		t.up(i)
	}
	// remove item is the last node
	t.timers[last].index = -1 // for safety
	t.timers = t.timers[:last]
	if Debug {
		fmt.Printf("timer: remove item key: %s, expire: %s, index: %d\n", td.Key, td.ExpireString(), td.index)
	}
	return true
}

// Set update timer data.
func (t *Timer) Set(td *TimerData, expire time.Duration) {
	t.lock.Lock()
	t.del(td)
	td.expire = time.Now().Add(expire)
	t.add(td)
	t.lock.Unlock()
	return
}

// start start the timer.
func (t *Timer) start() {
	for {
		t.expire()
		<-t.signal.C
	}
}

// expire removes the minimum element (according to Less) from the heap.
// The complexity is O(log(n)) where n = max.
// It is equivalent to Del(0).
func (t *Timer) expire() {
	var (
		td *TimerData
		d  time.Duration
	)
	t.lock.Lock()
	for {
		if len(t.timers) == 0 {
			d = infiniteDuration
			if Debug {
				fmt.Println("timer: no other instance")
			}
			break
		}
		td = t.timers[0]
		if d = td.Delay(); d > 0 {
			break
		}
		if td.fn == nil {
			fmt.Println("expire timer no fn")
		} else {
			if Debug {
				fmt.Printf("timer key: %s, expire: %s, index: %d expired, call fn\n", td.Key, td.ExpireString(), td.index)
			}
			td.fn()
		}
		// let caller put back
		t.del(td)
	}
	t.signal.Reset(d)
	if Debug {
		fmt.Printf("timer: expier reset delay %d ms\n", int64(d)/int64(time.Millisecond))
	}
	t.lock.Unlock()
	return
}

func (t *Timer) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !t.less(j, i) {
			break
		}
		t.swap(i, j)
		j = i
	}
}

func (t *Timer) down(i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !t.less(j1, j2) {
			j = j2 // = 2*i + 2  // right child
		}
		if !t.less(j, i) {
			break
		}
		t.swap(i, j)
		i = j
	}
}

func (t *Timer) less(i, j int) bool {
	return t.timers[i].expire.Before(t.timers[j].expire)
}

func (t *Timer) swap(i, j int) {
	//log.Debug("swap(%d, %d)", i, j)
	t.timers[i], t.timers[j] = t.timers[j], t.timers[i]
	t.timers[i].index = i
	t.timers[j].index = j
}

/*
func TestTimer() {
	timer := NewTimer(100)
	tds := make([]*TimerData, 20)
	for i := 0; i < 10; i++ {
		tmp := i
		tds[tmp*2] = timer.Add(time.Duration(i*2)*time.Second, func() {
			fmt.Println("idx: ", tmp, time.Duration(tmp*2)*time.Second)
			printTimer(timer)
		})
	}
	for i := 0; i < 10; i++ {
		tmp := i
		tds[tmp*2+1] = timer.Add(time.Duration(i*2+1)*time.Second, func() {
			fmt.Println("idx: ", tmp, time.Duration(tmp*2+1)*time.Second)
			printTimer(timer)
		})
	}
	printTimer(timer)
	for i := 0; i < 10; i++ {
		fmt.Println("td: %s, %s, %d", tds[i*2].Key, tds[i*2].ExpireString(), tds[i*2].index)
		timer.Del(tds[i*2])
	}

	test.Test2()
	time.Sleep(time.Second * 3600)

}

func printTimer(timer *Timer) {
	fmt.Printf("----------timers: %d ----------\n", len(timer.timers))
	for i := 0; i < len(timer.timers); i++ {
		fmt.Printf("timer: %s, %s, index: %d\n", timer.timers[i].Key, timer.timers[i].ExpireString(), timer.timers[i].index)
	}
	fmt.Println("--------------------")
}

func startListener(addr string) {
	defer fmt.Println("TcpServer Stopped.")

	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)

	go func() {
		time.Sleep(time.Second)
		listener.Close()
	}()

	for {
		_, err := listener.AcceptTCP()

		if err != nil {
			fmt.Println("exit--->")
			break
		} else {

		}
	}

}

func main() {
	startListener("127.0.0.1:8888")
	TestTimer()
}
*/