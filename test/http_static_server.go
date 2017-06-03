package main

import(
  "log"
  "net/http"
  "os"
)

func main() {
  wd, err := os.Getwd()
  if err != nil {
    log.Fatal(err)
  }
  
  http.Handle("/",
    http.StripPrefix("/",
      http.FileServer(http.Dir(wd))))
      
  err = http.ListenAndServe(":8080"), nil)
  if err != nil {
    log.Fatal(err)
  }
}
