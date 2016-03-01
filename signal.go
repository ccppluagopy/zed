package zed

import (
	"os"
	"os/signal"
	"syscall"
)

func HandleSignal(maskAll bool) {
	go func() {
		var (
			sig      os.Signal
			chSignal = make(chan os.Signal, 1)
		)

		handlemsg := func(sig string) {
			if maskAll {
				Printf("Handled Signal %s!", sig)
			} else {
				LogInfo(LOG_IDX, LOG_IDX, "Exit By Signal %s!", sig)
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
}

/*
func Stop() {
	close(chSignal)
}*/
