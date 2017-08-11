package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"time"
)
/*
	time.Async是我自己对go/src/time/sleep.go打的补丁，补丁代码在test/sleep.go中
	objmgr可以对比https://github.com/golang/go/blob/master/src/sync/map.go根据实际场景需要选择使用
*/


type Int int

func (i Int) HashIdx() int {
	return int(i)
}

func main() {
	om := zed.NewObjMgr(10)

	loop := 1000000

	setcor := func() {
		for i := 0; i < loop; i++ {
			om.Set(Int(i), i)
		}
	}

	t1 := time.Now()
	setcor()
	time.Async(func() {
		fmt.Println("size: ", om.Size())
		fmt.Println("time used: ", time.Since(t1).Seconds())
	})
	

	v, ok := om.Get(Int(100))
	fmt.Println(v, ok)
	v, ok = om.Get(Int(999999))
	fmt.Println(v, ok)
	v, ok = om.Get(Int(10000009))
	fmt.Println(v, ok)

	time.Sleep(time.Hour)
}
