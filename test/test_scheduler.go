package main

import(
  "fmt"
  "runtime"
  "time"
)

func main() {
  runtime.GOMAXPROCS(1)
  
  print := func() {
    i := 0
    for {
      i++
      fmt.Println(i)
      time.Sleep(time.Second)
    }
  }
  
  call := func(i *int) {
    (*i)++
    (*i)--
  }
  calc := func() {
    i := 0
    for {
      /* 函数调用不是调度器goroutine切换的必要条件，要系统调用才会引起切换，linux/windows都测过，效果一致 */
      call(&i)
    }
  }
  
  go print()
  
  go calc()
  
  time.Sleep(time.Hour)
}
