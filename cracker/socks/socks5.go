package socks

import (
	"io"
	"net"
	"strconv"

	"github.com/lovedboy/cracker/cracker/logger"
	"github.com/lovedboy/cracker/cracker/proxy"
)

var g = logger.GetLogger()

type Socks5 struct {
	Raddr  string
	Secret string
}

func (s *Socks5) HandleConn(msg []byte, conn net.Conn) {

	//客户端回应：Socks服务端不需要验证方式
	conn.Write([]byte{0x05, 0x00})
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	var host, port string
	switch buf[3] {
	case 0x01: //IP V4
		host = net.IPv4(buf[4], buf[5], buf[6], buf[7]).String()
	case 0x03: //域名
		host = string(buf[5 : n-2]) //b[4]表示域名的长度
	case 0x04: //IP V6
		host = net.IP{buf[4], buf[5], buf[6], buf[7], buf[8], buf[9], buf[10], buf[11], buf[12], buf[13], buf[14], buf[15], buf[16], buf[17], buf[18], buf[19]}.String()
	}
	port = strconv.Itoa(int(buf[n-2])<<8 | int(buf[n-1]))

	addr := net.JoinHostPort(host, port)

	lp, err := proxy.Connect(s.Raddr, addr, s.Secret)
	if err != nil {
		g.Errorf("proxy connect err:%s", err)
		return
	}
	g.Debugf("connect %s success", addr)
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //响应客户端连接成功
	//进行转发
	go func() {
		_, err := io.Copy(conn, lp)
		if err != nil && err != io.EOF {
			g.Debugf("read err: %s", err)
		}
		lp.CloseRead()
	}()
	io.Copy(lp, conn)
	lp.Close()
	g.Debugf("disconnect %s", addr)
}
