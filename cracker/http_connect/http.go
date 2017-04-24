package http_connect

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"net/url"
	"strings"

	"github.com/lovedboy/cracker/cracker/logger"
	"github.com/lovedboy/cracker/cracker/proxy"
)

var g = logger.GetLogger()

type HttpConnect struct {
	Raddr  string
	Secret string
}

func (h *HttpConnect) HandleConn(msg []byte, conn net.Conn) {

	var method, addr, host string
	fmt.Sscanf(string(msg[:bytes.IndexByte(msg, '\n')]), "%s%s", &method, &host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		g.Error(err)
		return
	}
	if strings.Index(hostPortURL.Host, ":") == -1 {
		if hostPortURL.Opaque == "443" {
			addr = hostPortURL.Scheme + ":443"
		} else {
			addr = hostPortURL.Host + ":80"
		}
	} else {
		addr = hostPortURL.Host
	}
	g.Debugf("will connect %s ... ", addr)
	lp, err := proxy.Connect(h.Raddr, addr, h.Secret)
	if err != nil {
		g.Errorf("proxy connect err:%s", err)
		return
	}
	if method == "CONNECT" {
		conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")) //响应客户端连接成功
	} else {
		lp.Write(msg)
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
}
