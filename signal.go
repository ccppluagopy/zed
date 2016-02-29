package zed

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	maskAllSig bool
	chAppStop  chan string
	chSignal   = make(chan os.Signal, 1)
)

func Start(chStop chan string, maskAll bool) {
	chAppStop = chStop
	maskAllSig = maskAll

	go func() {
		var sig os.Signal

		signal.Notify(chSignal, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

		for {
			sig = <-chSignal
			if sig == nil {
				return
			}

			switch sig {
			case syscall.SIGQUIT:
				if maskAllSig {
					Println("Handled Signal SIGQUIT!")
					continue
				}
				chAppStop <- "Signal SIGQUIT"
			case syscall.SIGTERM:
				if maskAllSig {
					Println("Handled Signal SIGTERM!")
					continue
				}
				chAppStop <- "Signal SIGTERM"
			case syscall.SIGINT:
				if maskAllSig {
					Println("Handled Signal SIGINT!")
					continue
				}
				chAppStop <- "Signal SIGINT"
			case syscall.SIGHUP:
				if maskAllSig {
					Println("Handled Signal SIGHUP!")
					continue
				}
				chAppStop <- "Signal SIGHUP"
			default:
			}
		}
	}()
}

func Stop() {
	//close(chSignal)
}
