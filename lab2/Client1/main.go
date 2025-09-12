package main

import (
	"Client1/funClient"
	"fmt"
	"net"
)

func main() {
	// 设置随机数种子
	funClient.InitRand()
	funClient.InitSRRand() // 初始化 SR 协议的随机数生成器

	// 人工切换 GBN 和 SR 客户端的布尔变量
	useSR := false // 设置为 true 切换到 SR 客户端，false 使用 GBN 客户端

	// 初始化客户端连接
	conn, _, err := funClient.InitUDPConnection(":8081")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error closing connection: %v\n", err)
		}
	}()

	// 初始化服务器地址
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	if err != nil {
		fmt.Printf("Error resolving server address: %v\n", err)
		return
	}

	// 根据useSR选择GBN或SR协议客户端
	if useSR {
		fmt.Println("Running SR Client")
		funClient.RunSRClient(conn, serverAddr)
	} else {
		fmt.Println("Running GBN Client")
		funClient.RunServer(conn, serverAddr) // GBN 客户端
	}
}
