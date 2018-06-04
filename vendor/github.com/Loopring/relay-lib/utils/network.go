package utils

import (
	"net"
	"strings"
)

func GetLocalIp() string {
	var res = "unknown"
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return res
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				res = ipnet.IP.To4().String()
			}
		}
	}
	return res
}

func GetLocalIpByPrefix(prefix string) string {
	var res = "unknown"
	interfaces, err := net.Interfaces()
	if err != nil {
		return res
	}

	for _, i := range interfaces {
		if addresses, err := i.Addrs(); err == nil {
			for _, v := range addresses {
				parts := strings.Split(v.String(), ":")
				if len(parts) > 1 { //ipv6 address
					continue
				} else {
					parts := strings.Split(v.String(), "/")
					if len(parts) == 2 && strings.HasPrefix(parts[0], prefix) {
						return parts[0]
					} else {
						continue
					}
				}
			}
		}
	}
	return res
}
