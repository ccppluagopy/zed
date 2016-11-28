package zed

import (
	//"github.com/ccppluagopy/zed"
	"sync"
	"time"
)

var (
	mtxdebug                          = false
	mtxlockTimeout                    = (time.Second * 5)
	mtxtimerWheel  *TimerWheel        = nil
	mtxmutexs      map[string]*WTimer = nil
	mtxlock                           = &sync.Mutex{}
)

func getLockTimer(key string) *WTimer {
	mtxlock.Lock()
	defer mtxlock.Unlock()
	if timer, ok := mtxmutexs[key]; ok {
		return timer
	}
	return nil
}

func saveLockTimer(key string, timer *WTimer) {
	mtxlock.Lock()
	defer mtxlock.Unlock()
	mtxmutexs[key] = timer
}

func unsaveLockTimer(key string) {
	mtxlock.Lock()
	defer mtxlock.Unlock()
	delete(mtxmutexs, key)
}

func SetMutexDebug(flag bool, args ...interface{}) {
	mtxdebug = flag
	if mtxdebug {
		if mtxtimerWheel == nil {
			mtxtimerWheel = NewTimerWheel(time.Second, 15)
		}
		if mtxmutexs == nil {
			mtxmutexs = make(map[string]*WTimer)
		}
	}
	if len(args) == 1 {
		t, ok := args[0].(time.Duration)
		if ok {
			mtxlockTimeout = t
			Printf("zsync.SetDebug set mtxlockTimeout: %v\n", t)
		}
	}
}

type Mutex struct {
	sync.Mutex
	unlockkey string
	lastCall  string
}

func (mt *Mutex) Lock() {
	if mtxdebug {
		key := Sprintf("%vl", &mt)
		if timer := getLockTimer(key); timer == nil {
			t1 := time.Now()
			stack := GetStackInfo()
			timer = mtxtimerWheel.NewTimer(key, mtxlockTimeout, func(t *WTimer) {
				Printf("zsync.Mutex Warn: Lock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				Println("now: ", t1.UnixNano())
				Println("this Call :", stack)
				Println("last Call :", mt.lastCall)
				unsaveLockTimer(key)
			}, 0)
			saveLockTimer(key, timer)
		}
	}

	mt.Mutex.Lock()

	if mtxdebug {
		{
			key := Sprintf("%vl", &mt)
			if timer := getLockTimer(key); timer != nil {
				mtxtimerWheel.DeleteTimer(timer)
				unsaveLockTimer(key)
				mt.lastCall = GetStackInfo()
			}
		}

		{
			if mt.unlockkey == "" {
				mt.unlockkey = Sprintf("%vul", &mt)
			}
			//Println("Lock Unlock key:", mt.unlockkey)
			if timer := getLockTimer(mt.unlockkey); timer == nil {
				t1 := time.Now()
				//stack := GetStackInfo()
				timer = mtxtimerWheel.NewTimer(mt.unlockkey, mtxlockTimeout, func(t *WTimer) {
					Printf("zsync.Mutex Warn: Wait Unlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
					Println("now: ", t1.UnixNano())
					//Println("this Call :", stack)
					Println("last Call :", mt.lastCall)
					unsaveLockTimer(mt.unlockkey)
				}, 0)
				saveLockTimer(mt.unlockkey, timer)
			}
		}
	}
}

func (mt *Mutex) Unlock() {
	mt.Mutex.Unlock()
	if mtxdebug {
		//key := Sprintf("%vul", &mt.Mutex)
		//Println("Unlock key:", mt.unlockkey)
		if timer := getLockTimer(mt.unlockkey); timer != nil {
			mtxtimerWheel.DeleteTimer(timer)
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
	if mtxdebug {
		key := Sprintf("%vl", &rwmt)
		if timer := getLockTimer(key); timer == nil {
			t1 := time.Now()
			stack := GetStackInfo()
			timer = mtxtimerWheel.NewTimer(key, mtxlockTimeout, func(t *WTimer) {
				Printf("zsync.RWMutex Warn: Lock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				Println("now: ", t1.UnixNano())
				Println("this Call :", stack)
				Println("last Call :", rwmt.lastCall)
				unsaveLockTimer(key)
			}, 0)
			saveLockTimer(key, timer)
		}

	}

	rwmt.RWMutex.Lock()

	if mtxdebug {
		key := Sprintf("%vl", &rwmt)
		if timer := getLockTimer(key); timer != nil {
			mtxtimerWheel.DeleteTimer(timer)
			unsaveLockTimer(key)
			rwmt.lastCall = GetStackInfo()
		}

		{
			if rwmt.unlockkey == "" {
				rwmt.unlockkey = Sprintf("%vul", &rwmt)
			}
			if timer := getLockTimer(rwmt.unlockkey); timer == nil {
				t1 := time.Now()
				//stack := GetStackInfo()
				timer = mtxtimerWheel.NewTimer(rwmt.unlockkey, mtxlockTimeout, func(t *WTimer) {
					Printf("zsync.RWMutex Warn: Wait Unlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
					Println("now: ", t1.UnixNano())
					//Println("this Call :", stack)
					Println("last Call :", rwmt.lastCall)
					unsaveLockTimer(rwmt.unlockkey)
				}, 0)
				saveLockTimer(rwmt.unlockkey, timer)
			}
		}
	}
}

func (rwmt *RWMutex) Unlock() {
	rwmt.RWMutex.Unlock()
	if mtxdebug {
		if timer := getLockTimer(rwmt.unlockkey); timer != nil {
			mtxtimerWheel.DeleteTimer(timer)
			unsaveLockTimer(rwmt.unlockkey)
		}
	}
}

func (rwmt *RWMutex) RLock() {
	if mtxdebug {
		key := Sprintf("%vrl", &rwmt)
		if timer := getLockTimer(key); timer == nil {
			t1 := time.Now()
			stack := GetStackInfo()
			timer := mtxtimerWheel.NewTimer(key, mtxlockTimeout, func(t *WTimer) {
				Printf("zsync.RWMutex Warn: RLock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
				Println("now: ", t1.UnixNano())
				Println("this Call :", stack)
				Println("last Call :", rwmt.lastCall)
				unsaveLockTimer(key)
			}, 0)
			saveLockTimer(key, timer)
		}
	}

	rwmt.RWMutex.RLock()

	if mtxdebug {
		key := Sprintf("%vrl", &rwmt)
		if timer := getLockTimer(key); timer != nil {
			mtxtimerWheel.DeleteTimer(timer)
			unsaveLockTimer(key)
			rwmt.lastCall = GetStackInfo()
		}

		/*{
			key2 := Sprintf("%vrul", &rwmt)
			if timer := getLockTimer(key2); timer == nil {
				t1 := time.Now()
				stack := GetStackInfo()
				timer := mtxtimerWheel.NewTimer(key2, mtxlockTimeout, func(t *WTimer) {
					Printf("zsync.RWMutex Warn: Wait RUnlock Timeout(%v seconds), May Be DeadLock!\n", time.Since(t1).Seconds())
					Println("now: ", time.Now().UnixNano())
					Println(stack)
					unsaveLockTimer(key2)
				}, 0)
				saveLockTimer(key2, timer)
			}
		}*/
	}
}

func (rwmt *RWMutex) RUnlock() {
	rwmt.RWMutex.RUnlock()
	/*if mtxdebug {
		key := Sprintf("%vrul", &rwmt)
		if timer := getLockTimer(key); timer != nil {
			mtxtimerWheel.DeleteTimer(timer)
			unsaveLockTimer(key)
		}
	}*/
}
