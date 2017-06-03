package main 

import(
  "fmt"
)

type Test struct {
  a int
  b string
}

func main() {
  test := Test{3,"hehe"}
  fmt.Println("test 1: %+v", test)
  fmt.Println("test 2: %#v", test)
}
