package monitor

import (
	"Logger"
	"NetCore"
	"net/http"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	running       = false
	chAppStop     chan int
	ticker        *time.Ticker
	chMonitorStop = make(chan byte, 1)
)

func checkState() {
	zed.ZLog("Now: %v ClientsNum (%d) GoroutineNum: (%d).", time.Now(), NetCore.GetOnlineClientsNum(), runtime.NumGoroutine())
	//runtime.GC()
}

func heapHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("heap")
	p.WriteTo(w, 1)
}

func goroutineHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("goroutine")
	p.WriteTo(w, 1)
}

func Start(addr string, internal time.Duration, chStop chan int) {
	chAppStop = chStop

	ticker = time.NewTicker(internal)

	running = true

	go func() {
		for {
			if !running {
				break
			}

			select {
			case <-ticker.C:
				checkState()
			case <-chMonitorStop:
				return
			}
			//time.Sleep(time.Second * internal)

		}
	}()

	/*
		http: //localhost:10086/
	*/
	go func() {
		http.HandleFunc("/heap", heapHandler)
		http.HandleFunc("/goroutine", goroutineHandler)
		http.ListenAndServe(addr, nil)
	}()
}

func Stop() {
	running = false
	ticker.Stop()
	close(chMonitorStop)
	checkState()

	Logger.Println("[ShutDown] Monitor Stop!")
}
