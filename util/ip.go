package util

import (
	"github.com/gin-gonic/gin"
	"net"
	"strings"
)

func GetClientIP(c *gin.Context) string {
	const localIP string = "127.0.0.1"
	clientIP := c.ClientIP()
	RemoteIP := c.RemoteIP()

	ip := c.Request.Header.Get("X-Forwarded-For")

	if strings.Contains(ip, localIP) || ip == "" {
		ip = c.Request.Header.Get("X-real-ip")
	}
	if ip == "" {
		ip = localIP
	}

	if RemoteIP != localIP {
		ip = RemoteIP
	}
	if clientIP != localIP {
		ip = clientIP
	}
	return ip
}

// 获取本地ip地址
func GetLocalIP() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		ipAddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipAddr.IP.IsLoopback() {
			continue
		}
		if !ipAddr.IP.IsGlobalUnicast() {
			continue
		}
		return ipAddr.IP.String(), nil
	}
	return
}
