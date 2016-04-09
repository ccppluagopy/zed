package zed

import (
	"os"
)

func init() {
	workdir = "./"

	if len(os.Args) >= 2 {
		workdir = os.Args[1]

	}

	MakeNewLogDir()

	if zlogfile == nil {
		zlogfile = CreateLogFile("ZLog")
		zlogfile.NewFile()
	}
	//os.Exit(0)
}
