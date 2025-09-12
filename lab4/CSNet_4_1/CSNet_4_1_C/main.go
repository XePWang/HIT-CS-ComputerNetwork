package main

import (
	"fmt"
	"net"
	"time"
)

const (
	HostCIP = "192.168.245.135"
	Port    = "12345"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", HostCIP+":"+Port)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error listening on UDP port:", err)
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}()

	buffer := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP socket:", err)
			continue
		}

		message := string(buffer[:n])
		fmt.Printf("Message received at %s: %s\n", time.Now().Format(time.RFC3339), message)
	}
}
