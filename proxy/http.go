package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type httpProxy struct {
	addr string
}

func NewHttpProxy(addr string) *httpProxy {
	return &httpProxy{addr: addr}
}

func (hp *httpProxy) Listen() {
	http.HandleFunc("/proxy", hp.serveProxy)
	http.HandleFunc("/ping", hp.ping)
	log.Printf("listen at:[%s]\n", hp.addr)
	err := http.ListenAndServe(hp.addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (hp *httpProxy) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (hp *httpProxy) serveProxy(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("expected http.ResponseWriter to be an http.Flusher")
	}
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "image/jpeg")
	addr := r.Header.Get("proxy_addr")
	conn, err := net.DialTimeout("tcp", addr, time.Duration(time.Second*2))
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	log.Printf("connect %s success\n", addr)
	defer conn.Close()
	wait := make(chan bool, 1)
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if n != 0 {
				w.Write(buf[:n])
				flusher.Flush()
			}
			if err != nil {
				wait <- true
				break
			}
		}
	}()

	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := r.Body.Read(buf)
			if n != 0 {
				log.Printf("read:%s\n", string(buf[:n]))
				conn.Write(buf[:n])
			}
			if err != nil {
				wait <- true
				break
			}
		}
	}()
	for i := 0; i < 2; i++ {
		<-wait
	}
}

type localProxy struct {
	hpURL string
	addr  string
}

func NewLocalProxy(hpURL, addr string) *localProxy {
	return &localProxy{hpURL: hpURL, addr: addr}
}

func (lp *localProxy) Proxy(rd io.Reader) (res *http.Response, err error) {

	client := &http.Client{}

	req, err := http.NewRequest("POST", lp.hpURL, rd)
	if err != nil {
		return nil, err
	}

	req.ContentLength = -1
	req.Close = false
	req.TransferEncoding = []string{"chunked"}
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("Transfer-Encoding", "chunked")
	req.Header.Set("proxy_addr", lp.addr)
	return client.Do(req)
}
