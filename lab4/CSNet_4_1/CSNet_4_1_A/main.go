package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	HostBIP = "192.168.245.134"
	Port    = "12345"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", HostBIP+":"+Port)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing UDP connection:", err)
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter message to send: ")
		text, _ := reader.ReadString('\n')
		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Error sending message:", err)
		} else {
			fmt.Println("Message sent at:", time.Now().Format(time.RFC3339))
		}
	}
}
