package proxy

import (
	"net"
	"time"
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
}

func newProxyConn(remote net.Conn, uuid string) *proxyConn {
	return &proxyConn{remote: remote, uuid: uuid, readChannel: make(chan dataBodyTyp, 10),
		writeChannel: make(chan dataBodyTyp, 10),
		close:        make(chan bool),
	}
}

func (pc *proxyConn) work() {

	go func() {
		for {
			buf := make([]byte, 1024)
			pc.remote.SetReadDeadline(time.Now().Add(timeout * 1e9))
			n, err := pc.remote.Read(buf)
			if n > 0 {
				pc.readChannel <- dataBodyTyp{typ: DATA_TYP, body: buf[:n]}
			}
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					pc.readChannel <- dataBodyTyp{typ: HEART_TYP}
				} else {
					close(pc.close)
					pc.readChannel <- dataBodyTyp{typ: QUIT_TYP}
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
	defer pc.remote.Close()
}
