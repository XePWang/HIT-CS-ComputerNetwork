package doClient

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const (
	BUFFERSIZE = 1024
)

// 封装写入文件的错误处理
func writeFile(file *os.File, data []byte) {
	_, err := file.Write(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

// SendGetRequest 发送 GET 请求
func SendGetRequest(conn *net.UDPConn, addr *net.UDPAddr, filename string) {
	request := fmt.Sprintf("GET %s", filename)
	_, err := conn.Write([]byte(request))
	if err != nil {
		fmt.Println("Error sending GET request:", err, "\n", addr)
		return
	}

	// 创建本地文件准备接收数据
	filePath := filepath.Join("./", "download_"+filename)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
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
			fmt.Println("Error receiving data:", err, "\n", addr)
			break
		}

		writeFile(file, buffer[:n])

		if n < BUFFERSIZE {
			break
		}
	}

	fmt.Println("File downloaded successfully:", filePath)
}

// SendPushRequest 发送 PUSH 请求
func SendPushRequest(conn *net.UDPConn, addr *net.UDPAddr, filename string) {
	filePath := filepath.Join("./", filename)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file:", err)
		}
	}()

	request := fmt.Sprintf("PUSH %s", filename)
	_, err = conn.Write([]byte(request))
	if err != nil {
		fmt.Println("Error sending PUSH request:", err, "\n", addr)
		return
	}

	buffer := make([]byte, BUFFERSIZE)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			fmt.Println("Error reading file:", err)
			break
		}

		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Println("Error sending file data:", err, "\n", addr)
			break
		}

		if n < BUFFERSIZE {
			break
		}
	}

	// 等待确认消息
	buffer = make([]byte, BUFFERSIZE)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error receiving confirmation:", err)
		return
	}
	fmt.Println("Server response:", string(buffer[:n]))
}

// SendListRequest 发送 LIST 请求
func SendListRequest(conn *net.UDPConn, addr *net.UDPAddr) {
	request := "LIST"
	_, err := conn.Write([]byte(request))
	if err != nil {
		fmt.Println("Error sending LIST request:", err, "\n", addr)
		return
	}

	buffer := make([]byte, BUFFERSIZE)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error receiving file list:", err, "\n", addr)
		return
	}

	fmt.Println("Files on server:")
	fmt.Println(string(buffer[:n]))
}
