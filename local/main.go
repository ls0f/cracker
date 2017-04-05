package main

import (
	"cracker/http_connect"
	"cracker/logger"
	"cracker/proxy"
	"cracker/socks"
	"flag"
	"fmt"
	"os"
	"sync"
)

var (
	GitTag    = "2000.01.01.release"
	BuildTime = "2000-01-01T00:00:00+0800"
)

var g = logger.GetLogger()

func main() {
	socks5Addr := flag.String("socks5", "127.0.0.1:1080", "socks5 listen addr")
	httpAddr := flag.String("http", "127.0.0.1:8080", "http listen addr")
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
	wg := &sync.WaitGroup{}
	if *socks5Addr != "" {
		s, err := socks.NewSocks5(*socks5Addr, *raddr, *secret)
		if err != nil {
			g.Fatal(err)
		}
		wg.Add(1)
		go func() {
			s.Wait()
			wg.Done()
		}()
	}
	if *httpAddr != "" {
		h, err := http_connect.NewHttpConnect(*httpAddr, *raddr, *secret)
		if err != nil {
			g.Fatal(err)
		}
		wg.Add(1)
		go func() {
			h.Wait()
			wg.Done()
		}()
	}
	wg.Wait()
}
