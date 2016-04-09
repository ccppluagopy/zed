package zed

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	signalHandled = false
)

func HandleSignal(maskAll bool) {
	if !signalHandled {
		go func() {
			var (
				sig      os.Signal
				chSignal = make(chan os.Signal, 1)
			)

			handlemsg := func(sig string) {
				if maskAll {
					ZLog("Handle Signal %s!", sig)
				} else {
					ZLog("Exit By Signal %s!", sig)
					os.Exit(0)
				}
			}

			signal.Notify(chSignal, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

			for {
				sig = <-chSignal
				if sig == nil {
					return
				}

				switch sig {
				case syscall.SIGQUIT:
					handlemsg("SIGQUIT")

				case syscall.SIGTERM:
					handlemsg("SIGTERM")

				case syscall.SIGINT:
					handlemsg("SIGINT")

				case syscall.SIGHUP:
					handlemsg("SIGHUP")

				default:
				}
			}
		}()
		signalHandled = true
	}
}

/*
func Stop() {
	close(chSignal)
}*/
