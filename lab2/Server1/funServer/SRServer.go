package funServer

import (
	"fmt"
	"net"
	"time"
)

const (
	SRWindowSize   = 4
	SRTotalPackets = 10
	SRTimeout      = 2 * time.Second
)

type SRPacket struct {
	SeqNum int
	Data   string
	Acked  bool
}

func SendSRPacket(conn *net.UDPConn, addr *net.UDPAddr, packet SRPacket) {
	fmt.Printf("Sending packet: SeqNum=%d, Data=%s\n", packet.SeqNum, packet.Data)
	msg := fmt.Sprintf("%d:%s", packet.SeqNum, packet.Data)
	if _, err := conn.WriteToUDP([]byte(msg), addr); err != nil {
		fmt.Printf("Error sending packet: SeqNum=%d, Error=%v\n", packet.SeqNum, err)
	}
}

func ReceiveSRAck(conn *net.UDPConn) int {
	buffer := make([]byte, 1024)
	if err := conn.SetReadDeadline(time.Now().Add(SRTimeout)); err != nil {
		fmt.Printf("Error setting read deadline: %v\n", err)
		return -1
	}
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Printf("Error receiving ack: %v\n", err)
		return -1
	}
	var ack int
	if _, err := fmt.Sscanf(string(buffer[:n]), "SR_ACK:%d", &ack); err != nil {
		fmt.Printf("Error parsing ack: %v\n", err)
		return -1
	}
	return ack
}

func StartSRServer() {
	fmt.Printf("Starting SR Server\n")

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
	packets := make([]SRPacket, SRTotalPackets)
	acked := make([]bool, SRTotalPackets)

	for i := 0; i < SRTotalPackets; i++ {
		packets[i] = SRPacket{SeqNum: i, Data: fmt.Sprintf("Packet-%d", i), Acked: false}
	}

	timers := make([]time.Time, SRTotalPackets)

	for base < SRTotalPackets {
		for nextSeqNum < base+SRWindowSize && nextSeqNum < SRTotalPackets {
			if !acked[nextSeqNum] {
				SendSRPacket(conn, clientAddr, packets[nextSeqNum])
				timers[nextSeqNum] = time.Now()
			}
			nextSeqNum++
		}

		ack := ReceiveSRAck(conn)
		if ack != -1 && ack >= base && ack < base+SRWindowSize {
			fmt.Printf("Received ACK: %d\n", ack)
			packets[ack].Acked = true
			acked[ack] = true

			for base < SRTotalPackets && acked[base] {
				base++
			}
		}

		for i := base; i < nextSeqNum; i++ {
			if !acked[i] && !timers[i].IsZero() && time.Since(timers[i]) > SRTimeout {
				fmt.Printf("Timeout for packet %d, resending...\n", i)
				SendSRPacket(conn, clientAddr, packets[i])
				timers[i] = time.Now()
			}
		}
	}
	fmt.Println("All packets sent and acknowledged!")
}
