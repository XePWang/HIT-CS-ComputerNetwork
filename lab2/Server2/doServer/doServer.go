package doServer

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	BUFFERSIZE = 1024
)

// 处理GET请求
func handleGet(conn *net.UDPConn, addr *net.UDPAddr, filename string) {
	filePath := filepath.Join("./", filename)
	file, err := os.Open(filePath)
	if err != nil {
		msg := "File not found"
		_, writeErr := conn.WriteToUDP([]byte(msg), addr)
		if writeErr != nil {
			fmt.Println("Error sending file not found message:", writeErr)
		}
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file:", err)
		}
	}()

	buffer := make([]byte, BUFFERSIZE)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading file:", err)
			break
		}

		_, writeErr := conn.WriteToUDP(buffer[:n], addr)
		if writeErr != nil {
			fmt.Println("Error sending file data:", writeErr)
			break
		}
	}
}

// 处理PUSH请求
func handlePush(conn *net.UDPConn, addr *net.UDPAddr, filename string) {
	filePath := filepath.Join("./", filename)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		_, writeErr := conn.WriteToUDP([]byte("Error creating file"), addr)
		if writeErr != nil {
			fmt.Println("Error sending file creation error message:", writeErr)
		}
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file:", err)
		}
	}()

	buffer := make([]byte, BUFFERSIZE)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving file data:", err)
			_, writeErr := conn.WriteToUDP([]byte("Error receiving file data"), addr)
			if writeErr != nil {
				fmt.Println("Error sending receive error message:", writeErr)
			}
			break
		}

		_, writeErr := file.Write(buffer[:n])
		if writeErr != nil {
			fmt.Println("Error writing to file:", writeErr)
			_, sendErr := conn.WriteToUDP([]byte("Error writing to file"), addr)
			if sendErr != nil {
				fmt.Println("Error sending write error message:", sendErr)
			}
			break
		}

		if n < BUFFERSIZE {
			break
		}
	}

	// 上传完成，发送确认消息给客户端
	_, writeErr := conn.WriteToUDP([]byte("File uploaded successfully"), addr)
	if writeErr != nil {
		fmt.Println("Error sending success message:", writeErr)
	}
}

// 处理文件列表请求
func handleListFiles(conn *net.UDPConn, addr *net.UDPAddr) {
	files, err := os.ReadDir("./")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		_, writeErr := conn.WriteToUDP([]byte("Error reading directory"), addr)
		if writeErr != nil {
			fmt.Println("Error sending directory read error message:", writeErr)
		}
		return
	}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, file.Name())
		}
	}

	response := strings.Join(fileList, "\n")
	_, writeErr := conn.WriteToUDP([]byte(response), addr)
	if writeErr != nil {
		fmt.Println("Error sending file list:", writeErr)
	}
}

// HandleRequest 处理客户端请求
func HandleRequest(conn *net.UDPConn) {
	buffer := make([]byte, BUFFERSIZE)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		request := string(buffer[:n])
		command := strings.Split(request, " ")

		switch command[0] {
		case "GET":
			if len(command) > 1 {
				handleGet(conn, addr, command[1])
			}
		case "PUSH":
			if len(command) > 1 {
				handlePush(conn, addr, command[1])
			}
		case "LIST":
			handleListFiles(conn, addr)
		default:
			_, writeErr := conn.WriteToUDP([]byte("Invalid command"), addr)
			if writeErr != nil {
				fmt.Println("Error sending invalid command message:", writeErr)
			}
		}
	}
}
