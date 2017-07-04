package main

import (
	"../common"
	"fmt"
	"net"
)

const (
	udpRecvBufSize int = 1024
)

var (
	listenConn net.PacketConn
	proxyAddr net.Addr
	clientAddr net.Addr
)

func udp(listenString string, proxyString string) int {
	mac, ret := common.GetMac()
	if ret < 0 {
		fmt.Println("get local mac fail !")
		return -1
	}
	fmt.Println("real get mac =", mac)

	p, err := net.ResolveUDPAddr("udp", proxyString)
	if err != nil {
		fmt.Println(err.Error())
		return -2
	}

	proxyAddr = net.Addr(p)

	udp := &common.UDP{
		RecvBufSize:udpRecvBufSize,
	}

	listenConn, ret = udp.CreateReuseServer(listenString, udpParseProtoData)
	if ret < 0 {
		fmt.Println("udp.CreateReuseServer fail ! ret =", ret)
		return -2
	}

	fmt.Println("prepare to login assist !")
	udpServerTryToLoginAssist(listenConn, mac)
	return 0
}

func udpParseProtoData(conn net.PacketConn, raddr *net.Addr, recvData *[]byte) {
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
		udpServerLoginAssistRep(conn, raddr, realData)
	case common.AssistSendNATClientToServer:
		udpAssistSendNATClientToServerRep(conn, raddr, realData)
	case common.ClientSendToServer:
		udpClientSendToServerRep(conn, raddr, realData)
	default:
		fmt.Println("unknow udp package cmd ! ", head.Cmd)
	}
}


/*************************** passive *********************************/
func udpServerLoginAssistRep(conn net.PacketConn, raddr *net.Addr, realData []byte) {
	var result int8
	result = int8(realData[0])

	if result != 0 {
		fmt.Println("server login fail !")
		return
	}

	fmt.Println("server login success !")
}

func udpAssistSendNATClientToServerRep(conn net.PacketConn, raddr *net.Addr, realData []byte) {
	clientNAT := string(realData)
	fmt.Println("clientNAT = ", clientNAT)
	clientAddr, _ = common.UdpOpenHole(listenConn, clientNAT)
	//go udpServerSend(conn)
}

func udpClientSendToServerRep(conn net.PacketConn, raddr *net.Addr, realData []byte) {
	fmt.Println((*raddr).String(), "recv client send data :", realData)
	udpServerSendtoClientData(conn, realData)
}
/********************************************************************/

/************************ proactive **********************************/
func udpServerTryToLoginAssist(conn net.PacketConn, mac net.HardwareAddr) {
	dataLen := len(mac)
	data := []byte(mac)
	fmt.Println("serverTryToLoginAssist: data =", data, "dataLen =", dataLen)
	udpCreatePackageSend(conn, proxyAddr, common.ServerLoginAssist, data)
}

func udpServerSendtoClientData(conn net.PacketConn, data []byte) {
	fmt.Println("server return back data to client !")
	udpCreatePackageSend(conn, clientAddr, common.ServerSendToClient, data)
}

func udpServerProactiveSendtoClientData(conn net.PacketConn) {
	fmt.Println("server send data to client !")
	data := make([]byte, 3)
	data[0] = 1
	data[1] = 2
	data[2] = 3
	udpCreatePackageSend(conn, clientAddr, common.ServerSendToClient, data)
}
/**********************************************************************/

func udpCreatePackageSend(conn net.PacketConn, raddr net.Addr, cmd uint8, realData []byte) {
	sendPackage := common.CreateSendPackage(cmd, &realData)
	packageSize := len(*sendPackage)

	sendSize, err := conn.WriteTo(*sendPackage, raddr)
	if err != nil || sendSize != packageSize{
		fmt.Println("send package error !", cmd, packageSize, sendPackage)
		if err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println("send to", raddr, *sendPackage)
}
