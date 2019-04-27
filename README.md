```golang
package main

import (
	"fmt"
	"sort"
)

type Num struct {
	Value int64
	Count int64
	Text  string
}

func (num *Num) Desc() string {
	if num.Count == 1 {
		return fmt.Sprintf("%v", num.Value)
	}
	return "(" + num.Text + ")"
}

func MakeNums(nums map[Num]bool, num Num) map[Num]bool {
	ret := map[Num]bool{}
	for v, _ := range nums {
		if v.Value+num.Value >= 0 {
			ret[Num{v.Value + num.Value, v.Count + num.Count, fmt.Sprintf("%s+%s", v.Desc(), num.Desc())}] = true
		}
		if v.Value-num.Value >= 0 {
			ret[Num{v.Value - num.Value, v.Count + num.Count, fmt.Sprintf("%s-%s", v.Desc(), num.Desc())}] = true
		}
		if num.Value-v.Value >= 0 {
			ret[Num{num.Value - v.Value, v.Count + num.Count, fmt.Sprintf("%s-%s", num.Desc(), v.Desc())}] = true
		}
		if v.Value*num.Value >= 0 {
			ret[Num{v.Value * num.Value, v.Count + num.Count, fmt.Sprintf("%s*%s", v.Desc(), num.Desc())}] = true
		}
		if num.Value != 0 && v.Value%num.Value == 0 {
			ret[Num{v.Value / num.Value, v.Count + num.Count, fmt.Sprintf("%s/%s", v.Desc(), num.Desc())}] = true
		}
		if v.Value != 0 && num.Value%v.Value == 0 {
			ret[Num{num.Value / v.Value, v.Count + num.Count, fmt.Sprintf("%s/%s", num.Desc(), v.Desc())}] = true
		}
	}
	return ret
}

func main() {
	all := map[int64][]Num{
		1: []Num{Num{9, 1, "9"}},
	}

	ret := map[Num]bool{Num{9, 1, "9"}: true}
	for i := 1; i < 9; i++ {
		tmp := MakeNums(ret, Num{9, 1, "9"})
		for v, _ := range tmp {
			all[v.Count] = append(all[v.Count], v)
		}

		ret = tmp
	}
	mInt := map[int64]Num{}
	for i := int64(1); i < 5; i++ {
		for _, v1 := range all[i] {
			for _, v2 := range all[9-i] {
				if v1.Value+v2.Value >= 0 {
					mInt[v1.Value+v2.Value] = Num{v1.Value + v2.Value, v1.Count + v2.Count, fmt.Sprintf("%s+%s", v1.Desc(), v2.Desc())}
				}
				if v1.Value-v2.Value >= 0 {
					mInt[v1.Value-v2.Value] = Num{v1.Value - v2.Value, v1.Count + v2.Count, fmt.Sprintf("%s-%s", v1.Desc(), v2.Desc())}
				}
				if v2.Value-v1.Value >= 0 {
					mInt[v2.Value-v1.Value] = Num{v2.Value - v1.Value, v1.Count + v2.Count, fmt.Sprintf("%s-%s", v2.Desc(), v1.Desc())}
				}
				if v1.Value*v2.Value >= 0 {
					mInt[v2.Value*v1.Value] = Num{v1.Value * v2.Value, v1.Count + v2.Count, fmt.Sprintf("%s*%s", v1.Desc(), v2.Desc())}
				}
				if v2.Value != 0 && v1.Value%v2.Value == 0 {
					mInt[v1.Value/v2.Value] = Num{v1.Value / v2.Value, v1.Count + v2.Count, fmt.Sprintf("%s/%s", v1.Desc(), v2.Desc())}
				}
				if v1.Value != 0 && v2.Value%v1.Value == 0 {
					mInt[v2.Value/v1.Value] = Num{v2.Value / v1.Value, v1.Count + v2.Count, fmt.Sprintf("%s/%s", v2.Desc(), v1.Desc())}
				}
			}
		}
	}

	idx := 0
	arr := make([]int64, len(mInt))
	for v, _ := range mInt {
		arr[idx] = v
		idx++
	}
	//fmt.Println("len(arr):", len(arr))

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	//fmt.Println(len(arr))

	//i是想得到的表达式的值
	fmt.Println("----------------------")
	for i := int64(0); i < 10; i++ {
		if v, ok := mInt[i]; ok {
			fmt.Printf("%d: %s\n", i, v.Desc())
		}
	}

	fmt.Println("----------------------")
	for i, v := range arr {
		if int64(i) != v {
			//结果：无法表达的最小的自然数
			fmt.Println("result:", i)
			break
			//return
		}
	}
}
```

## Todo
增加SendData, msg.Encode返回buf？

## Install
* go get github.com/ccppluagopy/zed
* go get github.com/ccppluagopy/zed/timer
* go get github.com/ccppluagopy/zed/loadbalance
* go get github.com/ccppluagopy/zed/observer
 
## samples
* [timer](https://github.com/ccppluagopy/zed/blob/master/test/test_timer.go)
* [zsync](https://github.com/ccppluagopy/zed/blob/master/test/test_zsync.go)
* [logger](https://github.com/ccppluagopy/zed/blob/master/test/test_logger.go)
* [observer](https://github.com/ccppluagopy/zed/blob/master/test/test_observer.go)
* [loadbalance](https://github.com/ccppluagopy/zed/blob/master/test/test_loadbalance.go)


## sample tcpserver
```go
package main

import (
	"github.com/ccppluagopy/zed"
)
 
func main() {
	server := zed.NewTcpServer("testtcpserver")
	server.Start("127.0.0.1:10086")
}
```

## sample logger
```go
package main

import (
	"github.com/ccppluagopy/zed"
)

func main() {
	workDir := "./"
	logDir := "./logs/"

	const (
		TagZed = iota
		Tag1
		Tag2
	)

	var LogTags = map[int]string{
		TagZed: "--zed", /*'--'开头则关闭*/
		Tag1:   "Tag1",
		Tag2:   "Tag2",
	}

	var LogConf = map[string]int{
		"Info":         zed.LogFile,
		"Warn":         zed.LogCmd,
		"Error":        zed.LogCmd,
		"Action":       zed.LogCmd,
		"InfoCorNum":   2,
		"WarnCorNum":   3,
		"ErrorCorNum":  4,
		"ActionCorNum": 5,
	}

	zed.Init(workDir, logDir)
	zed.StartLogger(LogConf, LogTags, true)

	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			zed.LogInfo(Tag1, i, "test log i: %d", i)
		} else {
			zed.LogInfo(Tag2, i, "test log i: %d", i)
		}
	}

	zed.StopLogger()
}
```

## sample timer
```go
package main

import (
	"fmt"
	"github.com/ccppluagopy/zed/timer"
	"time"
)

func main() {
	t0 := time.Now()

	tm := timer.NewTimer()
	var (
		item1  *timer.TimeItem
		item3  *timer.TimeItem
		item5  *timer.TimeItem
		item8  *timer.TimeItem
		item10 *timer.TimeItem
	)

	for i := 0; i < 5; i++ {
		n := i*2 + 1
		str := fmt.Sprintf("%02d - ", n)
		item := tm.NewItem(time.Second*time.Duration(n), func() {
			fmt.Println(str, time.Since(t0).Seconds())
		})
		if n == 1 {
			item1 = item
		}
		if n == 3 {
			item3 = item
		}
		if n == 5 {
			item5 = item
		}
		if n == 8 {
			item8 = item
		}
		if n == 10 {
			item10 = item
		}
	}

	for i := 0; i < 5; i++ {
		n := (i + 1) * 2
		str := fmt.Sprintf("%02d - ", n)
		item := tm.NewItem(time.Second*time.Duration(n), func() {
			fmt.Println(str, time.Since(t0).Seconds())
		})
		if n == 3 {
			item3 = item
		}
		if n == 5 {
			item5 = item
		}
		if n == 8 {
			item8 = item
		}
		if n == 10 {
			item10 = item
		}
	}

	fmt.Println("000 Size: ", tm.Size())

	tm.DeleteItem(item3)
	tm.DeleteItem(item5)
	tm.DeleteItem(item1)
	tm.DeleteItem(item10)
	tm.DeleteItem(item8)
	fmt.Println("111 Size: ", tm.Size())

	time.Sleep(time.Second * 10)

	fmt.Println("222 Size: ", tm.Size())

	n := 0
	var scheduleItem *timer.TimeItem
	scheduleItem = tm.Schedule(time.Second*3, time.Second, func() {
		n++
		fmt.Println("Schedule: ", n, "pass:", time.Since(t0).Seconds())
		if n >= 5 {
			tm.DeleteItemInCall(scheduleItem)
		}

	})

	fmt.Println("333 Size: ", tm.Size())

	time.Sleep(time.Second * 10)

	fmt.Println("444 Size: ", tm.Size())

	fmt.Println("Over!")
}
```


## sample zsync
```go
package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"time"
)

func main() {

	mt := zed.Mutex{}
	zed.SetMutexDebug(true, time.Second*3)
	fmt.Println("111")
	mt.Lock()
	fmt.Println("222")
	mt.Lock()
	fmt.Println("333")
	time.Sleep(time.Second * 5)
}
```

## sample loadbalance
```go
package main

import (
	"fmt"
	lb "github.com/ccppluagopy/zed/loadbalance"
	"time"
)

func client(n int, stype string, stag string, addr string, addr2 string) {
	go func() {
		time.Sleep(time.Second)
		client := lb.NewLoadbalanceClient(addr)
		client.AddServer(stype, stag, addr2)
		client.UpdateLoad(stype, stag, n)
		time.Sleep(time.Second * 2)
		ret := client.GetMinLoadServerInfoByType(stype)
		fmt.Println(stype, stag, addr, " -- over ---", ret.Addr, ret.Num)
	}()
}

func main() {
	addr := "127.0.0.1:8888"
	server := lb.NewLoadbalanceServer(addr, time.Second*10)

	client(1, "test server", "test server 1", addr, "0.0.0.0:1")
	client(2, "test server", "test server 2", addr, "0.0.0.0:2")
	client(3, "test server", "test server 3", addr, "0.0.0.0:3")
	lb.StartLBServer(server)
	time.Sleep(time.Second * 3)
	fmt.Println("Over!")
}
```

## sample observer-cluster
```go
package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"github.com/ccppluagopy/zed/observer"
	"time"
)

func xxx(addr string, data string, n int) {
	for i := 0; i < 2; i++ {
		selfkey := fmt.Sprintf("%s-%d", addr, i)
		go func() {
			obsc := observer.NewOBClient(addr, selfkey, time.Second*20)
			if obsc == nil {
				fmt.Println("obsc is nil.....")
				return
			}

			event := "chat"
			obsc.Regist("xx", event, func(e interface{}, args []interface{}) {
				msg, ok := args[0].([]byte)
				if ok {
					fmt.Printf("--- %s recv chatMsg: %s %s\n", selfkey, e, string(msg))
				} else {
					fmt.Printf("xxx %s recv chatMsg: %s %s\n", selfkey, e, string(msg))
				}
			})

			time.Sleep(time.Second * 1)
			obsc.PublishAll(event, []byte(selfkey))
			time.Sleep(time.Second * 1)
			obsc.Stop()
		}()
	}
}

func main() {
	zed.SetMutexDebug(true)
	mgrAddr := "127.0.0.1:6666"
	nodeAddr1 := "127.0.0.1:7777"
	nodeAddr2 := "127.0.0.1:8888"
	nodeAddr3 := "127.0.0.1:9999"
	go observer.NewOBClusterMgr(mgrAddr)
	go observer.NewOBClusterNode(mgrAddr, nodeAddr1, time.Second*25).Start()
	go observer.NewOBClusterNode(mgrAddr, nodeAddr2, time.Second*25).Start()
	go observer.NewOBClusterNode(mgrAddr, nodeAddr3, time.Second*25).Start()

	time.Sleep(time.Second * 3)
	go xxx(nodeAddr1, "node 111", 1)
	go xxx(nodeAddr2, "node 222", 2)
	go xxx(nodeAddr3, "node 333", 3)

	time.Sleep(time.Hour)
}

```
