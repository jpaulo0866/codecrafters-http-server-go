package main

import (
	"fmt"
	"log"
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

	log.Printf("%s", os.Args)

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}

}

func validArguments() bool {
	return len(os.Args) > 3 && os.Args[1] == "--directory"
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	req := make([]byte, 1024)
	conn.Read(req)

	url := string(req)
	parts := strings.Split(url, "\r\n")
	urlParts := strings.Split(parts[0], " ")
	pathSegments := filter(strings.Split(urlParts[1], "/"), func(val string) bool {
		return len(strings.TrimSpace(val)) > 0
	})

	if len(pathSegments) > 0 && pathSegments[0] == "files" {

		filename := pathSegments[1]
		directory := os.Args[2]
		fullPath := directory + filename

		file, err := os.ReadFile(fullPath)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}

		fileContent := string(file)
		fileLength := len(fileContent)
		conn.Write([]byte(
			fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
				fileLength,
				fileContent),
		))
		return
	}

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
		return
	}

	if len(pathSegments) > 0 && pathSegments[0] == "echo" {
		val := pathSegments[1]
		length := len(val)
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", length, val)))
		return
	}

	if !strings.HasPrefix(string(req), "GET / HTTP/1.1") {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}

	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}
