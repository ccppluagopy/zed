package main

import(
  "log"
  "net/http"
  "os"
)
    
func main(){
    /* 静态文件 os 绝对路径 */
    wd, err := os.Getwd()
    if err != nil {
      log.Fatal(err)
    }
    /* 前缀去除，列出dir */
    http.Handle("/", 
      http.StripPrefix("/",
        http.FileServer(http.Dir(wd))))
        
    err = http.ListenAndServe("127.0.0.1:4000", nil)
    if err != nil {
      log.Fatal(err)
    }
        
  }
