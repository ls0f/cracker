package main

import (
	"flag"
	"herokuproxy/proxy"
)

func main() {

	addr := flag.String("addr", "", "listen addr")
	flag.Parse()
	p := proxy.NewHttpProxy(*addr)
	p.Listen()

}
