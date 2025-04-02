package main

import (
	"bytes"
	"fmt"
	"io/fs"
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

func handleRequest(conn net.Conn) {
	defer conn.Close()

	req := make([]byte, 1024)
	conn.Read(req)

	url := string(req)
	parts := filter(strings.Split(url, "\r\n"), func(val string) bool {
		return len(strings.TrimSpace(val)) > 0
	})
	urlParts := strings.Split(parts[0], " ")
	pathSegments := filter(strings.Split(urlParts[1], "/"), func(val string) bool {
		return len(strings.TrimSpace(val)) > 0
	})

	encodingHeaders := filter(parts, func(val string) bool {
		return strings.HasPrefix(val, "Accept-Encoding:")
	})

	requestEncoding := ""
	if len(encodingHeaders) > 0 {
		requestEncoding = strings.TrimSpace(strings.ReplaceAll(encodingHeaders[0], "Accept-Encoding:", ""))
	}

	shouldReturn := handleGetFfiles(pathSegments, urlParts, conn, requestEncoding)
	if shouldReturn {
		return
	}

	shouldReturn = handlePostFiles(pathSegments, urlParts, parts, conn, requestEncoding)
	if shouldReturn {
		return
	}

	shouldReturn = handleGetUserAgent(pathSegments, urlParts, parts, conn, requestEncoding)
	if shouldReturn {
		return
	}

	shouldReturn = handleGetEcho(pathSegments, conn, requestEncoding)
	if shouldReturn {
		return
	}

	if !strings.HasPrefix(string(req), "GET / HTTP/1.1") {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}

	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}

func handleGetEcho(pathSegments []string, conn net.Conn, requestEncoding string) bool {
	if len(pathSegments) > 0 && pathSegments[0] == "echo" {
		useEncoding, val, schema := encodeValue(pathSegments[1], requestEncoding)
		length := len(val)

		if useEncoding {
			conn.Write([]byte(
				fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
					schema,
					length,
					val),
			))
		} else {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", length, val)))
		}

		return true
	}

	return false
}

func handlePostFiles(pathSegments []string, urlParts []string, parts []string, conn net.Conn, requestEncoding string) bool {
	if len(pathSegments) > 0 && pathSegments[0] == "files" && urlParts[0] == "POST" {
		requestBodyIndex := len(parts) - 1
		requestBody := strings.TrimSuffix(parts[requestBodyIndex], "\\0")

		filename := pathSegments[1]
		directory := os.Args[2]
		fullPath := directory + filename

		err := os.WriteFile(fullPath, bytes.Trim([]byte(requestBody), "\x00"), fs.FileMode(os.O_CREATE))
		if err != nil {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			return true
		}

		conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		return true
	}
	return false
}

func handleGetFfiles(pathSegments []string, urlParts []string, conn net.Conn, requestEncoding string) bool {
	if len(pathSegments) > 0 && pathSegments[0] == "files" && urlParts[0] == "GET" {
		filename := pathSegments[1]
		directory := os.Args[2]
		fullPath := directory + filename

		file, err := os.ReadFile(fullPath)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return true
		}

		useEncoding, fileContent, schema := encodeValue(string(file), requestEncoding)
		fileLength := len(fileContent)
		if useEncoding {
			conn.Write([]byte(
				fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: %s\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
					schema,
					fileLength,
					fileContent),
			))
		} else {
			conn.Write([]byte(
				fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
					fileLength,
					fileContent),
			))
		}

		return true
	}
	return false
}

func handleGetUserAgent(pathSegments []string, urlParts []string, parts []string, conn net.Conn, requestEncoding string) bool {
	if len(pathSegments) > 0 && pathSegments[0] == "user-agent" {
		userAgentHeader := filter(parts, func(val string) bool {
			return strings.HasPrefix(val, "User-Agent:")
		})[0]
		userAgentHeader = strings.TrimSpace(strings.ReplaceAll(userAgentHeader, "User-Agent:", ""))
		useEncoding, userAgentHeaderEnc, schema := encodeValue(userAgentHeader, requestEncoding)
		userAgentLen := len(userAgentHeaderEnc)

		if useEncoding {
			conn.Write([]byte(
				fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
					schema,
					userAgentLen,
					userAgentHeaderEnc),
			))
		} else {
			conn.Write([]byte(
				fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
					userAgentLen,
					userAgentHeaderEnc),
			))
		}

		return true
	}

	return false
}
