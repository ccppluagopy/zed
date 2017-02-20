package monitor

import (
	"github.com/ccppluagopy/zed"
	"net/http"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	running       = false
	ticker        *time.Ticker
	chMonitorStop = make(chan byte, 1)
)

func checkState() {
	zed.ZLog("Now: %v GoroutineNum: (%d).", time.Now(), runtime.NumGoroutine())
	//runtime.GC()
}

func goroutineHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("goroutine")
	p.WriteTo(w, 1)
}

func threadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("threadcreate")
	p.WriteTo(w, 1)
}

func heapHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("heap")
	p.WriteTo(w, 1)
}

func blockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("block")
	p.WriteTo(w, 1)
}

func Start(addr string, internal time.Duration) {

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
		/*
			profiles.m = map[string]*Profile{
				"goroutine":    goroutineProfile,
				"threadcreate": threadcreateProfile,
				"heap":         heapProfile,
				"block":        blockProfile,
			}
		*/
		http.HandleFunc("/goroutine", goroutineHandler)
		http.HandleFunc("/threadcreate", threadHandler)
		http.HandleFunc("/heap", heapHandler)
		http.HandleFunc("/block", blockHandler)
		http.ListenAndServe(addr, nil)
	}()
}

func Stop() {
	running = false
	ticker.Stop()
	close(chMonitorStop)
	checkState()

	zed.ZLog("[ShutDown] Monitor Stop!")
}
