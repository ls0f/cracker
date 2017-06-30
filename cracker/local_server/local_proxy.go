package local_server

import (
	"net"

	"io"

	"github.com/lovedboy/cracker/cracker/http_connect"
	"github.com/lovedboy/cracker/cracker/logger"
	"github.com/lovedboy/cracker/cracker/socks"
)

var g = logger.GetLogger()

type localProxyServer struct {
	socks5 *socks.Socks5
	http   *http_connect.HttpConnect
	wait   chan bool
}

func (s *localProxyServer) handleConn(conn net.Conn) {

	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			g.Errorf("read err:%s", err)
		}
		return
	}

	if buf[0] == 0x04 {
		g.Warning("not support socks4")
	} else if buf[0] == 0x05 {
		s.socks5.HandleConn(buf[:n], conn)
	} else {
		s.http.HandleConn(buf[:n], conn)
	}
	g.Debugf("close connection with %s", conn.RemoteAddr().String())

}

func (s *localProxyServer) Wait() {
	<-s.wait
}

func NewLocalProxyServer(addr, raddr, secret string) (s *localProxyServer, err error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	g.Infof("socks5/http proxy listen at:[%s]", addr)
	s = &localProxyServer{
		socks5: &socks.Socks5{Raddr: raddr, Secret: secret},
		http:   &http_connect.HttpConnect{Raddr: raddr, Secret: secret},
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				g.Errorf("accept err:%s", err)
			}
			go s.handleConn(conn)

		}
	}()
	return s, nil
}
