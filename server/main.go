package main

import (
	"flag"

	"proxy"
)

func main() {

	addr := flag.String("addr", "", "listen addr")
	secret := flag.String("secret", "", "secret")
	bufSize := flag.Int("buf", 4, "buf size, Unit is KB")
	flag.Parse()
	p := proxy.NewHttpProxy(*addr, *secret)
	p.SetBufSize(*bufSize * 1024)
	p.Listen()

}
