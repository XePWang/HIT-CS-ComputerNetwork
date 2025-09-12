package funClient

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	packetLossProb = 0.2 // 丢包概率
	bufferSize     = 1024
)

// RandomGenerator 是一个自定义的随机数生成器
var RandomGenerator *rand.Rand

// InitRand 初始化一个随机数生成器
func InitRand() {
	// 创建一个新的随机数生成器实例
	RandomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// SimulatePacketLoss 模拟丢包
func SimulatePacketLoss() bool {
	// 使用我们初始化的随机数生成器而不是全局的 rand
	return RandomGenerator.Float32() < packetLossProb
}

// InitUDPConnection 初始化 UDP 连接
func InitUDPConnection(port string) (*net.UDPConn, *net.UDPAddr, error) {
	clientAddr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve UDP address: %v", err)
	}
	conn, err := net.ListenUDP("udp", clientAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on UDP: %v", err)
	}
	return conn, clientAddr, nil
}

// ReceivePacket 接收数据包
func ReceivePacket(conn *net.UDPConn) (int, string, error) {
	buffer := make([]byte, bufferSize)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return -1, "", fmt.Errorf("failed to read from UDP: %v", err)
	}
	var seqNum int
	var data string

	if _, err := fmt.Sscanf(string(buffer[:n]), "%d:%s", &seqNum, &data); err != nil {
		fmt.Printf("failed to parse packet: %v\n", err)
		return -1, "", fmt.Errorf("failed to parse packet: %v", err)
	}
	return seqNum, data, nil
}

// SendAck 发送ACK
func SendAck(conn *net.UDPConn, serverAddr *net.UDPAddr, seqNum int) error {
	ackMsg := fmt.Sprintf("ACK:%d", seqNum)
	_, err := conn.WriteToUDP([]byte(ackMsg), serverAddr)
	if err != nil {
		return fmt.Errorf("failed to send ACK: %v", err)
	}
	fmt.Printf("Sending ACK for packet %d\n", seqNum)
	return nil
}

// RunServer 主循环，处理数据包接收和ACK发送
func RunServer(conn *net.UDPConn, serverAddr *net.UDPAddr) {
	expectedSeqNum := 0
	for {
		seqNum, data, err := ReceivePacket(conn)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// 模拟丢包
		if SimulatePacketLoss() {
			fmt.Printf("Packet loss: SeqNum=%d\n", seqNum)
			continue
		}

		fmt.Printf("Received packet: SeqNum=%d, Data=%s\n", seqNum, data)

		// 正常处理数据包并发送ACK
		if seqNum == expectedSeqNum {
			if err := SendAck(conn, serverAddr, seqNum); err != nil {
				fmt.Println(err)
			}
			expectedSeqNum++
		} else {
			// 乱序数据包，重发最后一个ACK
			if err := SendAck(conn, serverAddr, expectedSeqNum-1); err != nil {
				fmt.Println(err)
			}
		}
	}
}
