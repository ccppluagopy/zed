package main

import(
  "fmt"
  "runtime"
  "time"
)

func main() {
  for i:=0; i < 10; i++ {
    n := time.Duration(i)
    time.AfterFunc(1, func() {
      time.Sleep(time.Second * n)
    })
  }
  for {
    fmt.Println("goroutine num: ", runtime.NumGoroutine())
    time.Sleep(time.Second)
  }
}
