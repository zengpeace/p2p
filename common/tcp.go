package common

import (
	"net"
	"fmt"
	"time"
	"github.com/jbenet/go-reuseport"
)

type TCP struct {
	RecvBufSize int
}

func (t *TCP) CreateServer(listenString string, parse func(net.Conn, *[]byte)) int {
	return t.server(listenString, parse, false)
}

func (t *TCP) CreateReuseServer(listenString string, parse func(net.Conn, *[]byte)) int {
	return t.server(listenString, parse, true)
}

func (t *TCP) server(listenString string, parse func(net.Conn, *[]byte), isReuse bool) int {
	listenFunc := net.Listen
	if isReuse {
		listenFunc = reuseport.Listen
	}

	listener, err := listenFunc("tcp", listenString)
	if err != nil {
		fmt.Println("net.Listen", listenString, "fail !")
		return -1
	}

	fmt.Println("assistServer listening tcp at", listenString)
	go t.listenThread(listener, parse)

	return 0
}

func (t *TCP) listenThread(listener net.Listener, parse func(net.Conn, *[]byte)) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept fail !")
			time.Sleep(2 * time.Second)
			continue
		}

		go t.recvThread(conn, parse)
	}
}

func (t *TCP) CreateClient(serverString string, parse func(net.Conn, *[]byte)) (net.Conn, int) {
	var c net.Conn
	conn, err := net.Dial("tcp", serverString)
	if err != nil {
		fmt.Println("net.Dial", serverString, "fail !")
		return c, -1
	}

	go t.recvThread(conn, parse)
	return conn, 0
}

func (t *TCP) CreateReuseClient(serverString string, clientString string, parse func(net.Conn, *[]byte)) (net.Conn, int) {
	var c net.Conn
	conn, err := reuseport.Dial("tcp", clientString, serverString)
	if err != nil {
		fmt.Println("reuseport.Dial fail !", clientString, serverString)
		return c, -1
	}

	go t.recvThread(conn, parse)
	return conn, 0
}

func (t *TCP) recvThread(conn net.Conn, parse func(net.Conn, *[]byte)) {
	dataBuf := make([]byte, t.RecvBufSize)
	for {
		packageSize, err := conn.Read(dataBuf[0:])
		if err != nil {
			fmt.Println("disconnect from", conn.RemoteAddr().String())
			return
		}

		if packageSize <= 0 || packageSize > t.RecvBufSize {
			fmt.Println("packageSize error !", packageSize)
			time.Sleep(1 * time.Second)
			continue
		}

		fmt.Println("recv tcp data from", conn.RemoteAddr().String(), dataBuf[0:packageSize])
		d := dataBuf[0:packageSize]
		parse(conn, &d)
	}

	defer conn.Close()
}
