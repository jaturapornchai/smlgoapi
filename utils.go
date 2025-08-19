package main

import (
	"net"
	"strings"
)

const (
	fallbackHost = "localhost"
	dnsResolver  = "8.8.8.8:80"
	localHost    = "0.0.0.0:"
)

// getLocalIP returns the local IP address of the machine
func getLocalIP() string {
	conn, err := net.Dial("udp", dnsResolver)
	if err != nil {
		// Fallback to localhost if can't determine IP
		return fallbackHost
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getDisplayURL returns a user-friendly URL for display
func getDisplayURL(serverAddr string) string {
	localIP := getLocalIP()

	// Replace 0.0.0.0 with actual IP for display purposes
	if strings.HasPrefix(serverAddr, localHost) {
		port := strings.TrimPrefix(serverAddr, localHost)
		return localIP + ":" + port
	}

	return serverAddr
}
