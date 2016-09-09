package zed

var (
	workdir    = "./"
	logdir     = ""
	worklogdir = "./log"
)

func Init(workDir string, logDir string) {
	workdir = workDir
	worklogdir = logDir

	MakeNewLogDir()

	/*if zlogfile == nil {
		zlogfile = CreateLogFile("ZLog")
		zlogfile.NewFile()
	}*/
	//os.Exit(0)
}
