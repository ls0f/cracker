package main

import (
	"flag"
	"fmt"
	"os"

	g "github.com/golang/glog"
	"github.com/ls0f/cracker"
)

var (
	GitTag    = "2000.01.01.release"
	BuildTime = "2000-01-01T00:00:00+0800"
)

func main() {

	addr := flag.String("addr", "", "listen addr")
	secret := flag.String("secret", "", "secret")
	version := flag.Bool("version", false, "version")
	https := flag.Bool("https", false, "https")
	cert := flag.String("cert", "", "cert file")
	key := flag.String("key", "", "private key file")
	flag.Parse()
	if *version {
		fmt.Printf("GitTag: %s \n", GitTag)
		fmt.Printf("BuildTime: %s \n", BuildTime)
		os.Exit(0)
	}
	p := cracker.NewHttpProxy(*addr, *secret, *https)
	if *https {
		f, err := os.Stat(*cert)
		if err != nil {
			g.Fatal(err)
		}
		if f.IsDir() {
			g.Fatal("cert should be file")
		}
		f, err = os.Stat(*key)
		if err != nil {
			g.Fatal(err)
		}
		if f.IsDir() {
			g.Fatal("key should be file")
		}
		p.ListenHTTPS(*cert, *key)
	} else {
		p.Listen()
	}

}
