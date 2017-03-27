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
	Args []string
	Cmd *exec.Cmd
}

func (spnode *SubProcNode) Start() {
	zed.NewCoreoutine(func() {
		path, _ := exec.LookPath(spnode.StartProc)

		spnode.Cmd = exec.Command(path, spnode.Args...)

		spnode.Cmd.Stderr = os.Stderr
		spnode.Cmd.Stdin = os.Stdin
		spnode.Cmd.Stdout = os.Stdout

		if err := spnode.Cmd.Run(); err != nil {
			cmdstr := spnode.StartProc
			for i := 0; i < len(spnode.Args); i++ {
				cmdstr += (" " + spnode.Args[i])
			}
			zed.ZLog("SubProcNode(%s) Start Err: %s", cmdstr, err.Error())
		}
		spnode.Cmd.Process.Wat()
	})
}
	

func (spnode *SubProcNode) Stop() {
	zed.NewCoroutine(func() {
		cmd := exec.Command(spnode.StopProc)
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout

		if err := spnode.Cmd.Run(); err != nil {
			zed.ZLog("SubProcNode(%s) Stop Err: %s", spnode.StopProc, err.Error())
		}
	})
}

func NewSubProcNode(start string, stop string, args[]string) *SubProcNode {
	spnode := &SubProcNode{
		StartProc: start,
		StopProc: stop,
		Detach: true,
		Args: args,
		Cmd: nil,
	}

	if args == nil {
		spnode.Args = []string{}
	}

	if spnode.Detach {
		spnode.Args = append([]string{spnode.StartProc}, spnode.Args...)
		if runtime.GOOS == "windows" {
			spnode.StartProc = "forkwin.bat"
		} else {
			spnode.StartProc = "./forklinux.bin"
		}
	}

	return spnode
}


