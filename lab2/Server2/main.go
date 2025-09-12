package main

import (
	"Server2/doServer" // 引入处理请求的包
	"fmt"
	"net"
	"os"
)

func main() {
	addr := net.UDPAddr{
		Port: 8080,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Error starting UDP server:", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("Error closing UDP connection:", err)
		}
	}()

	fmt.Println("Server is listening on port 8080...")

	// 调用 doServer 来处理请求
	doServer.HandleRequest(conn)
}
