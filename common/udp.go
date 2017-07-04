package common

import (
	"net"
	"fmt"
	"github.com/jbenet/go-reuseport"
	"time"
)

type UDP struct {
	RecvBufSize int
}

func (u *UDP) CreateServer(listenString string, parse func(*net.UDPConn, *net.UDPAddr, *[]byte)) (*net.UDPConn, int) {
	listenAddr, err := net.ResolveUDPAddr("udp", listenString)
	if err != nil {
		fmt.Println("net.ResolveUDPAddr fail !", listenString)
		return nil, -1
	}

	listenConn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		fmt.Println("net.ListenUDP fail !", listenAddr)
		return nil, -2
	}

	fmt.Println("listening udp at", listenString)
	go u.serverRecvThread(listenConn, parse)

	return listenConn, 0
}

func (u *UDP) serverRecvThread(listenConn *net.UDPConn, parse func(*net.UDPConn, *net.UDPAddr, *[]byte)) {
	recvBuf := make([]byte, u.RecvBufSize)
	for  {
		readBytes, raddr, err := listenConn.ReadFromUDP(recvBuf)
		if err != nil || raddr == nil || readBytes <= 0 {
			fmt.Println("listenConn.ReadFromUDP fail ! readBytes =", readBytes)
			if err != nil {
				fmt.Println(err.Error())
			}
			if raddr != nil {
				fmt.Println("raddr =", raddr)
			}
			time.Sleep(1 * time.Second)
		}

		d := recvBuf[0:readBytes]
		parse(listenConn, raddr, &d)
	}

	defer listenConn.Close()
}

func (u *UDP) CreateReuseServer(listenString string, parse func(net.PacketConn, *net.Addr, *[]byte)) (net.PacketConn, int) {
	var p net.PacketConn
	listenConn, err := reuseport.ListenPacket("udp", listenString)
	if err != nil {
		fmt.Println("reuseport.ListenPacket fail !", listenString)
		return p, -1
	}

	fmt.Println("reuse listening udp at", listenString)
	go u.serverReuseRecvThread(listenConn, parse)

	return listenConn, 0
}

func (u *UDP) serverReuseRecvThread(listenConn net.PacketConn, parse func(net.PacketConn, *net.Addr, *[]byte)) {
	recvBuf := make([]byte, u.RecvBufSize)
	for {
		readBytes, raddr, err := listenConn.ReadFrom(recvBuf)
		if err != nil || raddr == nil || readBytes <= 0 {
			fmt.Println("listenConn.ReadFromUDP fail ! readBytes =", readBytes)
			if err != nil {
				fmt.Println(err.Error())
			}
			if raddr != nil {
				fmt.Println("raddr =", raddr)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		d := recvBuf[0:readBytes]
		parse(listenConn, &raddr, &d)
	}

	defer listenConn.Close()
}


func (u *UDP) CreateClient(serverString string, parse func(net.Conn, *[]byte)) (net.Conn, int) {
	return u.client(serverString, "", parse, false)
}

func (u *UDP) CreateReuseClient(serverString string, clientString string, parse func(net.Conn, *[]byte)) (net.Conn, int) {
	return u.client(serverString, clientString, parse, true)
}

func (u *UDP) client(serverString string, clientString string, parse func(net.Conn, *[]byte), isReuse bool) (net.Conn, int) {
	var conn net.Conn
	var err error
	if isReuse == false {
		conn, err = net.Dial("udp", serverString)
		if err != nil {
			fmt.Println("net.Dial udp fail !", serverString)
			return conn, -1
		}
		fmt.Println("connect to udp server success !", serverString)
	} else {
		conn, err = reuseport.Dial("udp", clientString, serverString)
		if err != nil {
			fmt.Println("reuseport.Dial udp fail !", serverString, clientString)
			return conn, -2
		}
		fmt.Println("connect to udp server success !", serverString, clientString)
	}

	go u.clientRecvThread(conn, parse)

	return conn, 0
}

func (u *UDP) clientRecvThread(conn net.Conn, parse func(net.Conn, *[]byte)) {
	recvBuf := make([]byte, u.RecvBufSize)
	for {
		readBytes, err := conn.Read(recvBuf)
		if err != nil || readBytes <= 0 {
			fmt.Println("listenConn.ReadFromUDP fail ! readBytes =", readBytes)
			if err != nil {
				fmt.Println(err.Error())
			}
			time.Sleep(1 * time.Second)
			continue
		}

		d := recvBuf[0:readBytes]
		parse(conn, &d)
	}

	defer conn.Close()
}

