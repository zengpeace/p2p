package main

import (
	"fmt"
	"../common"
	"net"
	"time"
)

const (
	tcpRecvBufSize int = 4096
)

func tcp(clientString string, proxyString string) int {
	tcp := &common.TCP{
		RecvBufSize:tcpRecvBufSize,
	}

	proxyConn, ret := tcp.CreateReuseClient(proxyString, clientString, tcpParseProtoData)
	if ret < 0 {
		fmt.Println("tcp.CreateReuseClient fail !", proxyString, clientString)
		return -1
	}

	tcpClientTryToLoginAssist(proxyConn)
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
		tcpClientTryToLoginTransitRep(conn, realData)
	case common.ClientLoginAssist:
		tcpClientTryToLoginTransitRep(conn, realData)
	case common.ServerSendOpenHoleDataToClient:
		tcpServerSendOpenHoleDataToClientRep(conn, realData)
	case common.ServerSendToClient:
		tcpServerSendtoClientRep(conn, realData)
	default:
		fmt.Println("unknow tcp package cmd ! ", head.Cmd)
	}
}


/*************************** passive *********************************/
func tcpClientTryToLoginTransitRep(conn net.Conn, realData []byte) {
	var result int8
	result = int8(realData[0])

	if result != 0 {
		fmt.Println("client login fail !")
		return
	}

	serverNAT := string(realData[1:])
	fmt.Println("client login success ! serverNAT = ", serverNAT)

	//tcpOpenHole(serverNAT)

	time.Sleep(2 * time.Second)

	tcp := &common.TCP{
		RecvBufSize:tcpRecvBufSize,
	}

	connectServerSuccess := false
	var serverConn net.Conn
	var ret int
	for i := 0; i < 3; i ++ {
		serverConn, ret = tcp.CreateReuseClient(serverNAT, localTcpBusinessAddr, tcpParseProtoData)
		if ret != 0 {
			time.Sleep(2 * time.Second)
			continue
		} else {
			connectServerSuccess = true
			break
		}
	}

	if connectServerSuccess == false {
		fmt.Println("client connect backend server fail !")
	} else {
		fmt.Println("connectServerSuccess ! serverConn local addr =", serverConn.LocalAddr().String())
		tcpClientSendDataToServer(serverConn)
		tcpClientSendDataToServer(serverConn)
	}
}

func tcpServerSendOpenHoleDataToClientRep(conn net.Conn, realData []byte) {
	fmt.Println("serverSendOpenHoleDataToClientRep: len of realData is", len(realData))
}

func tcpServerSendtoClientRep(conn net.Conn, realData []byte) {
	fmt.Println(conn.RemoteAddr().String(), "recv server send data :", realData)
}
/********************************************************************/

/************************ proactive **********************************/
func tcpClientTryToLoginAssist(conn net.Conn) {
	data := []byte(serverMac)
	tcpCreatePackageSend(conn, common.ClientLoginAssist, data)
}

func tcpClientSendDataToServer(conn net.Conn) {
	fmt.Println("client send data to server !")
	data := make([]byte, 3)
	data[0] = 1
	data[1] = 2
	data[2] = 3
	tcpCreatePackageSend(conn, common.ClientSendToServer, data)
}
/**********************************************************************/

func tcpOpenHole(stringNAT string) {
	addrNAT, err := net.ResolveUDPAddr("udp", stringNAT)
	if err != nil {
		fmt.Println("ResolveUDPAddr udp fail !", stringNAT)
		return
	}

	/*tcpConn, err1 := reuseport.Dial("tcp", localTcpClientAddr, stringNAT)
	if err1 != nil {
		fmt.Println("the same as we think, tcpConn NATA is fail !")
	} else {
		fmt.Println("unbelievable ! tcpConn NATA success !")
	}*/

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

	defer func (udpConn *net.UDPConn){
		if udpConn != nil{
			udpConn.Close()
		}
		fmt.Println("run defer close udpConn")
	}(udpConn)
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
