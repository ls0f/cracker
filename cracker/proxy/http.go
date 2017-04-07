package proxy

import (
	"net/http"
	"strconv"
	"time"

	"fmt"
	"net"

	"io/ioutil"

	"errors"

	"sync"

	"github.com/lovedboy/cracker/cracker/logger"

	"github.com/pborman/uuid"
)

var g = logger.GetLogger()

const (
	CONNECT = "/connect"
	PING    = "/ping"
	PULL    = "/pull"
	PUSH    = "/push"
)
const (
	DATA_TYP  = "data"
	QUIT_TYP  = "quit"
	HEART_TYP = "heart"
)

const (
	timeout  = 10
	signTTL  = 10
	heartTTL = 30
)

type httpProxy struct {
	addr     string
	secret   string
	proxyMap map[string]*proxyConn
	sync.Mutex
	https bool
}

func NewHttpProxy(addr, secret string, https bool) *httpProxy {
	return &httpProxy{addr: addr,
		secret:   secret,
		proxyMap: make(map[string]*proxyConn),
		https:    https,
	}
}

func (hp *httpProxy) Listen() {
	http.HandleFunc(CONNECT, hp.connect)
	http.HandleFunc(PULL, hp.pull)
	http.HandleFunc(PUSH, hp.push)
	http.HandleFunc(PING, hp.ping)
	g.Infof("listen at:[%s]", hp.addr)
	var err error
	if hp.https {
		err = http.ListenAndServeTLS(hp.addr, "cert.pem", "key.pem", nil)
	} else {
		err = http.ListenAndServe(hp.addr, nil)
	}
	if err != nil {
		g.Fatal("ListenAndServe: ", err)
	}
}

func (hp *httpProxy) verify(r *http.Request) error {
	ts := r.Header.Get("timestamp")
	sign := r.Header.Get("sign")
	tm, err := strconv.ParseInt(ts, 10, 0)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	if now-tm > signTTL {
		return errors.New("timestamp expire")
	}
	if VerifyHMACSHA1(hp.secret, ts, sign) {
		return nil
	} else {
		return errors.New("sign invalid")
	}
}

func (hp *httpProxy) before(w http.ResponseWriter, r *http.Request) error {
	err := hp.verify(r)
	if err != nil {
		g.Debug(err)
		WriteNotFoundError(w, "404")
	}
	return err
}

func (hp *httpProxy) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (hp *httpProxy) pull(w http.ResponseWriter, r *http.Request) {
	if err := hp.before(w, r); err != nil {
		return
	}
	flusher, _ := w.(http.Flusher)
	uuid := r.Header.Get("UUID")
	hp.Lock()
	pc, ok := hp.proxyMap[uuid]
	hp.Unlock()
	if ok {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Transfer-Encoding", "chunked")
		buf := make([]byte, 10240)
		for {
			flusher.Flush()
			n, err := pc.remote.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
			}
			if err != nil {
				g.Debugf("read err:%s", err)
				close(pc.close)
				return

			}
		}
	} else {
		WriteHTTPError(w, "uuid don't exist")
	}
}

func (hp *httpProxy) push(w http.ResponseWriter, r *http.Request) {
	if err := hp.before(w, r); err != nil {
		return
	}
	uuid := r.Header.Get("UUID")
	hp.Lock()
	pc, ok := hp.proxyMap[uuid]
	hp.Unlock()
	if ok {

		var d dataBodyTyp
		typ := r.Header.Get("TYP")
		if typ == HEART_TYP {
			d = dataBodyTyp{typ: HEART_TYP}
		} else if typ == QUIT_TYP {
			d = dataBodyTyp{typ: QUIT_TYP}

		} else {

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				WriteHTTPError(w, fmt.Sprintf("read err:%v", err))
				return
			}
			d = dataBodyTyp{typ: DATA_TYP, body: body}
		}

		select {
		case <-pc.close:
			WriteHTTPError(w, "server close conn")
		case pc.writeChannel <- d:
			WriteHTTPOK(w, "send")
		}
	} else {
		WriteHTTPError(w, "uuid don't exist")
	}
}

func (hp *httpProxy) connect(w http.ResponseWriter, r *http.Request) {

	if err := hp.before(w, r); err != nil {
		return
	}

	host := r.Header.Get("DSTHOST")
	port := r.Header.Get("DSTPORT")
	addr := fmt.Sprintf("%s:%s", host, port)
	g.Debugf("Connecting to %s:...", addr)
	remote, err := net.DialTimeout("tcp", addr, time.Duration(time.Second*timeout))
	if err != nil {
		WriteHTTPError(w, fmt.Sprintf("Could not connect to %s", addr))
		return
	}
	g.Debugf("Connect to %s: success ...", addr)
	proxyID := uuid.New()
	pc := newProxyConn(remote, proxyID)
	hp.Lock()
	hp.proxyMap[proxyID] = pc
	hp.Unlock()

	go func() {
		pc.work()
		remote.Close()
		g.Debugf("close connection with %s ... ", remote.RemoteAddr().String())
		<-time.After(time.Duration(time.Second * heartTTL))
		hp.Lock()
		g.Debugf("delete uuid:%s ... \n", proxyID)
		delete(hp.proxyMap, proxyID)
		hp.Unlock()
	}()
	WriteHTTPOK(w, proxyID)
}
