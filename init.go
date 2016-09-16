package zed

var (
	workdir        = "./"
	logdir         = ""
	worklogdir     = "./log"
	showClientData = false
)

func Init(workDir string, logDir string, showClient bool) {
	ZLog("Init, Working Dir:\"%s\", Log Dir:\"%s\", Log Data In:\"%v\", Log Data Out:\"%v\"", workDir, logDir, showClientData)
	workdir = workDir
	worklogdir = logDir
	showClientData = showClient
	MakeNewLogDir()

	/*if zlogfile == nil {
		zlogfile = CreateLogFile("ZLog")
		zlogfile.NewFile()
	}*/
	//os.Exit(0)
}
