package main

import (
	"github.com/iooojik-dev/proxy/proxy"
)

func main() {
	runProxyServer()
}

func runProxyServer() {
	_, err := proxy.RunHttpsProxy()
	if err != nil {
		panic(err)
	}
}
