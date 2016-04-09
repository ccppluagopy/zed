package zed

func Init(wdir string) {
	workdir = wdir

	MakeNewLogDir()

	if zlogfile == nil {
		zlogfile = CreateLogFile("ZLog")
		zlogfile.NewFile()
	}
	//os.Exit(0)
}
