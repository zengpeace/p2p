package main

import (
	"fmt"
	"net"
	"../common"
)

const (
	udpRecvBufSize int = 1024
)

var (
	udpConnMap = make(map[string]*net.UDPAddr)
)

func udp(proxyListenPort int) int {
	if proxyListenPort <= 0 {
		fmt.Println("proxyListenPort error ! proxyListenPort =", proxyListenPort)
		return -1
	}

	udp := &common.UDP{
		RecvBufSize:udpRecvBufSize,
	}

	listenString := fmt.Sprintf("%s:%d", proxyListenIp, proxyListenPort)
	_, ret := udp.CreateServer(listenString, udpParseProtoData)
	if ret < 0 {
		fmt.Println("udp.CreateServer error ! ret =", ret)
		return -2
	}

	return 0
}

func udpParseProtoData(conn *net.UDPConn, raddr *net.UDPAddr, recvData *[]byte) {
	b := (*recvData)[0:]
	head := &common.ProtoHeadType{uint16(b[0]) | uint16(b[1])<<8, b[2], b[3], uint16(b[4]) | uint16(b[5])<<8}
	if head.Tag != common.ProtoTag {
		fmt.Println("head.tag error ! now is", head.Tag, "but we want", common.ProtoTag)
		return
	}

	packageSize := len(*recvData)
	if head.BodySize + uint16(common.ProtoHeadSize) != uint16(packageSize) {
		fmt.Println("bodySize is not match to packageSize ! bodySize = ", head.BodySize, "packageSize =", packageSize)
		return
	}

	realData := (*recvData)[common.ProtoHeadSize:packageSize]
	switch head.Cmd {
	case common.ServerLoginAssist:
		udpServerLoginTransitRep(conn, raddr, realData)
	case common.ClientLoginAssist:
		udpClientLoginTransitRep(conn, raddr, realData)
	default:
		fmt.Println("unknow tcp package cmd ! ", head.Cmd)
	}
}

func udpCreatePackageSend(conn *net.UDPConn, raddr *net.UDPAddr, cmd uint8, realData []byte) {
	sendPackage := common.CreateSendPackage(cmd, &realData)
	packageSize := len(*sendPackage)


	sendSize, err := conn.WriteToUDP(*sendPackage, raddr)
	if err != nil || sendSize != packageSize{
		fmt.Println("send package error !", cmd, packageSize, sendPackage)
		if err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println("send to", raddr, sendPackage)
}

/*************  passive *************/
func udpServerLoginTransitRep(conn *net.UDPConn, raddr *net.UDPAddr, realData []byte) {
	var state int8
	mac := net.HardwareAddr(realData)
	fmt.Println("serverLoginTransitRep: realData =", realData, "mac =", mac)
	ret := udpSaveConnToMap(raddr, mac)
	if ret < 0 {
		state = -1
	} else {
		state = 0
	}

	var data [1]byte
	data[0] = byte(state)
	udpCreatePackageSend(conn, raddr, common.ServerLoginAssist, data[0:])
}

func udpClientLoginTransitRep(conn *net.UDPConn, raddr *net.UDPAddr, realData []byte) {
	var state int8
	if len(realData) != 6 {
		fmt.Println("clientLoginTransitRep: realData not right ! it must be 6 bytes ! realData =", realData)
		return
	}
	mac := net.HardwareAddr(realData)
	fmt.Println("clientLoginTransitRep: realData =", realData, "mac =", mac)
	serverAddr, ok := udpGetConnInMap(mac)
	if ok {
		state = 0
	} else {
		state = -1
	}

	if state == 0 {
		//saddr := common.ChangePrivateIpToPubilc(serverAddr.String())
		saddr := serverAddr.String()
		dataLen := 1 + len(saddr)
		data := make([]byte, dataLen)
		data[0] = byte(state)
		copy(data[1:], []byte(saddr))
		udpCreatePackageSend(conn, raddr, common.ClientLoginAssist, data)

		spubAddr, err := net.ResolveUDPAddr("udp", saddr)
		if err != nil {
			fmt.Println("ResolveUDPAddr fail !", saddr)
			return
		}

		udpAssistSendNATClientDataToServer(conn, raddr, spubAddr)
	} else {
		dataLen := 1
		data := make([]byte, dataLen)
		data[0] = byte(state)
		udpCreatePackageSend(conn, raddr, common.ClientLoginAssist, data)
	}
}
/********************************************************/

/******************** proactive *************************/
func udpAssistSendNATClientDataToServer(conn *net.UDPConn, clientAddr *net.UDPAddr, serverAddr *net.UDPAddr) {
	caddr := common.ChangePrivateIpToPubilc(clientAddr.String())
	dataLen := len(caddr)
	data := make([]byte, dataLen)
	copy(data[0:], []byte(caddr))
	udpCreatePackageSend(conn, serverAddr, common.AssistSendNATClientToServer, data)
}
/*********************************************************/


func udpSaveConnToMap(raddr *net.UDPAddr, mac net.HardwareAddr) int {
	m := common.ChangeMacToString(mac)
	_, ok := tcpConnMap[m]
	if ok {
		fmt.Println("this mac has been marked !", mac, raddr.String())
		return -1
	}

	udpConnMap[m] = raddr
	return 0
}

func udpGetConnInMap(mac net.HardwareAddr) (*net.UDPAddr, bool) {
	m := common.ChangeMacToString(mac)
	raddr, ok := udpConnMap[m]
	if ok {
		return raddr, ok
	} else {
		return nil, ok
	}
}