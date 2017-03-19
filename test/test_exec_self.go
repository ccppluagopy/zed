package main

import (
	"fmt"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func copy(src string, dst string) {
	if file, err := os.OpenFile(src, os.O_RDONLY, 0777); err == nil {
		if newfile, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0777); err == nil {
			buf, _ := ioutil.ReadAll(file)
			if _, err := newfile.Write(buf); err == nil {
				fmt.Println("copy success: ", src+" ==>> "+dst)
			}
			newfile.Close()
		} else {
			fmt.Println("open dst err:", err)
		}
		file.Close()
	} else {
		fmt.Println("open src err:", err)
	}
}

const (
	origname = "test.exe"
)

func main() {
	var cmd *exec.Cmd
	binname := os.Args[0]

	fmt.Println("::::", binname)

	if origname == binname {
		copyname := "./tmp_" + binname
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", "del", copyname)
			//cmd = exec.Command("tasklist")
		} else {
			cmd = exec.Command("rm", copyname)
		}

		buf, err := cmd.Output()
		str := string(buf)

		copy(binname, copyname)
		fmt.Println(111)

		/*cmd = exec.Command(copyname)
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		fmt.Println(222)
		cmd.Run()
		fmt.Println("----:", copyname, str, err)
		*/
		args := []string{"ls", "-a", "-l", "-h"}
		env := os.Environ()
		execErr := syscall.Exec(copyname, args, env)
		if execErr != nil {
			panic(execErr)
		}
	}

	newfile, _ := os.OpenFile("./log.md", os.O_CREATE|os.O_RDWR, 0777)

	n := 0
	for {
		n++
		newfile.WriteString(fmt.Sprintf("%d: %s\n", n, binname))
		fmt.Println(n, binname)
		time.Sleep(time.Second * 2)
	}

}
