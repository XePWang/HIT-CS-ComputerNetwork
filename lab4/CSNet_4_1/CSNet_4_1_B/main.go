package main

import (
	"fmt"
	"net"
	"time"
)

const (
	HostBIP = "192.168.245.134"
	HostCIP = "192.168.245.135"
	Port    = "12345"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", HostBIP+":"+Port)
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

	udpAddrForward, err := net.ResolveUDPAddr("udp", HostCIP+":"+Port)
	if err != nil {
		fmt.Println("Error resolving forward UDP address:", err)
		return
	}

	buffer := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP socket:", err)
			continue
		}

		message := string(buffer[:n])
		fmt.Printf("Message received at %s: %s\n", time.Now().Format(time.RFC3339), message)

		_, err = conn.WriteToUDP(buffer[:n], udpAddrForward)
		if err != nil {
			fmt.Println("Error forwarding message:", err)
		} else {
			fmt.Println("Message forwarded at:", time.Now().Format(time.RFC3339))
		}
	}
}
