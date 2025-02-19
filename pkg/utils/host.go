package utils

import (
	"net"

	"github.com/pkg/errors"
)

func GetLocalIP() (string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", errors.WithStack(err)
	}

	for _, iface := range netInterfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not IPv4
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("no local ip found")
}
