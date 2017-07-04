package main

import (
	"fmt"
	"net"
	"../common"
)

const (
	recvTcpBufSize int = 1024
	proxyListenIp string = "0.0.0.0"
)

var (
	tcpConnMap = make(map[string]*net.Conn)
)

func tcp(proxyListenPort int) int {
	if proxyListenPort <= 0 {
		fmt.Println("proxyListenPort error ! proxyListenPort =", proxyListenPort)
		return -1
	}

	tcp := common.TCP{
		RecvBufSize:recvTcpBufSize,
	}

	listenString := fmt.Sprintf("%s:%d", proxyListenIp, proxyListenPort)
	ret := tcp.CreateServer(listenString, tcpParseProtoData)
	if ret < 0 {
		fmt.Println("tcp.CreateServer fail ! ret =", ret, listenString)
		return -2
	}

	return 0
}

func tcpParseProtoData(conn net.Conn, recvData *[]byte) {
	packageSize := len(*recvData)
	b := (*recvData)[0:common.ProtoHeadSize]
	head := &common.ProtoHeadType{uint16(b[0]) | uint16(b[1])<<8, b[2], b[3], uint16(b[4]) | uint16(b[5])<<8}
	if head.Tag != common.ProtoTag {
		fmt.Println("head.tag error ! now is", head.Tag, "but we want", common.ProtoTag)
		return 
	}

	if head.BodySize + uint16(common.ProtoHeadSize) != uint16(packageSize) {
		fmt.Println("bodySize is not match to packageSize ! bodySize = ", head.BodySize, "packageSize =", packageSize)
		return
	}

	realData := (*recvData)[common.ProtoHeadSize:packageSize]
	switch head.Cmd {
	case common.ServerLoginAssist:
		tcpServerLoginAssistRep(conn, realData)
	case common.ClientLoginAssist:
		tcpClientLoginAssistRep(conn, realData)
	default:
		fmt.Println("unknow tcp package cmd ! ", head.Cmd)
	}
}

func tcpCreatePackageSend(conn net.Conn, cmd uint8, realData []byte) {
	sendPackage := common.CreateSendPackage(cmd, &realData)
	packageSize := len(*sendPackage)

	sendSize, err := conn.Write(*sendPackage)
	if err != nil || sendSize != packageSize{
		fmt.Println("send package error !", cmd, packageSize, sendPackage)
		if err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println("send to", conn.RemoteAddr().String(), *sendPackage)
}

/*************  passive *************/
func tcpServerLoginAssistRep(conn net.Conn, realData []byte) {
	var state int8
	mac := net.HardwareAddr(realData)
	fmt.Println("serverLoginTransitRep: realData =", realData, "mac =", mac)
	ret := tcpSaveConnToMap(conn, mac)
	if ret < 0 {
		state = -1
	} else {
		state = 0
	}

	var data [1]byte
	data[0] = byte(state)
	tcpCreatePackageSend(conn, common.ServerLoginAssist, data[0:])
}

func tcpClientLoginAssistRep(conn net.Conn, realData []byte) {
	var state int8
	if len(realData) != 6 {
		fmt.Println("clientLoginTransitRep: realData not right ! it must be 6 bytes ! realData =", realData)
		return
	}
	mac := net.HardwareAddr(realData)
	fmt.Println("clientLoginTransitRep: realData =", realData, "mac =", mac)
	serverConn, ok := tcpGetConnInMap(mac)
	if ok {
		state = 0
	} else {
		state = -1
	}

	if state == 0 {
		serverAddr := common.ChangePrivateIpToPubilc(serverConn.RemoteAddr().String())
		dataLen := 1 + len(serverAddr)
		data := make([]byte, dataLen)
		data[0] = byte(state)
		copy(data[1:], []byte(serverAddr))
		tcpCreatePackageSend(conn, common.ClientLoginAssist, data)
		tcpTransitSendNATClientDataToServer(conn, serverConn)
	} else {
		dataLen := 1
		data := make([]byte, dataLen)
		data[0] = byte(state)
		tcpCreatePackageSend(conn, common.ClientLoginAssist, data)
	}
}
/********************************************************/

/******************** proactive *************************/
func tcpTransitSendNATClientDataToServer(clientConn net.Conn, serverConn net.Conn) {
	clientAddr := common.ChangePrivateIpToPubilc(clientConn.RemoteAddr().String())
	dataLen := len(clientAddr)
	data := make([]byte, dataLen)
	copy(data[0:], []byte(clientAddr))
	tcpCreatePackageSend(serverConn, common.AssistSendNATClientToServer, data)
}
/*********************************************************/


func tcpSaveConnToMap(conn net.Conn, mac net.HardwareAddr) int {
	m := common.ChangeMacToString(mac)
	_, ok := tcpConnMap[m]
	if ok {
		fmt.Println("this mac has been marked !", mac, conn.RemoteAddr().String())
		return -1
	}

	tcpConnMap[m] = &conn
	return 0
}

func tcpGetConnInMap(mac net.HardwareAddr) (net.Conn, bool) {
	m := common.ChangeMacToString(mac)
	conn, ok := tcpConnMap[m]
	if ok {
		return *conn, ok
	} else {
		return nil, ok
	}
}
