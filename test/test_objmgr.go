package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"time"
)

type Int int

func (i Int) HashIdx() int {
	return int(i)
}

func main() {
	om := zed.NewObjMgr(10)

	loop := 1000000

	getcor := func() {
		for i := 0; i < loop; i++ {
			//fmt.Println(om.Get(Int(i)) == i)
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

	fmt.Println(om.Size())

	v, ok := om.Get(Int(100))
	fmt.Println(v, ok)
	v, ok = om.Get(Int(999999))
	fmt.Println(v, ok)
	v, ok = om.Get(Int(10000009))
	fmt.Println(v, ok)

	time.Sleep(time.Hour)
}
