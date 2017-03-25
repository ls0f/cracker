package main

import (
	"flag"
	"fmt"
	"logger"
	"os"
	"socks"
	"proxy"
)

var (
	GitTag    = "2000.01.01.release"
	BuildTime = "2000-01-01T00:00:00+0800"
)

func main() {
	laddr := flag.String("laddr", "", "listen addr")
	raddr := flag.String("raddr", "", "remote http url(e.g, https://example.com)")
	secret := flag.String("secret", "", "secret key")
	debug := flag.Bool("debug", false, "debug mode")
	version := flag.Bool("v", false, "version")
	flag.Parse()

	if *version {
		fmt.Printf("GitTag: %s \n", GitTag)
		fmt.Printf("BuildTime: %s \n", BuildTime)
		os.Exit(0)
	}
	logger.InitLogger(*debug)
	proxy.Init()
	socks.NewSocks5(*laddr, *raddr, *secret)
}
