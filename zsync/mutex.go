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
	mutexs      map[string]*zed.WTimer = nil
	lock                               = &sync.Mutex{}
)

func getLockTimer(key string) *zed.WTimer {
	lock.Lock()
	defer lock.Unlock()
	if timer, ok := mutexs[key]; ok {
		return timer
	}
	return nil
}

func saveLockTimer(key string, timer *zed.WTimer) {
	lock.Lock()
	defer lock.Unlock()
	mutexs[key] = timer
}

func unsaveLockTimer(key string) {
	lock.Lock()
	defer lock.Unlock()
	delete(mutexs, key)
}

func SetDebug(flag bool, args ...interface{}) {
	debug = flag
	if debug {
		if timerWheel == nil {
			timerWheel = zed.NewTimerWheel(time.Second, 15)
		}
		if mutexs == nil {
			mutexs = make(map[string]*zed.WTimer)
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
		key := zed.Sprintf("%vl", &mt)
		if timer := getLockTimer(key); timer == nil {
			t1 := time.Now()
			stack := zed.GetStackInfo()
			timer = timerWheel.NewTimer(key, lockTimeout, func(t *zed.WTimer) {
				zed.Printf("zsync.Mutex Warn: Lock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				zed.Println(stack)
				delete(mutexs, key)
			}, 0)
			saveLockTimer(key, timer)
		}
	}

	mt.Mutex.Lock()

	if debug {
		{
			key := zed.Sprintf("%vl", &mt)
			if timer := getLockTimer(key); timer != nil {
				timerWheel.DeleteTimer(timer)
				unsaveLockTimer(key)
			}
		}

		{
			key2 := zed.Sprintf("%vul", &mt)
			if timer := getLockTimer(key2); timer == nil {
				t1 := time.Now()
				stack := zed.GetStackInfo()
				timer = timerWheel.NewTimer(key2, lockTimeout, func(t *zed.WTimer) {
					zed.Printf("zsync.Mutex Warn: Wait Unlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
					zed.Println(stack)
					delete(mutexs, key2)
				}, 0)
				saveLockTimer(key2, timer)
			}
		}
	}
}

func (mt *Mutex) Unlock() {
	mt.Mutex.Unlock()
	if debug {
		key := zed.Sprintf("%vul", &mt)
		if timer := getLockTimer(key); timer != nil {
			timerWheel.DeleteTimer(timer)
			unsaveLockTimer(key)
		}
	}
}

type RWMutex struct {
	sync.RWMutex
}

func (rwmt *RWMutex) Lock() {
	if debug {

		key := zed.Sprintf("%vl", &rwmt)
		if timer := getLockTimer(key); timer == nil {
			t1 := time.Now()
			stack := zed.GetStackInfo()
			timer = timerWheel.NewTimer(key, lockTimeout, func(t *zed.WTimer) {
				zed.Printf("zsync.RWMutex Warn: Lock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				zed.Println(stack)
				delete(mutexs, key)
			}, 0)
			saveLockTimer(key, timer)
		}

	}

	rwmt.RWMutex.Lock()

	if debug {
		key := zed.Sprintf("%vl", &rwmt)
		if timer := getLockTimer(key); timer != nil {
			timerWheel.DeleteTimer(timer)
			unsaveLockTimer(key)
		}

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
		key := zed.Sprintf("%vrl", &rwmt)
		if timer := getLockTimer(key); timer == nil {
			t1 := time.Now()
			stack := zed.GetStackInfo()
			timer := timerWheel.NewTimer(key, lockTimeout, func(t *zed.WTimer) {
				zed.Printf("zsync.RWMutex Warn: RLock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				zed.Println(stack)
				delete(mutexs, key)
			}, 0)
			saveLockTimer(key, timer)
		}
	}

	rwmt.RWMutex.RLock()

	if debug {
		key := zed.Sprintf("%vrl", &rwmt)
		if timer := getLockTimer(key); timer != nil {
			timerWheel.DeleteTimer(timer)
			unsaveLockTimer(key)
		}

		key2 := zed.Sprintf("%vrul", &rwmt)
		if timer := getLockTimer(key2); timer == nil {
			t1 := time.Now()
			stack := zed.GetStackInfo()
			timer := timerWheel.NewTimer(key2, lockTimeout, func(t *zed.WTimer) {
				zed.Printf("zsync.RWMutex Warn: Wait RUnlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				zed.Println(stack)
				delete(mutexs, key2)
			}, 0)
			saveLockTimer(key2, timer)
		}
	}
}

func (rwmt *RWMutex) RUnlock() {
	rwmt.RWMutex.RUnlock()
	if debug {
		key := zed.Sprintf("%vrul", &rwmt)
		if timer := getLockTimer(key); timer != nil {
			timerWheel.DeleteTimer(timer)
			unsaveLockTimer(key)
		}
	}
}
