package zed

import (
	"os"
	"os/signal"
	"syscall"
)

func Start(chStop chan string, maskAll bool) {
	chStop = chStop

	go func() {
		var (
			sig      os.Signal
			chSignal = make(chan os.Signal, 1)
		)

		handlemsg := func(sig string) {
			if maskAll {
				Printf("Handled Signal %s!", sig)
				continue
			}
			if chStop != nil {
				chStop <- fmt.Sprintf("Signal %s", sig)
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

func Stop() {
	//close(chSignal)
}
