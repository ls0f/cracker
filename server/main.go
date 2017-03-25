package main

import (
	"flag"
	"fmt"
	"logger"
	"os"
	"proxy"
)

var (
	GitTag    = "2000.01.01.release"
	BuildTime = "2000-01-01T00:00:00+0800"
)

var g = logger.GetLogger()

func main() {

	addr := flag.String("addr", "", "listen addr")
	secret := flag.String("secret", "", "secret")
	debug := flag.Bool("debug", false, "debug mode")
	version := flag.Bool("v", false, "version")
	https := flag.Bool("https", false, "https")
	flag.Parse()
	logger.InitLogger(*debug)
	if *version {
		fmt.Printf("GitTag: %s \n", GitTag)
		fmt.Printf("BuildTime: %s \n", BuildTime)
		os.Exit(0)
	}
	if *https {
		f, err := os.Stat("cert.pem")
		if err != nil {
			g.Fatal(err)
		}
		if f.IsDir() {
			g.Fatal("cert.pem should be file")
		}
		f, err = os.Stat("key.pem")
		if err != nil {
			g.Fatal(err)
		}
		if f.IsDir() {
			g.Fatal("key.pem should be file")
		}
	}
	p := proxy.NewHttpProxy(*addr, *secret, *https)
	p.Listen()

}
