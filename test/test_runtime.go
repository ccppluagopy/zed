package main 

import(
	"fmt"
	"runtime"
	"time"
)


func print() {
	i := 0
	for {
		i++
		fmt.Println(i)
		time.Sleep(time.Second)
	}		
}

func call(i *int) {
	i := 0
	for {
		(*i)++
		(*i)--
	}		
}

func calc(){
		i := 0
		for {
			/* 此处不会出让cpu，不触发goroutine调度，所以因为设置了使用单核，这里执行后 go print() 的地方就不会打印 */
			call(&i)
		}
}

func main() {
	runtime.GOMAXPROCS(1)

	go print()
	go calc()

	time.Sleep(time.Hour)
}
