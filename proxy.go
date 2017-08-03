package cracker

import (
	"net"
	"sync"
	"time"
)

type proxyConn struct {
	remote net.Conn
	uuid   string
	close  chan struct{}
	heart  chan struct{}
	sync.Mutex
	hasClosed bool
}

func newProxyConn(remote net.Conn, uuid string) *proxyConn {
	return &proxyConn{remote: remote, uuid: uuid,
		close: make(chan struct{}),
		heart: make(chan struct{}),
	}
}

func (pc *proxyConn) Close() {
	pc.Lock()
	pc.hasClosed = true
	pc.Unlock()
	select {
	case pc.close <- struct{}{}:
	default:

	}
}

func (pc *proxyConn) IsClosed() bool {
	pc.Lock()
	defer pc.Unlock()
	return pc.hasClosed
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
		case <-time.After(time.Second * heartTTL):
			return
		case <-pc.close:
			return
		case <-pc.heart:
			continue
		}
	}
}
