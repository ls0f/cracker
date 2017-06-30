package main

import (
	"flag"
	"fmt"
	"os"

	. "github.com/lovedboy/cracker/cracker/local_server"
	"github.com/lovedboy/cracker/cracker/logger"
	"github.com/lovedboy/cracker/cracker/proxy"
)

var (
	GitTag    = "2000.01.01.release"
	BuildTime = "2000-01-01T00:00:00+0800"
)

var g = logger.GetLogger()

func main() {
	addr := flag.String("addr", "127.0.0.1:1080", "listen addr")
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
	s, err := NewLocalProxyServer(*addr, *raddr, *secret)
	if err != nil {
		g.Fatal(err)
	}
	s.Wait()
}
