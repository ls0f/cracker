package proxy

import (
	"net"
	"time"
)

const (
	sendTimeOut = 15
)

type dataBodyTyp struct {
	typ  string
	body []byte
}

type proxyConn struct {
	remote       net.Conn
	uuid         string
	readChannel  chan dataBodyTyp
	writeChannel chan dataBodyTyp
	close        chan bool
	// buf size
	bs int
}

func newProxyConn(remote net.Conn, uuid string) *proxyConn {
	return &proxyConn{remote: remote, uuid: uuid, readChannel: make(chan dataBodyTyp, 10),
		writeChannel: make(chan dataBodyTyp, 10),
		close:        make(chan bool),
		bs:           bufSize,
	}
}

func (pc *proxyConn) SetBufSize(bs int) {
	pc.bs = bs
}

func (pc *proxyConn) sendMsg(body dataBodyTyp) {
	select {
	case pc.readChannel <- body:
	case <-time.After(time.Second * sendTimeOut):
	}
}

func (pc *proxyConn) work() {

	go func() {
		for {
			buf := make([]byte, pc.bs)
			pc.remote.SetReadDeadline(time.Now().Add(timeout * time.Second))
			n, err := pc.remote.Read(buf)
			if n > 0 {
				pc.sendMsg(dataBodyTyp{typ: DATA_TYP, body: buf[:n]})
			}
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					pc.sendMsg(dataBodyTyp{typ: HEART_TYP})
				} else {
					close(pc.close)
					pc.sendMsg(dataBodyTyp{typ: QUIT_TYP})
					return

				}
			}
		}
	}()

	for {
		select {
		case <-time.After(time.Duration(time.Second * heartTTL)):
			return
		case <-pc.close:
			return
		case d := <-pc.writeChannel:
			switch d.typ {
			case QUIT_TYP:
				return
			case HEART_TYP:
				continue
			case DATA_TYP:
				if _, err := pc.remote.Write(d.body); err != nil {
					return
				}
			}

		}
	}
}
