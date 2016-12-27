package zed

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	inited = false
)

func HandleSignal(maskAll bool, handler func(sig os.Signal)) {
	if !inited {
		NewCoroutine(func() {
			var (
				sig      os.Signal
				chSignal = make(chan os.Signal, 1)
			)

			handlemsg := func(sig string) {
				if maskAll {
					ZLog("Handle Signal %s!", sig)
				} else {
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

				if handler != nil {
					//ZLog("Handle Signal %s!", sig)
					handler(sig)
				}
			}
		})
		inited = true
	}
}

func WaitSignal(maskAll bool, handler func(sig os.Signal)) {
	if !inited {
		//NewCoroutine(func() {
		var (
			sig      os.Signal
			chSignal = make(chan os.Signal, 1)
		)

		handlemsg := func(sig string) {
			if maskAll {
				ZLog("Handle Signal %s!", sig)
			} else {
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

			if handler != nil {
				//ZLog("Handle Signal %s!", sig)
				handler(sig)
			}
		}
		//})
		inited = true
	}
}

/*
func Stop() {
	close(chSignal)
}*/
