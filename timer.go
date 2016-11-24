package zed

import (
	//"fmt"
	"time"
)

/*
timerWheel := Timer.NewTimerWheel(int64(tickTime), int64(time.Second), 2)
cb1 := func() {
	fmt.Println("cb1")
}
cb2 := func() {
	fmt.Println("cb2")
}
timerWheel.NewTimer("cb1", cb1, true)
time.Sleep(time.Second * 1)
timerWheel.NewTimer("cb2", cb2, true)
*/

/*
timerMgr := Timer.NewTimerMgr(int64(tickTime))

cb1 := func() {
	fmt.Println("cb1")
}
cb2 := func() {
	fmt.Println("cb2")
}
timerMgr.NewTimer("cb1", int64(time.Second), cb1, true)
timerMgr.NewTimer("cb2", int64(time.Second*2), cb2, true)
*/

func wtimerHandler(handler WTimerCallBack, timer *WTimer) {
	defer PanicHandle(true)
	handler(timer)
}

func timerHandler(handler TimerCallBack) {
	defer PanicHandle(true)
	handler()
}

type WTimer struct {
	key      interface{}
	active   bool
	delay    int64
	loop     int64
	loopCnt  int64
	wheelIdx int64
	born     int64
	start    int64
	callback WTimerCallBack
}

func (timer *WTimer) Stop() {
	timer.loop = 0
}

type wheel map[interface{}]*WTimer

func (timerWheel *TimerWheel) howmanyloops(delay int64) int64 {
	n := int64(delay) / (int64(timerWheel.internal) * int64(timerWheel.wheelNum))
	//Println("homanyloops 111: ", delay, n)
	if int64(delay)%(int64(timerWheel.internal)*int64(timerWheel.wheelNum)) > 0 {
		n++
		//Println("homanyloops 222: ", delay, n)
	}
	//Println("homanyloops 333: ", delay, n)
	return n
}

func (timerWheel *TimerWheel) NewTimer(key interface{}, delay time.Duration, callback WTimerCallBack, loopInternal time.Duration) *WTimer {
	timerWheel.Lock()
	defer timerWheel.Unlock()

	if _, ok := timerWheel.timers[key]; ok {
		ZLog("TimerWheel NewTimer Error: key(%v) has been exist.", key)
		return nil
	}

	now := time.Now().UnixNano()
	timer := &WTimer{
		key:      key,
		delay:    int64(delay),
		callback: callback,
		loop:     int64(loopInternal),
		born:     now,
		start:    now,
	}

	if timer.loop > 0 && timer.loop < timerWheel.internal {
		ZLog("TimerWheel NewTimer Error: loopInternal is not 0 and is less then TimerWheel's internal.")
		return nil
	}

	if timer.delay == 0 {
		wtimerHandler((timer).callback, timer)
		if timer.loop == 0 {
			return nil
		} else if timer.loop <= timerWheel.internal {
			timer.wheelIdx = (timerWheel.currWheel + 1) % timerWheel.wheelNum
			timer.loopCnt = timerWheel.howmanyloops(timer.loop)
		} else {
			timer.wheelIdx = (timerWheel.currWheel + (timer.loop)/timerWheel.internal) % timerWheel.wheelNum
			timer.loopCnt = timerWheel.howmanyloops(timer.loop)
		}
		//Println("NewTimer 111, currWheel, start, born, delay, wheelIdx", timerWheel.currWheel, timer.start, timerWheel.born, delay, timer.wheelIdx)
	} else if timer.delay <= timerWheel.internal {
		timer.wheelIdx = (timerWheel.currWheel + 1) % timerWheel.wheelNum
		timer.loopCnt = timerWheel.howmanyloops(timer.delay)
		//Println("NewTimer 222, currWheel, start, born, delay, wheelIdx", timerWheel.currWheel, timer.start, timerWheel.born, delay, timer.wheelIdx)
	} else {
		//sum := (timer.start-timerWheel.born)%timerWheel.internal + timerWheel.internal/2 + timer.delay - timerWheel.internal
		sum := (timerWheel.internal/2 + timer.delay)
		timer.wheelIdx = (timerWheel.currWheel + sum/timerWheel.internal) % timerWheel.wheelNum
		timer.loopCnt = timerWheel.howmanyloops(sum)
		//Println("NewTimer 333,", sum, timer.wheelIdx, timer.loopCnt)
	}

	timerWheel.wheels[timer.wheelIdx][timer.key] = timer

	timerWheel.timers[key] = timer

	return timer
}

func (timerWheel *TimerWheel) DeleteTimer(timer *WTimer) {
	if timer != nil {
		timerWheel.Lock()
		defer timerWheel.Unlock()

		delete(timerWheel.wheels[timer.wheelIdx], timer.key)
		delete(timerWheel.timers, timer.key)
	}
}

/*
func (timerWheel *TimerWheel) DeleteTimerByKey(key interface{}) {
	for _, wl := range timerWheel.wheels {
		for _, timer := range wl {
			if key == timer.key {
				timerWheel.DeleteTimer(timer)
				return
			}
		}
	}
}
*/
func (timerWheel *TimerWheel) Stop() {
	timerWheel.running = false
}

func (timerWheel *TimerWheel) IsRunning() bool {
	return timerWheel.running
}

func NewTimerWheel(wheelInternal time.Duration, wheelNum int64) *TimerWheel {
	var (
		now            = time.Now().UnixNano()
		sub      int64 = 0
		currTick int64 = 0
		internal       = int64(wheelInternal)
		//halfInternal       = internal / 2
	)

	timerWheel := TimerWheel{
		currWheel: 0,
		wheels:    make([]wheel, wheelNum),
		wheelNum:  wheelNum,
		born:      now,
		lastTick:  now,
		internal:  internal,
		timers:    make(map[interface{}]*WTimer),
		running:   true,
	}

	for i := 0; i < int(wheelNum); i++ {
		timerWheel.wheels[i] = make(map[interface{}]*WTimer)
	}

	timerfunc := func() {
		tickFunc := func() bool {
			timerWheel.Lock()
			defer timerWheel.Unlock()
			if !timerWheel.running {
				return false
			}

			for {
				currTick = time.Now().UnixNano()
				sub = currTick - timerWheel.lastTick
				if currTick-timerWheel.lastTick >= internal {
					timerWheel.currWheel = (timerWheel.currWheel + 1) % wheelNum
					wl := timerWheel.wheels[timerWheel.currWheel]
					for _, timer := range wl {
						timer.loopCnt--
						if timer.delay > 0 {
							timer.delay = 0
							timer.born = currTick
						}
						//Println("timer.loopCnt: ", timer.loopCnt)
						if timer.loopCnt <= 0 {
							wtimerHandler((timer).callback, timer)

							delete(wl, (timer).key)

							if timer.loop == 0 {
								delete(timerWheel.timers, timer.key)
							} else if timer.loop <= timerWheel.internal {
								timer.start = currTick
								timer.wheelIdx = (timerWheel.currWheel + 1) % timerWheel.wheelNum
								timer.loopCnt = timerWheel.howmanyloops(timer.loop)
								timerWheel.wheels[timer.wheelIdx][timer.key] = timer
							} else {
								timer.start = currTick
								//sum := ((timer.start-timerWheel.born)%timerWheel.internal - (timer.start-timer.born)%timer.loop + timerWheel.internal/2 + timer.loop - timerWheel.internal) / timerWheel.internal
								sum := (timerWheel.internal/2 + timer.loop) / timerWheel.internal
								timer.wheelIdx = (timer.wheelIdx + sum) % timerWheel.wheelNum
								timer.loopCnt = timerWheel.howmanyloops(sum)
								//Println("NewTimer 444,", sum, timer.wheelIdx, timer.loopCnt, ((timer.start-timer.born)%timer.loop)/timerWheel.internal)
								timerWheel.wheels[timer.wheelIdx][timer.key] = timer

							}
						}

					}
					timerWheel.lastTick = currTick
				} else {
					break
				}
			}
			return true
		}
		for {
			//n := int64(time.Duration(timerWheel.internal - (int64(time.Now().UnixNano())-timerWheel.born)%timerWheel.internal))
			//Println("Sleep n:", n, sub+internal)
			//time.Sleep(time.Duration(sub + internal))
			time.Sleep(time.Duration(timerWheel.internal - (int64(time.Now().UnixNano())-timerWheel.born)%timerWheel.internal))
			if !tickFunc() {
				break
			}
		}
	}
	NewCoroutine(timerfunc)

	return &timerWheel
}
