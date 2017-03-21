package main

import (
	"flag"

	"proxy"
)

func main() {

	addr := flag.String("addr", "", "listen addr")
	secret := flag.String("secret", "", "secret")
	bufSize := flag.Int("buf", 1, "buf size, Unit is MB")
	flag.Parse()
	p := proxy.NewHttpProxy(*addr, *secret)
	p.SetBufSize(*bufSize * 1024 * 1024)
	p.Listen()

}
