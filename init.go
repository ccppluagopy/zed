package zed

var (
	workdir    = "./"
	logdir     = ""
	worklogdir = "./log"
	logDataIn  = false
	logDataOut = false
)

func Init(workDir string, logDir string, showDataIn bool, showDataOut bool) {
	workdir = workDir
	worklogdir = logDir
	logDataIn = showDataIn
	logDataOut = showDataOut
	MakeNewLogDir()

	/*if zlogfile == nil {
		zlogfile = CreateLogFile("ZLog")
		zlogfile.NewFile()
	}*/
	//os.Exit(0)
}
