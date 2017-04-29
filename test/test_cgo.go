package main

/*
#include <stdio.h>

int Hello()
{
  printf("hello cgo\n");
}
*/

import "C"

import("fmt")

func main() {
  ret := C.Hello()
  fmt.Printf("\r%d%%\n", ret)
}
