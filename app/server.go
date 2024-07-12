package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var fileDirectory string

func main() {
	fmt.Println("Logs from your program will appear here!")

	// Parse the command line arguments
	for i, arg := range os.Args {
		if arg == "--directory" {
			fileDirectory = os.Args[i+1]
			break
		}
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()
	
	for { // Infinite loop to keep the server running
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection", err.Error())
			continue
		}
		// Process each incoming connection in a separate goroutine, allowing
		// hanling multiple connections concurrently
		go handleConnection(conn)
	}
}


func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Read the request line
	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	fmt.Println("Request line: ", requestLine)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return
	}
	// Read the headers and content length from the request
	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) != 3 {
		fmt.Println("Invalid request line")
		return
	}
	method, path, _ := parts[0], parts[1], parts[2]
	headers := make(map[string]string)
	var contentLength int
	var acceptEncoding string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading headers: ", err.Error())
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
			if parts[0] == "Content-Length" {
				contentLength, _ = strconv.Atoi(parts[1])
			}
			if parts[0] == "Accept-Encoding" {
				acceptEncoding = parts[1]
			}
		}
	}

	var status string
	var responseBody string
	var contentEncoding string

	switch {
	case path == "/":
		status = "200 OK"
		sendResponse(conn, status, "", contentEncoding)
	case strings.HasPrefix(path, "/echo/"):
		status = "200 OK"
		responseBody = strings.TrimPrefix(path, "/echo/")
		contentEncoding = determineContentEncoding(acceptEncoding)
		sendResponse(conn, status, responseBody, contentEncoding)
	case path == "/user-agent":
		status = "200 OK"
		contentEncoding = determineContentEncoding(acceptEncoding)
		sendResponse(conn, status, headers["User-Agent"], contentEncoding)
	case strings.HasPrefix(path, "/files/"):
		filename := strings.TrimPrefix(path, "/files/")
		if method == "GET" {
			status, filePath := handleFileRequest(filename)
			if status == "200 OK" {
				contentEncoding = determineContentEncoding(acceptEncoding)
				sendFileResponse(conn, status, filePath, contentEncoding)
			} else {
				sendResponse(conn, status, "", "")
			}
		} else if method == "POST" {
			status := handleFileCreation(reader, filename, contentLength)
			sendResponse(conn, status, "", "")
		}
	default:
		status = "404 Not Found"
		sendResponse(conn, status, "", "")
	}

	// Log the request and response status
	logRequest(conn, requestLine, status)
}

func sendResponse(conn net.Conn, status string, body string, contentEncoding string) {
	var responseBody []byte
	var contentLength int

	if contentEncoding == "gzip" {
		compressed, err := compressString(body)
		if err != nil {
			log.Printf("Error compressing response: %v", err)
			body = ""
			contentEncoding = ""
			responseBody = []byte(body)
			contentLength = len(responseBody)
		} else {
			responseBody = compressed
			contentLength = len(responseBody)
		}
	} else {
		responseBody = []byte(body)
		contentLength = len(responseBody)
	}

	response := fmt.Sprintf("HTTP/1.1 %s\r\nContent-Type: text/plain\r\n", status)
	if contentEncoding != "" {
		response += fmt.Sprintf("Content-Encoding: %s\r\n", contentEncoding)
	}
	response += fmt.Sprintf("Content-Length: %d\r\n\r\n", contentLength)

	conn.Write([]byte(response))
	conn.Write(responseBody)
}

func sendFileResponse(conn net.Conn, status string, filePath string, contentEncoding string) {
    fileData, err := os.ReadFile(filePath)
    if err != nil {
        log.Printf("Error reading file: %v", err)
        sendResponse(conn, "500 Internal Server Error", "", "")
        return
    }

    var responseBody []byte
    var contentLength int

    if contentEncoding == "gzip" {
        compressed, err := compressBytes(fileData)
        if err != nil {
            log.Printf("Error compressing file response: %v", err)
            sendResponse(conn, "500 Internal Server Error", "", "")
            return
        }
        responseBody = compressed
    } else {
        responseBody = fileData
    }
    contentLength = len(responseBody)

    response := fmt.Sprintf("HTTP/1.1 %s\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n", status, contentLength)
    if contentEncoding != "" {
        response += fmt.Sprintf("Content-Encoding: %s\r\n", contentEncoding)
    }
    response += "\r\n"

    conn.Write([]byte(response))
    conn.Write(responseBody)
}

func compressString(s string) ([]byte, error) {
	return compressBytes([]byte(s))
}

func compressBytes(data []byte) ([]byte, error) {
    var b bytes.Buffer
    gz, err := gzip.NewWriterLevel(&b, gzip.DefaultCompression)
    if err != nil {
        return nil, err
    }
    if _, err := gz.Write(data); err != nil {
        return nil, err
    }
    if err := gz.Close(); err != nil {
        return nil, err
    }
    return b.Bytes(), nil
}

func determineContentEncoding(acceptEncoding string) string {
	if strings.Contains(acceptEncoding, "gzip") {
		return "gzip"
	}
	return ""
}

func handleFileCreation(reader *bufio.Reader, filename string, contentLength int) string {
    filePath := filepath.Join(fileDirectory, filename)
    file, err := os.Create(filePath)
    if err != nil {
        log.Printf("Error creating file: %v", err)
        return "500 Internal Server Error"
    }
    defer file.Close()

    _, err = io.CopyN(file, reader, int64(contentLength))
    if err != nil {
        log.Printf("Error writing to file: %v", err)
        return "500 Internal Server Error"
    }

    return "201 Created"
}

func handleFileRequest(filename string) (string, string) {
    filePath := filepath.Join(fileDirectory, filename)
    _, err := os.ReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            log.Printf("File not found: %s", filePath)
            return "404 Not Found", ""
        }
        log.Printf("Error reading file: %v", err)
        return "500 Internal Server Error", ""
    }
    return "200 OK", filePath
}

func logRequest(conn net.Conn, requestLine string, status string) {
	remoteAddr := conn.RemoteAddr().String()
	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	method, path := parts[0], parts[1]
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	log.Printf("[%s] %s - %s %s - %s\n", timestamp, remoteAddr, method, path, status)
}
