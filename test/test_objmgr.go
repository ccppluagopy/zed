package main

import (
	"fmt"
	"github.com/ccppluagopy/zed/objmgr"
	"time"
)

type Int int

func (i Int) HashIdx() int {
	return int(i)
}

func main() {
	om := objmgr.NewObjMgr(10)

	loop := 1000000

	getcor := func() {
		for i := 0; i < loop; i++ {
			fmt.Println(om.Get(Int(i)) == i)
		}
	}

	setcor := func() {
		for i := 0; i < loop; i++ {
			om.Set(Int(i), i)
		}
	}

	setcor()

	go getcor()
	go setcor()
	go getcor()
	go setcor()
	go getcor()
	go setcor()

	time.Sleep(time.Hour)
}
