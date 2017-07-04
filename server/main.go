package main

import (
	"time"
)

const (
	localTcpBusinessAddr string = "0.0.0.0:5102"
	proxyTcpListenAddr string = "39.108.5.110:1937"


	localUdpBusinessAddr string = "0.0.0.0:5103"

	//proxyUdpListenAddr string = "39.108.5.110:1938"
	proxyUdpListenAddr string = "incastyun.cn:1937"
	//proxyUdpListenAddr string = "39.108.5.110:1937"
)

func main()  {
	//tcp(localTcpBusinessAddr, proxyTcpListenAddr)
	udp(localUdpBusinessAddr, proxyUdpListenAddr)

	for {
		time.Sleep(10 * time.Second)
		continue
	}
}

