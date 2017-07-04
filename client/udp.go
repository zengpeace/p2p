package main

import (
	"fmt"
	"time"
	"net"
	"../common"
)

const (
	udpRecvBufSize int = 4096
)

var (
	listenConn net.PacketConn
	proxyAddr net.Addr
	serverAddr net.Addr
)

func udp(clientString string, proxyString string) int {
	p, err := net.ResolveUDPAddr("udp", proxyString)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}
	proxyAddr = net.Addr(p)

	udp := &common.UDP{
		RecvBufSize:udpRecvBufSize,
	}

	listenConn, ret := udp.CreateReuseServer(clientString, udpParseProtoData)
	if ret < 0 {
		fmt.Println("udp.CreateReuseServer fail !", clientString)
		return -2
	}

	udpClientTryToLoginAssist(listenConn)
	return 0
}

func udpParseProtoData(conn net.PacketConn, raddr *net.Addr, recvData *[]byte) {
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
	case common.ClientLoginAssist:
		udpClientTryToLoginTransitRep(conn, raddr, realData)
	case common.ServerSendOpenHoleDataToClient:
		udpServerSendOpenHoleDataToClientRep(conn, raddr, realData)
	case common.ServerSendToClient:
		udpServerSendtoClientRep(conn, raddr, realData)
	default:
		fmt.Println("unknow tcp package cmd ! ", head.Cmd)
	}
}


/*************************** passive *********************************/
func udpClientTryToLoginTransitRep(conn net.PacketConn, raddr *net.Addr, realData []byte) {
	var result int8
	result = int8(realData[0])

	if result != 0 {
		fmt.Println("client login fail !")
		return
	}

	serverNAT := string(realData[1:])
	fmt.Println("client login success ! serverNAT = ", serverNAT, *raddr)

	time.Sleep(8 * time.Second)

	var ret int
	/*serverAddr, ret = common.UdpOpenHole(conn, serverNAT)
	if ret < 0 {
		fmt.Println("UdpOpenHole fail !", serverNAT, ret)
		return
	}*/

	serverAddr, ret = common.GetAddr(serverNAT)
	if ret < 0 {
		fmt.Println("Get Addr fail !", serverNAT, ret)
		return
	}

	go udpClientSendTestDataToSever(conn)
}

func udpServerSendOpenHoleDataToClientRep(conn net.PacketConn, raddr *net.Addr, realData []byte) {
	fmt.Println("serverSendOpenHoleDataToClientRep: len of realData is", len(realData), *raddr)
}

func udpServerSendtoClientRep(conn net.PacketConn, raddr *net.Addr, realData []byte) {
	fmt.Println("recv server send data :", realData, *raddr)
}
/********************************************************************/

/************************ proactive **********************************/
func udpClientTryToLoginAssist(conn net.PacketConn) {
	data := []byte(serverMac)
	udpCreatePackageSend(conn, &proxyAddr, common.ClientLoginAssist, data)
}

func udpClientSendDataToServer(conn net.PacketConn) {
	//fmt.Println("client send data to server !")
	data := make([]byte, 3)
	data[0] = 1
	data[1] = 2
	data[2] = 3
	udpCreatePackageSend(conn, &serverAddr, common.ClientSendToServer, data)
}
/**********************************************************************/

func udpCreatePackageSend(conn net.PacketConn, raddr *net.Addr, cmd uint8, realData []byte) {
	sendPackage := common.CreateSendPackage(cmd, &realData)
	packageSize := len(*sendPackage)

	sendSize, err := conn.WriteTo(*sendPackage, *raddr)
	if err != nil || sendSize != packageSize{
		fmt.Println("send package error !", cmd, packageSize, sendPackage)
		if err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println("send to", *raddr, *sendPackage)
}

func udpClientSendTestDataToSever(conn net.PacketConn) {
	time.Sleep(6 * time.Second)

	for i := 0; i < 3; i ++ {
		udpClientSendDataToServer(conn)
		time.Sleep(1 * time.Second)
	}

}
