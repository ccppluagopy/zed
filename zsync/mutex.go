package zsync

import (
	"github.com/ccppluagopy/zed"
	"sync"
	"time"
)

var (
	debug                                = false
	lockTimeout                          = (time.Second * 5)
	locktimer   *zed.Timer               = nil
	mutexs      map[string]*zed.TimeItem = nil
	lock                                 = &sync.Mutex{}
)

func getLockTimer(key string) *zed.TimeItem {
	lock.Lock()
	defer lock.Unlock()
	if timer, ok := mutexs[key]; ok {
		return timer
	}
	return nil
}

func saveLockTimer(key string, timer *zed.TimeItem) {
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
		if locktimer == nil {
			locktimer = zed.NewTimer()
		}
		if mutexs == nil {
			mutexs = make(map[string]*zed.TimeItem)
		}
	}
	if len(args) == 1 {
		t, ok := args[0].(time.Duration)
		if ok {
			lockTimeout = t
			zed.Printf("zsync.SetDebug set lockTimeout: %v\n", t)
		}
	}
}

type Mutex struct {
	sync.Mutex
	unlockkey string
	lastCall  string
}

func (mt *Mutex) Lock() {
	if debug {
		key := zed.Sprintf("%vl", &mt)
		if timer := getLockTimer(key); timer == nil {
			t1 := time.Now()
			stack := zed.GetStackInfo()
			timer = locktimer.NewItem(lockTimeout, func() {
				zed.Printf("zsync.Mutex Warn: Lock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				zed.Println("now: ", t1.UnixNano())
				zed.Println("this Call :", stack)
				zed.Println("last Call :", mt.lastCall)
				unsaveLockTimer(key)
			})
			saveLockTimer(key, timer)
		}
	}

	mt.Mutex.Lock()

	if debug {
		{
			key := zed.Sprintf("%vl", &mt)
			if timer := getLockTimer(key); timer != nil {
				zed.Println("====", timer, locktimer)
				locktimer.DeleteItem(timer)
				unsaveLockTimer(key)
				mt.lastCall = zed.GetStackInfo()
			}
		}

		{
			if mt.unlockkey == "" {
				mt.unlockkey = zed.Sprintf("%vul", &mt)
			}
			//zed.Println("Lock Unlock key:", mt.unlockkey)
			if timer := getLockTimer(mt.unlockkey); timer == nil {
				t1 := time.Now()
				//stack := zed.GetStackInfo()
				timer = locktimer.NewItem(lockTimeout, func() {
					zed.Printf("zsync.Mutex Warn: Wait Unlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
					zed.Println("now: ", t1.UnixNano())
					//zed.Println("this Call :", stack)
					zed.Println("last Call :", mt.lastCall)
					unsaveLockTimer(mt.unlockkey)
				})
				saveLockTimer(mt.unlockkey, timer)
			}
		}
	}
}

func (mt *Mutex) Unlock() {
	mt.Mutex.Unlock()
	if debug {
		//key := zed.Sprintf("%vul", &mt.Mutex)
		//zed.Println("Unlock key:", mt.unlockkey)
		if timer := getLockTimer(mt.unlockkey); timer != nil {
			locktimer.DeleteItem(timer)
			unsaveLockTimer(mt.unlockkey)
		}
	}
}

type RWMutex struct {
	sync.RWMutex
	unlockkey string
	lastCall  string
}

func (rwmt *RWMutex) Lock() {
	if debug {
		key := zed.Sprintf("%vl", &rwmt)
		if timer := getLockTimer(key); timer == nil {
			t1 := time.Now()
			stack := zed.GetStackInfo()
			timer = locktimer.NewItem(lockTimeout, func() {
				zed.Printf("zsync.RWMutex Warn: Lock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				zed.Println("now: ", t1.UnixNano())
				zed.Println("this Call :", stack)
				zed.Println("last Call :", rwmt.lastCall)
				unsaveLockTimer(key)
			})
			saveLockTimer(key, timer)
		}
	}

	rwmt.RWMutex.Lock()

	if debug {
		key := zed.Sprintf("%vl", &rwmt)
		if timer := getLockTimer(key); timer != nil {
			locktimer.DeleteItem(timer)
			unsaveLockTimer(key)
			rwmt.lastCall = zed.GetStackInfo()
		}

		{
			if rwmt.unlockkey == "" {
				rwmt.unlockkey = zed.Sprintf("%vul", &rwmt)
			}
			if timer := getLockTimer(rwmt.unlockkey); timer == nil {
				t1 := time.Now()
				//stack := zed.GetStackInfo()
				timer = locktimer.NewItem(lockTimeout, func() {
					zed.Printf("zsync.RWMutex Warn: Wait Unlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
					zed.Println("now: ", t1.UnixNano())
					//zed.Println("this Call :", stack)
					zed.Println("last Call :", rwmt.lastCall)
					unsaveLockTimer(rwmt.unlockkey)
				})
				saveLockTimer(rwmt.unlockkey, timer)
			}
		}
	}
}

func (rwmt *RWMutex) Unlock() {
	rwmt.RWMutex.Unlock()
	if debug {
		if timer := getLockTimer(rwmt.unlockkey); timer != nil {
			locktimer.DeleteItem(timer)
			unsaveLockTimer(rwmt.unlockkey)
		}
	}
}

func (rwmt *RWMutex) RLock() {
	if debug {
		key := zed.Sprintf("%vrl", &rwmt)
		if timer := getLockTimer(key); timer == nil {
			t1 := time.Now()
			stack := zed.GetStackInfo()
			timer := locktimer.NewItem(lockTimeout, func() {
				zed.Printf("zsync.RWMutex Warn: RLock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				zed.Println("now: ", t1.UnixNano())
				zed.Println("this Call :", stack)
				zed.Println("last Call :", rwmt.lastCall)
				unsaveLockTimer(key)
			})
			saveLockTimer(key, timer)
		}
	}

	rwmt.RWMutex.RLock()

	if debug {
		key := zed.Sprintf("%vrl", &rwmt)
		if timer := getLockTimer(key); timer != nil {
			locktimer.DeleteItem(timer)
			unsaveLockTimer(key)
			rwmt.lastCall = zed.GetStackInfo()
		}

		/*{
			key2 := zed.Sprintf("%vrul", &rwmt)
			if timer := getLockTimer(key2); timer == nil {
				t1 := time.Now()
				stack := zed.GetStackInfo()
				timer := locktimer.NewTimer(key2, lockTimeout, func(t *zed.WTimer) {
					zed.Printf("zsync.RWMutex Warn: Wait RUnlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
					zed.Println("now: ", time.Now().UnixNano())
					zed.Println(stack)
					delete(mutexs, key2)
				}, 0)
				saveLockTimer(key2, timer)
			}
		}*/
	}
}

func (rwmt *RWMutex) RUnlock() {
	rwmt.RWMutex.RUnlock()
	/*if debug {
		key := zed.Sprintf("%vrul", &rwmt)
		if timer := getLockTimer(key); timer != nil {
			timerWheel.DeleteTimer(timer)
			unsaveLockTimer(key)
		}
	}*/
}
