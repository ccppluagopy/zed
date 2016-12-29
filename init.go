package zed

var (
	workdir    = "./"
	logdir     = ""
	worklogdir = "./logs/"

	/*
		showClientData = false
		dataInSupervisor  func(*NetMsg) = nil
		dataOutSupervisor func(*NetMsg) = nil
	*/
)

func Init(workDir string, logDir string) {
	ZLog("Init, Working Dir:\"%s\", Log Dir:\"%s\"", workDir, logDir)
	workdir = workDir
	worklogdir = logDir
	MakeDir(logDir)

	//showClientData = showClient
	/*

		dataInSupervisor = inSupervisor
		dataOutSupervisor = outSupervisor
	*/
	MakeNewLogDir()

	/*if zlogfile == nil {
		zlogfile = CreateLogFile("ZLog")
		zlogfile.NewFile()
	}*/
	//os.Exit(0)
}
