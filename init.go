package zed

var (
	workdir    = "./"
	logdir     = ""
	worklogdir = "./log"
	logDataIn  = false
	logDataOut = false
)

func Init(workDir string, logDir string, showDataIn bool, showDataOut bool) {
	ZLog("Init, Working Dir:\"%s\", Log Dir:\"%s\", Log Data In:\"%v\", Log Data Out:\"%v\"", workDir, logDir, showDataIn, showDataOut)
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
