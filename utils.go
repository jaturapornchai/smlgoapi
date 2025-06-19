package main

import (
	"net"
	"strings"
)

// getLocalIP returns the local IP address of the machine
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback to localhost if can't determine IP
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getDisplayURL returns a user-friendly URL for display
func getDisplayURL(serverAddr string) string {
	localIP := getLocalIP()

	// Replace 0.0.0.0 with actual IP for display purposes
	if strings.HasPrefix(serverAddr, "0.0.0.0:") {
		port := strings.TrimPrefix(serverAddr, "0.0.0.0:")
		return localIP + ":" + port
	}

	return serverAddr
}
