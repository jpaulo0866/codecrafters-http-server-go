package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	req := make([]byte, 1024)
	conn.Read(req)

	url := string(req)
	parts := strings.Split(url, "\r\n")
	urlParts := strings.Split(parts[0], " ")
	pathSegments := filter(strings.Split(urlParts[1], "/"), func(val string) bool {
		return len(strings.TrimSpace(val)) > 0
	})

	if len(pathSegments) > 0 && pathSegments[0] == "user-agent" {
		userAgentHeader := filter(parts, func(val string) bool {
			return strings.HasPrefix(val, "User-Agent:")
		})[0]
		userAgentHeader = strings.TrimSpace(strings.ReplaceAll(userAgentHeader, "User-Agent:", ""))
		userAgentLen := len(userAgentHeader)
		conn.Write([]byte(
			fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
				userAgentLen,
				userAgentHeader),
		))
		conn.Close()
		return
	}

	if len(pathSegments) > 0 && pathSegments[0] == "echo" {
		val := pathSegments[1]
		length := len(val)
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", length, val)))
		conn.Close()
		return
	}

	if !strings.HasPrefix(string(req), "GET / HTTP/1.1") {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		conn.Close()
		return
	}

	defer conn.Close()
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}
