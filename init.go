package zed

var (
	workdir    = "./"
	logdir     = ""
	worklogdir = "./log"
)

func Init(wdir string, ldir string) {
	workdir = wdir
	worklogdir = ldir

	MakeNewLogDir()

	/*if zlogfile == nil {
		zlogfile = CreateLogFile("ZLog")
		zlogfile.NewFile()
	}*/
	//os.Exit(0)
}
