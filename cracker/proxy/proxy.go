package proxy

import (
	"net"
	"time"
)

type proxyConn struct {
	remote net.Conn
	uuid   string
	close  chan struct{}
	heart  chan struct{}
}

func newProxyConn(remote net.Conn, uuid string) *proxyConn {
	return &proxyConn{remote: remote, uuid: uuid,
		close: make(chan struct{}),
		heart: make(chan struct{}),
	}
}

func (pc *proxyConn) Close() {
	select {
	case pc.close <- struct{}{}:
	default:

	}
}

func (pc *proxyConn) Heart() {
	select {
	case pc.heart <- struct{}{}:
	default:
	}
}

func (pc *proxyConn) Do() {

	defer pc.remote.Close()

	for {
		select {
		case <-time.After(time.Duration(time.Second * heartTTL)):
			return
		case <-pc.close:
			return
		case <-pc.heart:
			continue
		}
	}
}
