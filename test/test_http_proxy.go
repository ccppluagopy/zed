package main

/*
import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Panic(err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}

		go handleClientRequest(client)
	}
}

func handleClientRequest(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}
	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}

	if hostPortURL.Opaque == "443" {
		address = hostPortURL.Scheme + ":443"
	} else {
		if strings.Index(hostPortURL.Host, ":") == -1 {
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
	}

	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n")
	} else {
		server.Write(b[:n])
	}

	go io.Copy(server, client)
	io.Copy(client, server)
}
*/

import (
	//"bytes"
	//"fmt"
	"io"
	"log"
	//"net"
	"net/http"
	//"net/url"
	"strings"
)

func main() {
	localHost := "127.0.0.1:9001"
	targetHost := "127.0.0.1:8001"
	httpsServer(localHost, targetHost)
	log.Println("http server down!!!")
}

func httpsServer(addr string, remote_addr string) {

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		cli := &http.Client{}
		body := make([]byte, 0)
		n, err := io.ReadFull(req.Body, body)
		if err != nil {
			io.WriteString(w, "Request Data Error")
			return
		}
		reqUrl := "http://" + remote_addr + req.URL.Path

		req2, err := http.NewRequest(req.Method, reqUrl, strings.NewReader(string(body)))
		if err != nil {
			io.WriteString(w, "Request Error")
			return
		}
		// set request content type
		contentType := req.Header.Get("Content-Type")
		req2.Header.Set("Content-Type", contentType)
		// request
		rep2, err := cli.Do(req2)
		if err != nil {
			io.WriteString(w, "Not Found!")
			return
		}
		defer rep2.Body.Close()
		n, err = io.ReadFull(rep2.Body, body)
		if err != nil {
			io.WriteString(w, "Request Error")
			return
		}
		// set response header
		for k, v := range rep2.Header {
			w.Header().Set(k, v[0])
		}
		io.WriteString(w, string(body[:n]))
	})
	var err error = nil
	err = http.ListenAndServe(":12307", nil)
	if err != nil {
		log.Fatal("server down!!!")
	}
}
