package main

import (
	"fmt"
	"github.com/ccppluagopy/zed/zhttp"
	"net"
	"net/http"
	"time"
)

func serve8888(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.Host + " " + r.URL.String() + " " + r.Method))
}

func serve9999(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.Host + " " + r.URL.String() + " " + r.Method))
	//r.Context().Done()
	r.Header.Add("key", "head-9999")
	fmt.Println("--------------------")
	fmt.Println("body: ", r.Response, r.Context())
	fmt.Println("====================")
}

func serverHttp(addr string, handler func(http.ResponseWriter, *http.Request)) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}

	server := &http.Server{Handler: http.HandlerFunc(handler)}
	server.Serve(l)
}

func main() {
	go zhttp.NewServer("127.0.0.1:8888", serve8888)
	zhttp.NewServer("127.0.0.1:9999", serve9999)

	time.Sleep(time.Hour)
}
