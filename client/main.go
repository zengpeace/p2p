package main

import (
	"time"
)

type ProtoHeadType struct {
	tag uint16
	version uint8
	cmd uint8
	bodySize uint16
}

const (
	localTcpBusinessAddr string = "0.0.0.0:6102"
	proxyTcpListenAddr string = "39.108.5.110:1937"


	localUdpBusinessAddr string = "0.0.0.0:6103"

	//proxyUdpListenAddr string = "39.108.5.110:1938"
	proxyUdpListenAddr string = "incastyun.cn:1937"
	//proxyUdpListenAddr string = "39.108.5.110:1937"
)

var (
	//serverMac = []byte{0x00, 0x0c, 0x29, 0x41, 0xfa, 0x88}
	//serverMac = []byte{0x00, 0x50, 0x56, 0xA6, 0x4C, 0xC5}
	//serverMac = []byte{0x00, 0x0c, 0x29, 0x64, 0x57, 0x5a}
	serverMac = []byte{0x00, 0x50, 0x56, 0x89, 0x6D, 0xBD}
	//serverMac = []byte{0x00, 0x16, 0x3e, 0x08, 0xf5, 0x36}
)

func main() {
	//tcp(localTcpBusinessAddr, proxyTcpListenAddr)
	udp(localUdpBusinessAddr, proxyUdpListenAddr)

	for {
		time.Sleep(10 * time.Second)
		continue
	}
}

