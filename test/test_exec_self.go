package main

import (
	"fmt"
	_ "net/http/pprof"
	//"time"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
)

func copy(src string, dst string) {
	if file, err := os.OpenFile(src, os.O_RDONLY, 0777); err == nil {
		if newfile, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0777); err == nil {
			buf, _ := ioutil.ReadAll(file)
			if _, err := newfile.Write(buf); err == nil {
				fmt.Println("copy success: ", src+" ==>> "+dst)
			}
		} else {
			fmt.Println("open dst err:", err)
		}
	} else {
		fmt.Println("open src err:", err)
	}
}

func main() {
	var cmd *exec.Cmd
	binname := os.Args[0]
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

	fmt.Println(binname, str, err)
}
