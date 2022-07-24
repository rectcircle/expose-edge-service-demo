package helper

import (
	"errors"
	"net"
	"os"
)

// GetIP 获取本机 IP，仅示例，请勿用于生产。
func GetIP() (string, error) {
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		return podIP, nil
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.IsPrivate() || ip.IsGlobalUnicast() {
				return ip.String(), nil
			}
			// 应该有更多逻辑
		}
	}
	return "", errors.New("no ip found")
}
