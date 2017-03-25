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
	writeChannel chan dataBodyTyp
	close        chan bool
}

func newProxyConn(remote net.Conn, uuid string) *proxyConn {
	return &proxyConn{remote: remote, uuid: uuid,
		writeChannel: make(chan dataBodyTyp, 10),
		close:        make(chan bool),
	}
}

func (pc *proxyConn) work() {

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
					g.Debugf("write err: %s", err)
					return
				}
			}

		}
	}
}
