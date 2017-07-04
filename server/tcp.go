package main

import (
	"../common"
	"fmt"
	"net"
	"time"
	"github.com/jbenet/go-reuseport"
)

const (
	tcpRecvBufSize int = 1024
)

func tcp(listenString string, proxyString string) int {
	mac, ret := common.GetMac()
	if ret < 0 {
		fmt.Println("get local mac fail !")
		return -1
	}
	fmt.Println("real get mac =", mac)

	tcp := &common.TCP{
		RecvBufSize:tcpRecvBufSize,
	}

	ret = tcp.CreateReuseServer(listenString, tcpParseProtoData)
	if ret < 0 {
		fmt.Println("tcp.CreateReuseServer fail ! ret =", ret)
		return -2
	}

	proxyConn, ret := tcp.CreateReuseClient(proxyString, listenString, tcpParseProtoData)
	if ret < 0 {
		fmt.Println("tcp.CreateReuseClient fail ! ret =", ret)
		return -3
	}

	tcpServerTryToLoginAssist(proxyConn, mac)

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
	case common.AssistSendNATClientToServer:
		tcpAssistSendNATClientToServerRep(conn, realData)
	case common.ClientSendToServer:
		tcpClientSendToServerRep(conn, realData) //maybe something wrong !
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

/*************************** passive *********************************/
func tcpServerLoginAssistRep(conn net.Conn, realData []byte) {
	var result int8
	result = int8(realData[0])

	if result != 0 {
		fmt.Println("server login fail !")
		return
	}

	fmt.Println("server login success !")
}

func tcpAssistSendNATClientToServerRep(conn net.Conn, realData []byte) {
	clientNAT := string(realData)
	fmt.Println("clientNAT = ", clientNAT)
	tcpOpenHole(clientNAT)
}

func tcpClientSendToServerRep(conn net.Conn, realData []byte) {
	fmt.Println(conn.RemoteAddr().String(), "recv client send data :", realData)
	tcpServerSendtoClientData(conn, realData)
}
/********************************************************************/

/************************ proactive **********************************/
func tcpServerTryToLoginAssist(conn net.Conn, mac net.HardwareAddr) {
	dataLen := len(mac)
	data := []byte(mac)
	fmt.Println("serverTryToLoginTransit: data =", data, "dataLen =", dataLen)
	tcpCreatePackageSend(conn, common.ServerLoginAssist, data)
}

func tcpServerSendtoClientData(conn net.Conn, data []byte) {
	tcpCreatePackageSend(conn, common.ServerSendToClient, data)
}
/**********************************************************************/

func tcpOpenHole(stringNAT string) {
	addrNAT, err := net.ResolveUDPAddr("udp", stringNAT)
	if err != nil {
		fmt.Println("ResolveUDPAddr udp fail !", stringNAT)
		return
	}

	tcpConn, err1 := reuseport.Dial("tcp", localTcpBusinessAddr, stringNAT)
	if err1 != nil {
		fmt.Println("the same as we think, tcpConn NATA is fail !")
	} else {
		fmt.Println("unbelievable ! tcpConn NATA success !")
	}

	addr ,err := net.ResolveUDPAddr("udp", localTcpBusinessAddr)
	if err!= nil {
		fmt.Println("resolve udp addr fail !")
		return
	}

	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("listen udp fail !", localTcpBusinessAddr)
		return
	}

	var writeData [2]byte
	writeData[0] = common.ServerSendOpenHoleDataToClient
	writeData[1] = 46
	for i := 0; i < 3; i++ {
		writeBytes, err := udpConn.WriteToUDP(writeData[0:], addrNAT)
		fmt.Println(i, "serverSendOpenHoleDataToClient:writeBytes =", writeBytes)
		if err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(2 * time.Second)
	}

	defer func (tcpConn net.Conn, udpConn *net.UDPConn){
		if tcpConn != nil {
			tcpConn.Close()
		}
		if udpConn != nil{
			udpConn.Close()
		}
		fmt.Println("run defer close tcpConn and udpConn")
	}(tcpConn, udpConn)
}