package zsync

import (
	"github.com/ccppluagopy/zed"
	"sync"
	"time"
)

var (
	debug                              = false
	lockTimeout                        = (time.Second * 3)
	timerWheel  *zed.TimerWheel        = nil
	mutexs      map[string]interface{} = nil
)

func SetDebug(flag bool, args ...interface{}) {
	debug = flag
	if debug {
		if timerWheel == nil {
			timerWheel = zed.NewTimerWheel(time.Second, 15)
		}
		if mutexs == nil {
			mutexs = make(map[string]interface{})
		}
	}
	if len(args) == 1 {
		t, ok := args[0].(time.Duration)
		if ok {
			lockTimeout = t
		}
	}
}

type Mutex struct {
	sync.Mutex
}

func (mt *Mutex) Lock() {
	if debug {
		t1 := time.Now()
		key := zed.Sprintf("%vl", &mt)
		stack := zed.GetStackInfo()
		timer := timerWheel.NewTimer(key, lockTimeout, func(t *zed.WTimer) {
			zed.Printf("zsync.Mutex Warn: Lock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
			zed.Println(stack)
			delete(mutexs, key)
		}, 0)
		mutexs[key] = timer
	}

	mt.Mutex.Lock()

	if debug {
		key := zed.Sprintf("%vl", &mt)
		if t, ok := mutexs[key]; ok {
			if timer, ok2 := t.(*zed.WTimer); ok2 {
				timerWheel.DeleteTimer(timer)
			}
		}
		delete(mutexs, key)
	}
}

/*
func (mt *Mutex) Unlock() {
	if debug {
		t1 := time.Now()
		key := zed.Sprintf("%vu", &mt)
		stack := zed.GetStackInfo()
		timer := timerWheel.NewTimer(key, lockTimeout, func(t *zed.WTimer) {
			zed.Printf("zsync.Mutex Warn: Unlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
			zed.Println(stack)
			delete(mutexs, key)
		}, 0)

		mutexs[key] = timer
	}

	mt.Mutex.Unlock()

	if debug {
		key := zed.Sprintf("%vu", &mt)
		if t, ok := mutexs[key]; ok {
			if timer, ok2 := t.(*zed.WTimer); ok2 {
				timerWheel.DeleteTimer(timer)
			}
		}
		delete(mutexs, key)
	}

}
*/

type RWMutex struct {
	sync.RWMutex
}

func (rwmt *RWMutex) Lock() {
	if debug {
		t1 := time.Now()
		key := zed.Sprintf("%vl", &rwmt)
		stack := zed.GetStackInfo()
		timer := timerWheel.NewTimer(key, lockTimeout, func(t *zed.WTimer) {
			zed.Printf("zsync.RWMutex Warn: Lock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
			zed.Println(stack)
			delete(mutexs, key)
		}, 0)

		mutexs[key] = timer
	}

	rwmt.RWMutex.Lock()

	if debug {
		key := zed.Sprintf("%vl", &rwmt)
		if t, ok := mutexs[key]; ok {
			if timer, ok2 := t.(*zed.WTimer); ok2 {
				timerWheel.DeleteTimer(timer)
			}
		}
		delete(mutexs, key)
	}
}

/*
func (rwmt *RWMutex) Unlock() {
	if debug {
		t1 := time.Now()
		key := zed.Sprintf("%vu", &rwmt)
		stack := zed.GetStackInfo()
		timer := timerWheel.NewTimer(key, lockTimeout, func(t *zed.WTimer) {
			zed.Printf("zsync.RWMutex Warn: Unlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
			zed.Println(stack)
			delete(mutexs, key)
		}, 0)

		mutexs[key] = timer
	}

	rwmt.RWMutex.Unlock()

	if debug {
		key := zed.Sprintf("%vu", &rwmt)
		if t, ok := mutexs[key]; ok {
			if timer, ok2 := t.(*zed.WTimer); ok2 {
				timerWheel.DeleteTimer(timer)
			}
		}
		delete(mutexs, key)
	}
}
*/

func (rwmt *RWMutex) RLock() {
	if debug {
		t1 := time.Now()
		key := zed.Sprintf("%vrl", &rwmt)
		stack := zed.GetStackInfo()
		timer := timerWheel.NewTimer(key, lockTimeout, func(t *zed.WTimer) {
			zed.Printf("zsync.RWMutex Warn: RLock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
			zed.Println(stack)
			delete(mutexs, key)
		}, 0)

		mutexs[key] = timer
	}

	rwmt.RWMutex.RLock()

	if debug {
		key := zed.Sprintf("%vrl", &rwmt)
		if t, ok := mutexs[key]; ok {
			if timer, ok2 := t.(*zed.WTimer); ok2 {
				timerWheel.DeleteTimer(timer)
			}
		}
		delete(mutexs, key)
	}
}

/*
func (rwmt *RWMutex) RUnlock() {
	if debug {
		t1 := time.Now()
		key := zed.Sprintf("%vru", &rwmt)
		stack := zed.GetStackInfo()
		timer := timerWheel.NewTimer(key, lockTimeout, func(t *zed.WTimer) {
			zed.Printf("zsync.RWMutex Warn: RUnlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
			zed.Println(stack)
			delete(mutexs, key)
		}, 0)

		mutexs[key] = timer
	}

	rwmt.RWMutex.RUnlock()

	if debug {
		key := zed.Sprintf("%vru", &rwmt)
		if t, ok := mutexs[key]; ok {
			if timer, ok2 := t.(*zed.WTimer); ok2 {
				timerWheel.DeleteTimer(timer)
			}
		}
		delete(mutexs, key)
	}
}
*/
