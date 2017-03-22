package proxy

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"fmt"
	"net"

	"io/ioutil"

	"errors"

	"sync"


	"github.com/pborman/uuid"
	"gopkg.in/bufio.v1"
)

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
	bufSize  = 4096
)

type httpProxy struct {
	addr     string
	secret   string
	proxyMap map[string]*proxyConn
	sync.Mutex
	// buf size
	bs int
}

func NewHttpProxy(addr, secret string) *httpProxy {
	return &httpProxy{addr: addr,
		secret:   secret,
		proxyMap: make(map[string]*proxyConn),
		bs:       bufSize,
	}
}

func (hp *httpProxy) Listen() {
	http.HandleFunc(CONNECT, hp.connect)
	http.HandleFunc(PULL, hp.pull)
	http.HandleFunc(PUSH, hp.push)
	http.HandleFunc(PING, hp.ping)
	http.HandleFunc("/debug", hp.debug)
	log.Printf("listen at:[%s]\n", hp.addr)
	err := http.ListenAndServe(hp.addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (hp *httpProxy) SetBufSize(bs int) {
	hp.bs = bs
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
		WriteHTTPError(w, err.Error())
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
	uuid := r.Header.Get("UUID")
	hp.Lock()
	pc, ok := hp.proxyMap[uuid]
	hp.Unlock()
	if ok {
		select {
		case d := <-pc.readChannel:
			switch d.typ {
			case DATA_TYP:
				WriteHTTPData(w, d.body)
			case QUIT_TYP:
				WriteHTTPQuit(w, "quit")
			case HEART_TYP:
				WriteHTTPHeart(w, "alive")

			}
		case <-time.After(time.Duration(time.Second * heartTTL)):
			WriteHTTPError(w, "server timeout")
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
	log.Printf("Connecting to %s:...\n", addr)
	remote, err := net.DialTimeout("tcp", addr, time.Duration(time.Second*timeout))
	if err != nil {
		WriteHTTPError(w, fmt.Sprintf("Could not connect to %s", addr))
		return
	}
	log.Printf("Connect to %s: success ...\n", addr)
	proxyID := uuid.New()
	pc := newProxyConn(remote, proxyID)
	pc.SetBufSize(hp.bs)
	hp.Lock()
	hp.proxyMap[proxyID] = pc
	hp.Unlock()

	go func() {
		pc.work()
		remote.Close()
		log.Printf("close connection with %s ... \n", remote.RemoteAddr().String())
		<-time.After(time.Duration(time.Second * heartTTL))
		hp.Lock()
		log.Printf("delete uuid:%s ... \n", proxyID)
		delete(hp.proxyMap, proxyID)
		hp.Unlock()
	}()
	WriteHTTPOK(w, proxyID)
}

func (hp *httpProxy) debug(w http.ResponseWriter, r *http.Request) {
	hp.Lock()

	buf := bufio.Buffer{}
	for k, v := range hp.proxyMap {
		buf.WriteString(fmt.Sprintf("%s  %s\n", k, v.remote.RemoteAddr().String()))
	}
	hp.Unlock()
	WriteHTTPOK(w, buf.String())

}
