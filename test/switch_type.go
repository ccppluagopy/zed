func check(i interface{}) {
  switch v:=i.(type) {
  case int16,int32,int64:
    fmt.Println(V, unsafe.Sizeof(v))
  }
}
