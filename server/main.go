package main

import (
	"flag"

	"proxy"
)

func main() {

	addr := flag.String("addr", "", "listen addr")
	secret := flag.String("secret", "", "secret")
	flag.Parse()
	p := proxy.NewHttpProxy(*addr, *secret)
	p.Listen()

}
