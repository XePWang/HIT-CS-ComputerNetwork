package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	cacheTTL    = 60 * time.Second // 缓存生存时间
	fishingSrc  = "www.hit.edu.cn"
	fishingDest = "http://today.hit.edu.cn"
)

// 禁止访问的网站
var invalidWebsites = []string{
	"http://www.hit.edu.cn",
}

// 限制访问的用户
var restrictHosts = []string{
	"127.0.0.1",
}

// 控制禁止访问开关
var isAccessForbiddenHostEnabled = false // 将此设置为 false 以允许用户访问
var isAccessForbiddenSiteEnabled = false // 将此设置为 false 以允许网站访问

var cache = make(map[string]*cachedResponse)

type cachedResponse struct {
	response  *http.Response
	body      []byte
	timestamp time.Time
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 检查用户过滤（如果开关启用）
	if isAccessForbiddenHostEnabled && isRestrictedHost(r.RemoteAddr) {
		fmt.Println("Access forbidden", r.RemoteAddr)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// 检查网站过滤（如果开关启用）
	if isAccessForbiddenSiteEnabled && isInvalidWebsite(r.URL.String()) {
		fmt.Println("Access denied", r.URL.String())
		http.Error(w, "Access to this website is forbidden", http.StatusForbidden)
		return
	}

	// fmt.Println(r.URL.String())

	// 钓鱼网站引导
	if isFishingSite(r) {
		fmt.Println("Redirecting to", fishingDest)
		redirectToFishingSite(w, r)
		return
	}

	//// 处理HTTPS请求
	//if r.Method == http.MethodConnect {
	//	handleConnect(w, r)
	//	return
	//}

	// 检查缓存
	cachedResp, found := cache[r.URL.String()]
	if found && time.Since(cachedResp.timestamp) < cacheTTL {
		// 如果缓存有效，检查并添加 If-Modified-Since 头部
		lastModified := cachedResp.response.Header.Get("Last-Modified")
		if lastModified != "" {
			r.Header.Set("If-Modified-Since", lastModified)
			fmt.Printf("URL: %s\n", r.URL.String())
			fmt.Printf("缓存的 Last-Modified: %s\n", lastModified)
			fmt.Printf("添加的 If-Modified-Since 头部: %s\n", r.Header.Get("If-Modified-Since"))
		}
	}

	// 转发请求到原服务器
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, "Error connecting to the upstream server", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close() // 关闭响应体

	// 处理响应状态
	if resp.StatusCode == http.StatusNotModified {
		// 如果响应为 304 Not Modified，返回缓存的响应
		if found {
			fmt.Println("HTTP:304")
			writeResponse(w, cachedResp)
			return
		}
	} else if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response body", http.StatusInternalServerError)
			return
		}

		// 打印调试信息
		fmt.Printf("HTTP:%d\n", resp.StatusCode)

		// 缓存响应，包括响应头和体
		cache[r.URL.String()] = &cachedResponse{
			response:  resp,
			body:      body, // 保存读取的响应体
			timestamp: time.Now(),
		}

		// 将响应写回客户端
		writeResponse(w, &cachedResponse{
			response: resp,
			body:     body,
		})
		return
	}

	// 如果不是 200, 206 或 304，直接写回原始响应
	writeResponse(w, &cachedResponse{
		response: resp,
		body:     nil, // 不缓存其他状态码的响应体
	})

}

//func handleConnect(w http.ResponseWriter, r *http.Request) {
//	// 响应状态码 200
//	w.WriteHeader(http.StatusOK)
//
//	// 创建 TCP 连接
//	conn, err := net.Dial("tcp", r.URL.Host)
//	if err != nil {
//		http.Error(w, "Error connecting to target server", http.StatusBadGateway)
//		return
//	}
//	// 在 defer 中处理 conn.Close() 的错误
//	defer func() {
//		if err := conn.Close(); err != nil {
//			fmt.Println("Error closing the connection:", err)
//		}
//	}()
//
//	// 响应头部完成后开始读取和写入数据
//	_, err = io.WriteString(w, "HTTP/1.1 200 Connection Established\r\n\r\n")
//	if err != nil {
//		http.Error(w, "Error writing response", http.StatusInternalServerError)
//		return
//	}
//
//	// 从请求体读取数据并发送到目标连接
//	go func() {
//		defer func() {
//			if err := conn.Close(); err != nil {
//				fmt.Println("Error closing the connection in goroutine:", err)
//			}
//		}()
//		_, err := io.Copy(conn, r.Body)
//		if err != nil {
//			fmt.Println("Error copying data to target server:", err)
//		}
//	}()
//
//	// 从目标连接读取数据并发送到响应
//	_, err = io.Copy(w, conn)
//	if err != nil {
//		fmt.Println("Error copying data from target server:", err)
//	}
//}

func isInvalidWebsite(requestURL string) bool {
	for _, invalidWebsite := range invalidWebsites {
		if strings.Contains(requestURL, invalidWebsite) {
			return true
		}
	}
	return false
}

func isRestrictedHost(remoteAddr string) bool {
	for _, host := range restrictHosts {
		if strings.HasPrefix(remoteAddr, host) {
			return true
		}
	}
	return false
}

// 检查是否是钓鱼网站
func isFishingSite(r *http.Request) bool {
	return r.URL.Host == fishingSrc && r.URL.Path == "/"
}

// 重定向到钓鱼网站
func redirectToFishingSite(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fishingDest, http.StatusFound)
}

func writeResponse(w http.ResponseWriter, cachedResp *cachedResponse) {
	// 复制缓存响应的头部到响应
	for key, values := range cachedResp.response.Header {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}

	// 设置状态码
	w.WriteHeader(cachedResp.response.StatusCode)

	// 如果状态码为 200 或 206，写入缓存的响应体
	if cachedResp.response.StatusCode == http.StatusOK || cachedResp.response.StatusCode == http.StatusPartialContent {
		// 调试输出缓存的状态码、头部信息和响应体大小
		fmt.Printf("写入缓存响应，状态码: %d\n", cachedResp.response.StatusCode)
		fmt.Printf("缓存响应头部: %v\n", cachedResp.response.Header)
		fmt.Printf("缓存响应体大小: %d bytes\n", len(cachedResp.body))

		_, err := w.Write(cachedResp.body)
		if err != nil {
			http.Error(w, "Error writing response body", http.StatusInternalServerError)
		}
	} else {
		// 对于非 200 或 206 状态码，不写入响应体
		fmt.Printf("状态码为 %d，不写入响应体\n", cachedResp.response.StatusCode)
	}
}
