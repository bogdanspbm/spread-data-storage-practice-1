package ports

import (
	"fmt"
	"net"
)

func IsPortOpen(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false // Порт уже используется или недоступен
	}
	defer listener.Close()
	return true // Порт свободен
}

func FindAvailablePort(startPort int) int {
	for port := startPort; port <= 65535; port++ {
		if IsPortOpen(port) {
			return port
		}
	}
	return 0 // Ни один порт не доступен
}
