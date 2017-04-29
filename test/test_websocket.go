package main

import(
  "fmt"
  "golang.org/x/net/websocket"
  "net/http"
  "time"
)

func echo(ws *websocket.Conn) {
  msg := make([]byte, 512)
  n, err = ws.Read(msg)
  if err != nil {
    fmt.Println("read err: ", err)
  }
  fmt.Printf("recv: %s\n", string(msg[:n]))
  n, err = ws.Write(msg[:n])
  if err != nil {
    fmt.Println("send err: ", err)
  }
  
  fmt.Println("send: %s\n", string(msg[:n]))
}

const(
  u1 = "ws://127.0.0.1:8080/echo"
  o1 = "http://127.0.0.1:8080/echo"
  u2 = "ws://127.0.0.1:8090/echo"
  o2 = "http://127.0.0.1:8090/echo"
)

func client(){
  clientecho := func(u string, o string){
    ws, err := websocket.Dial(u, "", o)
    if err != nil {
      fmt.Println("client dail err: ", err)
      return
    }
    msg := []byte("hello world")
    _, err = ws.Write(msg)
    if err != nil {
      fmt.Println("client send err: ", err)
      return
    }
    buf := make([]byte, 512)
    n, err2 = ws.Read(buf)
    if err2 != nil {
      fmt.Println("client read err: ", err)
      return
    }
    fmt.Printf("client recv: %s\n", string(buf[:n]))
  }
  time.Sleep(time.Second)
  go clientecho(u1, o1)
  go clientecho(u2, o2)
}
func main() {
  http.Handle("/echo", websocket.Handler(echo))
  
  go http.ListenAndServe(":8080", nil)
  http.ListenAndServe(":8090", nil)
  for {
    time.Sleep(time.Hour)
  }
}
