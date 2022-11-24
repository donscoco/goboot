package util

import (
	"net"
	"strings"
)

const (
	Subnetwork_192_168 = "192.168."
	Subnetwork_10      = "10."
	Subnetwork_lo      // 本地回环子网 127.0.0.1
)

func GetIpv4_192_168() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ip, _, err := net.ParseCIDR(addr.String()); err == nil {
			sip := ip.To4().String()
			if strings.Index(addr.String(), Subnetwork_192_168) == 0 {
				return sip
			}
		}
	}
	return ""
}
