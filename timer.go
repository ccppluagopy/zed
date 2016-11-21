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
	wheelIdx int64
	start    int64
	callback WTimerCallBack
}

func (timer *WTimer) Stop() {
	timer.loop = 0
}

type wheel map[interface{}]*WTimer

func (timerWheel *TimerWheel) NewTimer(key interface{}, delay time.Duration, callback WTimerCallBack, loopInternal time.Duration) *WTimer {
	timerWheel.Lock()
	defer timerWheel.Unlock()

	if _, ok := timerWheel.timers[key]; ok {
		ZLog("TimerWheel NewTimer Error: key(%v) has been exist.", key)
		return nil
	}

	timer := &WTimer{}
	timer.key = key
	timer.delay = int64(delay)
	timer.callback = callback
	timer.loop = int64(loopInternal)
	//timer.active = true
	timer.start = time.Now().UnixNano()
	//timerWheel.chTimer <- timer

	if timer.loop > 0 && timer.loop < timerWheel.internal {
		ZLog("TimerWheel NewTimer Error: loopInternal is not 0 and is less then TimerWheel's internal.")
		return nil
	}

	if timer.delay == 0 {
		wtimerHandler((timer).callback, timer)
		timer.wheelIdx = (timerWheel.currWheel + (timer.start - timerWheel.born + timerWheel.internal/2 + timer.loop)) / timerWheel.internal % timerWheel.wheelNum
	} else if timer.delay < timerWheel.internal {
		timer.wheelIdx = timerWheel.currWheel + 1
	} else {
		timer.wheelIdx = (timerWheel.currWheel + (timer.start - timerWheel.born + timerWheel.internal/2 + timer.delay)) / timerWheel.internal % timerWheel.wheelNum
	}
	//Println("NewTimer currWheel, wheelIdx:", timerWheel.currWheel, timer.wheelIdx, timerWheel.wheelNum, timer.start, timer.delay, timerWheel.internal/2)
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
	//close(timerWheel.chTimer)
	//timerWheel.ticker.Stop()
}

func (timerWheel *TimerWheel) IsRunning() bool {
	return timerWheel.running
}

func NewTimerWheel(wheelInternal time.Duration, wheelNum int64) *TimerWheel {
	var timerWheel TimerWheel

	//timerWheel.chTimer = make(chan *WTimer, 10)
	timerWheel.currWheel = 0
	timerWheel.wheels = make([]wheel, wheelNum)
	timerWheel.wheelNum = wheelNum

	timerWheel.born = time.Now().UnixNano()
	timerWheel.internal = int64(wheelInternal)
	timerWheel.timers = make(map[interface{}]*WTimer)

	internal := timerWheel.internal
	var i int64
	for i = 0; i < wheelNum; i++ {
		timerWheel.wheels[i] = make(map[interface{}]*WTimer)
	}
	timerWheel.running = true

	var tickSum int64 = 0
	var currTick int64 = 0

	var halfInternal = timerWheel.internal / 2

	timerWheel.lastTick = timerWheel.born

	timerfunc := func() {
		tickFunc := func() bool {
			timerWheel.Lock()
			defer timerWheel.Unlock()
			if !timerWheel.running {
				return false
			}

			for {
				currTick = time.Now().UnixNano()
				tickSum += (currTick - timerWheel.lastTick)
				timerWheel.lastTick = currTick
				if tickSum >= internal {
					tickSum -= internal
					timerWheel.currWheel = (timerWheel.currWheel + 1) % wheelNum

					wl := timerWheel.wheels[timerWheel.currWheel]

					for _, timer := range wl {
						if timer.delay > 0 {
							if currTick-timer.start+halfInternal >= timer.delay {
								wtimerHandler((timer).callback, timer)
								timer.delay = 0

								delete(wl, (timer).key)
								if timer.loop > 0 {
									timer.start = currTick
									timer.wheelIdx = (timerWheel.currWheel + (tickSum+timerWheel.internal/2+timer.loop)/timerWheel.internal) % timerWheel.wheelNum
									timerWheel.wheels[timer.wheelIdx][timer.key] = timer
								}
							}
						} else {
							if currTick-timer.start+halfInternal >= timer.loop {
								wtimerHandler((timer).callback, timer)

								delete(wl, (timer).key)
								if timer.loop > 0 {
									timer.start = currTick
									timer.wheelIdx = (timerWheel.currWheel + (tickSum+timerWheel.internal/2+timer.loop)/timerWheel.internal) % timerWheel.wheelNum
									timerWheel.wheels[timer.wheelIdx][timer.key] = timer
								}
							}
						}

					}
				} else {
					break
				}
			}
			return true
		}
		for {
			time.Sleep(time.Duration(timerWheel.internal - (int64(time.Now().UnixNano())-timerWheel.born)%timerWheel.internal))
			if !tickFunc() {
				break
			}
		}
	}
	NewCoroutine(timerfunc)

	return &timerWheel
}

type mtimer struct {
	key      interface{}
	internal int64
	born     int64
	active   bool
	delay    int64
	loop     bool
	callback TimerCallBack
}

type TimerMgr struct {
	running bool
	chTimer chan *mtimer
	timers  map[interface{}]*mtimer
	ticker  *time.Ticker
}

func (timerMgr *TimerMgr) NewTimer(key interface{}, delay int64, internal int64, callback TimerCallBack, loop bool) *mtimer {
	var timer mtimer
	timer.key = key
	timer.internal = internal
	timer.callback = callback
	timer.loop = loop
	timer.born = time.Now().UnixNano() + delay
	timer.active = true
	timerMgr.chTimer <- &timer

	//LogInfo(LOG_IDX, LOG_IDX, "new %s %d %d\n", key, internal, timer.born)
	return &timer
}

func (timerMgr *TimerMgr) DeleteTimer(timer *mtimer) {
	timer.active = false
	timerMgr.chTimer <- timer
}

func (timerMgr *TimerMgr) Stop() {
	timerMgr.running = false
	close(timerMgr.chTimer)
	timerMgr.ticker.Stop()
}

func (timerMgr *TimerMgr) IsRunning() bool {
	return timerMgr.running
}

func NewTimerMgr(internal int64) *TimerMgr {
	var timerMgr TimerMgr

	timerMgr.chTimer = make(chan *mtimer)
	timerMgr.timers = make(map[interface{}]*mtimer)
	timerMgr.ticker = time.NewTicker(time.Duration(internal))
	timerMgr.running = true

	n := 0
	timerfunc := func() {
		for {
			if !timerMgr.running {
				break
			}

			select {
			case timer, ok := <-timerMgr.chTimer:
				if !ok {
					return
				}
				if (*timer).active {
					timerMgr.timers[timer.key] = timer
				} else {
					delete(timerMgr.timers, (timer).key)
				}

			case <-timerMgr.ticker.C:
				currTime := time.Now().UnixNano()
				for key, timer := range timerMgr.timers {
					if (currTime - timer.born) >= 0 { //timer.internal {
						timerHandler((timer).callback)
						(*timer).born = currTime + timer.internal

						if !(*timer).loop {
							delete(timerMgr.timers, key)
						}
					}
				}
				n = n + 1
			}
		}
	}
	NewCoroutine(timerfunc)

	return &timerMgr
}
