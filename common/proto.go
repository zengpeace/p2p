package common

import (
	"net"
	"fmt"
	"time"
)

type ProtoHeadType struct {
	Tag uint16
	Version uint8
	Cmd uint8
	BodySize uint16
}

const (
	ProxyListenIp string = "0.0.0.0"

	ProtoTag uint16 = 65535
	ProtoVersion uint8 = 1

	ServerLoginAssist uint8 = 101
	ClientLoginAssist uint8 = 102
	AssistSendNATClientToServer uint8 = 103
	ServerSendOpenHoleDataToClient uint8 = 104

	ClientSendToServer uint8 = 201
	ServerSendToClient uint8 = 202
)

var (
	protoHeadTmp ProtoHeadType
	ProtoHeadSize int = SizeStruct(protoHeadTmp)
)

func UdpOpenHole(l net.PacketConn, stringNAT string) (net.Addr, int) {
	var a net.Addr
	a, ret := GetAddr(stringNAT)
	if ret < 0 {
		fmt.Println("GetAddr fail !", stringNAT, ret)
		return a, -1
	}

	var writeData [8]byte
	writeData[0] = byte((ProtoTag >> 8) & 0x00FF)
	writeData[1] = byte(ProtoTag & 0x00FF)
	writeData[2] = ProtoVersion
	writeData[3] = ServerSendOpenHoleDataToClient
	writeData[4] = 2
	writeData[5] = 0
	writeData[6] = 41
	writeData[7] = 42

	for i := 0; i < 3; i++ {
		writeBytes, err := l.WriteTo(writeData[0:], a)
		fmt.Println(i, "serverSendOpenHoleDataToClient:writeBytes =", writeBytes, l.LocalAddr().String(), a.String())
		if err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	return a, 0
}

func CreateSendPackage(cmd uint8, realData *[]byte) *[]byte{
	bodySize := uint16(len(*realData))
	packageSize := ProtoHeadSize + int(bodySize)

	head := &ProtoHeadType{ProtoTag, ProtoVersion, cmd, bodySize}
	sendPackage := make([]byte, packageSize)
	ChangeHeadStructToSlice(head, sendPackage[0:ProtoHeadSize])
	copy(sendPackage[ProtoHeadSize:], *realData)

	return &sendPackage
}

func ChangeHeadStructToSlice(head *ProtoHeadType, slice []byte) {
	slice[0] = byte(head.Tag & 0x00FF)
	slice[1] = byte((head.Tag >> 8) & 0x00FF)
	slice[2] = head.Version
	slice[3] = head.Cmd
	slice[4] = byte(head.BodySize & 0x00FF)
	slice[5] = byte((head.BodySize >> 8) & 0x00FF)
}
