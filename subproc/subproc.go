package subproc

import(
	"github.com/ccppluagopy/zed"
	"os"
	"os/exec"
	"runtime"
)

type SubProcNode struct {
	StartProc string
	StopProc string
	Detach bool
	StartArgs []string
	StopArgs []string
	Cmd *exec.Cmd
}

func (spnode *SubProcNode) Start() {
	zed.NewCoreoutine(func() {
		path, _ := exec.LookPath(spnode.StartProc)

		spnode.Cmd = exec.Command(path, spnode.StartArgs...)

		spnode.Cmd.Stderr = os.Stderr
		spnode.Cmd.Stdin = os.Stdin
		spnode.Cmd.Stdout = os.Stdout

		if err := spnode.Cmd.Run(); err != nil {
			cmdstr := spnode.StartProc
			for i := 0; i < len(spnode.StartArgs); i++ {
				cmdstr += (" " + spnode.StartArgs[i])
			}
			zed.ZLog("SubProcNode(%s) Start Err: %s", cmdstr, err.Error())
		}
		spnode.Cmd.Process.Wat()
	})
}
	

func (spnode *SubProcNode) Stop() {
	zed.NewCoroutine(func() {
		cmd := exec.Command(spnode.StopProc, spnode.StopArgs)
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			cmdstr := spnode.StopProc
			for i := 0; i < len(spnode.StopArgs); i++ {
				cmdstr += (" " + spnode.StopArgs[i])
			}
			zed.ZLog("SubProcNode(%s) Stop Err: %s", spnode.StopProc, err.Error())
		}
	})
}

func NewSubProcNode(start string, startargs[]string, stop string, stopargs[]string) *SubProcNode {
	spnode := &SubProcNode{
		StartProc: start,
		StopProc: stop,
		Detach: true,
		StartArgs: startargs,
		StopArgs: stopargs,
		Cmd: nil,
	}

	if startargs == nil {
		spnode.StartArgs = []string{}
	}
	if stopargs == nil {
		spnode.StopArgs = []string{}
	}

	if spnode.Detach {
		spnode.StartArgs = append([]string{spnode.StartProc}, spnode.StartArgs...)
		if runtime.GOOS == "windows" {
			spnode.StartProc = "forkwin.bat"
		} else {
			spnode.StartProc = "./forklinux.bin"
		}
		spnode.StopArgs = append([]string{spnode.StopProc}, spnode.StopArgs...)
		if runtime.GOOS == "windows" {
			spnode.StopProc = "forkwin.bat"
		} else {
			spnode.StopProc = "./forklinux.bin"
		}
	}

	return spnode
}


