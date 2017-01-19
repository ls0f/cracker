package main

import (
	"github.com/lovedboy/cracker/proxy"
	"flag"
)

func main() {

	addr := flag.String("addr", "", "listen addr")
	secret := flag.String("secret", "", "secret")
	flag.Parse()
	p := proxy.NewHttpProxy(*addr, *secret)
	p.Listen()

}
