package main

import (
	"cracker/proxy"
	"flag"
	"io"
	"log"
	"net"
	"strconv"
)

func handleConn(conn net.Conn, raddr, secret string) {

	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}

	if buf[0] != 0x05 {
		//只处理Socks5协议
		log.Printf("only support sock5...\n")
		return
	}
	//客户端回应：Socks服务端不需要验证方式
	conn.Write([]byte{0x05, 0x00})
	n, _ = conn.Read(buf)
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
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("will connect %s ... \n", addr)
	lp, err := proxy.Connect(raddr, addr, secret)
	if err != nil {
		log.Println(err)
		return
	}
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //响应客户端连接成功
	//进行转发
	go io.Copy(conn, lp)
	io.Copy(lp, conn)
	close(lp.Close)
	lp.Quit()
	log.Printf("close connection with %s \n", conn.RemoteAddr().String())

}

func main() {
	laddr := flag.String("laddr", "", "listen addr")
	raddr := flag.String("raddr", "", "remote http url")
	secret := flag.String("secret", "", "secret key")
	flag.Parse()
	log.Printf("listen at:[%s]\n", *laddr)
	l, err := net.Listen("tcp", *laddr)
	if err != nil {
		log.Panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
		}
		go handleConn(conn, *raddr, *secret)

	}
}
