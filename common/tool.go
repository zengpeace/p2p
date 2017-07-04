package common

import (
	"fmt"
	"strings"
	"os/exec"
	"reflect"
	"net"
	"encoding/hex"
)

func GetAddr(s string) (net.Addr, int) {
	var a net.Addr
	addr, err := net.ResolveUDPAddr("udp", s)
	if err != nil {
		fmt.Println("ResolveUDPAddr udp fail !", s)
		return a, -1
	}
	a = net.Addr(addr)
	return a, 0
}

func GetMac() (net.HardwareAddr, int) {
	var mac net.HardwareAddr
	getted := 0
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("get net interfaces fail !")
		return mac, -1
	}

	for _, inter := range interfaces {
		if getted == 0 && inter.Name != "lo" {
			getted = 1
			mac = inter.HardwareAddr
		}
	}

	//fmt.Println("getMac:", *mac)
	return mac, 0
}

func ChangePrivateIpToPubilc(privateAddr string) string {
	needToChange := false

	all := strings.Split(privateAddr, ":")
	privateIp := all[0]
	privatePort := all[1]

	if strings.Compare(privateIp, "192.168.100.1") == 0 {
		needToChange = true
	}

	if needToChange {
		cmd := exec.Command("/bin/sh", "-c", "curl members.3322.org/dyndns/getip")
		out, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error(), "return org privateAddr", privateAddr)
			return privateAddr
		}

		o := out[0:len(out)-1]
		return string(o) + ":" + privatePort
	} else {
		return privateAddr
	}
}

func ChangeMacToString(mac net.HardwareAddr) string {
	return hex.EncodeToString([]byte(mac))
}

func SizeStruct(data interface{}) int {
	return sizeof(reflect.ValueOf(data))
}

func sizeof(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Map:
		sum := 0
		keys := v.MapKeys()
		for i := 0; i < len(keys); i++ {
			mapkey := keys[i]
			s := sizeof(mapkey)
			if s < 0 {
				return -1
			}
			sum += s
			s = sizeof(v.MapIndex(mapkey))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum
	case reflect.Slice, reflect.Array:
		sum := 0
		for i, n := 0, v.Len(); i < n; i++ {
			s := sizeof(v.Index(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.String:
		sum := 0
		for i, n := 0, v.Len(); i < n; i++ {
			s := sizeof(v.Index(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return 0
		}
		return sizeof(v.Elem())
	case reflect.Struct:
		sum := 0
		for i, n := 0, v.NumField(); i < n; i++ {
			if v.Type().Field(i).Tag.Get("ss") == "-" {
				continue
			}
			s := sizeof(v.Field(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Int:
		return int(v.Type().Size())

	default:
		fmt.Println("t.Kind() no found:", v.Kind())
	}

	return -1
}

/*func changeMacToString(mac net.HardwareAddr) string {
	org := []byte(mac)
	s := ""
	for i := 0; i < macBytes; i++ {
		if (org[i] & 0xF0) != 0x00 {
			s += fmt.Sprintf("%x", org[i])
		} else {
			s += fmt.Sprintf("0%x", org[i])
		}
	}
	fmt.Println("changeMacToString: mac =", mac, "s =", s)

	return s
}*/