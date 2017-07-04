package main

import (
	"time"
)

const (
	//tcpProxyListenPort int = 1937
	udpProxyListenPort int = 1937
)

func main() {
	//tcp(tcpProxyListenPort)
	udp(udpProxyListenPort)

	for {
		time.Sleep(10 * time.Second)
		continue
	}
}
