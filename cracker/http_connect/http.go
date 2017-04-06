package http_connect

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/lovedboy/cracker/cracker/logger"
	"github.com/lovedboy/cracker/cracker/proxy"
)

var g = logger.GetLogger()

type httpConnect struct {
	raddr  string
	secret string
	wait   chan bool
}

func (h *httpConnect) handleConn(conn net.Conn) {

	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		g.Errorf("read err:%s", err)
		return
	}
	var method, addr string
	fmt.Sscanf(string(buf[:bytes.IndexByte(buf[:n], '\n')]), "%s%s", &method, &addr)
	g.Debugf("will connect %s ... ", addr)
	lp, err := proxy.Connect(h.raddr, addr, h.secret)
	if err != nil {
		g.Errorf("proxy connect err:%s", err)
		return
	}
	if method == "CONNECT" {
		conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")) //响应客户端连接成功
	} else {
		conn.Write(buf[:n])
	}
	//进行转发
	go func() {
		_, err := io.Copy(conn, lp)
		if err != nil {
			g.Debugf("read err: %s", err)
		}
		lp.CloseRead()
	}()
	io.Copy(lp, conn)
	lp.Close()
	g.Debugf("close connection with %s", conn.RemoteAddr().String())

}

func (h *httpConnect) Wait() {
	<-h.wait
}

func NewHttpConnect(addr, raddr, secret string) (h *httpConnect, err error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	g.Infof("http proxy listen at:[%s]", addr)
	h = &httpConnect{
		raddr:  raddr,
		secret: secret,
		wait:   make(chan bool),
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				g.Errorf("accept err:%s", err)
			}
			go h.handleConn(conn)

		}
	}()
	return h, nil
}
