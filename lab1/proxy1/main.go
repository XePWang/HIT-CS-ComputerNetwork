package main

import (
	"fmt"
	"net/http"
)

func main() {
	go http.HandleFunc("/", handleRequest) // 使用 http.HandleFunc 注册请求处理
	fmt.Println("Proxy server is listening on :8080")
	err := http.ListenAndServe(":8080", nil) // 启动 HTTP 服务器
	if err != nil {
		fmt.Println("Error starting the proxy server:", err)
		return
	}
}
