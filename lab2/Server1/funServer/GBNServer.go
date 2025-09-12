package funServer

import (
	"fmt"
	"net"
	"time"
)

const (
	WindowSize   = 4
	TotalPackets = 10
	Timeout      = 2 * time.Second
)

type Packet struct {
	SeqNum int
	Data   string
}

func SendPacket(conn *net.UDPConn, addr *net.UDPAddr, packet Packet) {
	fmt.Printf("Sending packet: SeqNum=%d, Data=%s\n", packet.SeqNum, packet.Data)
	msg := fmt.Sprintf("%d:%s", packet.SeqNum, packet.Data)
	if _, err := conn.WriteToUDP([]byte(msg), addr); err != nil {
		fmt.Printf("Error sending packet: SeqNum=%d, Error=%v\n", packet.SeqNum, err)
	}
}

func ReceiveAck(conn *net.UDPConn) int {
	buffer := make([]byte, 1024)
	if err := conn.SetReadDeadline(time.Now().Add(Timeout)); err != nil {
		fmt.Printf("Error receiving ack: Error=%v\n", err)
		return -1
	}
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Printf("Error receiving ack: Error=%v\n", err)
		return -1
	}
	var ack int
	if _, err := fmt.Sscanf(string(buffer[:n]), "ACK:%d", &ack); err != nil {
		fmt.Printf("Error receiving ack: Error=%v\n", err)
		return -1
	}
	return ack
}

func StartGBNServer() {
	fmt.Println("Starting GBN Server")

	serverAddr, _ := net.ResolveUDPAddr("udp", ":8080")
	conn, _ := net.ListenUDP("udp", serverAddr)
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error closing connection: %v\n", err)
		}
	}()

	clientAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8081")

	base := 0
	nextSeqNum := 0
	packets := make([]Packet, TotalPackets)

	for i := 0; i < TotalPackets; i++ {
		packets[i] = Packet{SeqNum: i, Data: fmt.Sprintf("Packet-%d", i)}
	}

	timer := time.Time{}
	for base < TotalPackets {
		for nextSeqNum < base+WindowSize && nextSeqNum < TotalPackets {
			SendPacket(conn, clientAddr, packets[nextSeqNum])
			nextSeqNum++
		}

		if timer.IsZero() {
			timer = time.Now()
		}

		ack := ReceiveAck(conn)
		if ack != -1 {
			fmt.Printf("Received ACK: %d\n", ack)
			base = ack + 1
			timer = time.Time{}
		}

		if !timer.IsZero() && time.Since(timer) > Timeout {
			fmt.Println("Timeout, resending window...")
			timer = time.Now()
			for i := base; i < nextSeqNum; i++ {
				SendPacket(conn, clientAddr, packets[i])
			}
		}
	}
	fmt.Println("All packets sent and acknowledged!")
}
