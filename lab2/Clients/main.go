package main

import (
	"Clients/doClient" // 请将 'your_module' 替换为实际的模块名
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// 封装关闭连接的错误处理
func closeConnection(conn *net.UDPConn) {
	err := conn.Close()
	if err != nil {
		fmt.Println("Error closing connection:", err)
	}
}

//// 封装用户输入处理
//func handleUserInputError() string {
//	var input string
//	_, err := fmt.Scanln(&input)
//	if err != nil {
//		fmt.Println("Error reading input:", err)
//		return ""
//	}
//	return input
//}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <server_ip:port>")
		return
	}

	serverAddr := os.Args[1]
	udpAddr, err := net.ResolveUDPAddr("udp4", serverAddr)
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		return
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer closeConnection(conn) // 使用封装的关闭连接处理函数

	fmt.Println("Connected to server:", serverAddr)
	handleUserInput(conn, udpAddr)
}

func handleUserInput(conn *net.UDPConn, addr *net.UDPAddr) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nEnter a command (GET <filename>, PUSH <filename>, LIST):")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// 去除输入字符串末尾的换行符和空格
		input = strings.TrimSpace(input)

		// 使用空格分割输入，但最多分割成两部分
		parts := strings.SplitN(input, " ", 2)

		command := parts[0]

		switch command {
		case "GET":
			if len(parts) > 1 {
				filename := parts[1]
				doClient.SendGetRequest(conn, addr, filename)
			} else {
				fmt.Println("Usage: GET <filename>")
			}
		case "PUSH":
			if len(parts) > 1 {
				filename := parts[1]
				doClient.SendPushRequest(conn, addr, filename)
			} else {
				fmt.Println("Usage: PUSH <filename>")
			}
		case "LIST":
			doClient.SendListRequest(conn, addr)
		default:
			fmt.Println("Invalid command. Use GET, PUSH, or LIST.")
		}
	}
}
