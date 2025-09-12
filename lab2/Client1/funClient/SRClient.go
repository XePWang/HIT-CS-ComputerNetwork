package funClient

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	srPacketLossProb = 0.2 // SR协议的丢包概率
	srBufferSize     = 1024
	srWindowSize     = 4 // SR 接收窗口大小
	srMaxSeqNum      = 8 // 最大序列号，假设为 0-7 的序列号循环使用
)

// SRRandomGenerator 是一个自定义的随机数生成器（针对SR协议）
var SRRandomGenerator *rand.Rand

// InitSRRand 初始化一个随机数生成器（针对SR协议）
func InitSRRand() {
	// 创建一个新的随机数生成器实例
	SRRandomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// SimulateSRPacketLoss 模拟SR协议中的丢包
func SimulateSRPacketLoss() bool {
	// 使用初始化的随机数生成器
	return SRRandomGenerator.Float32() < srPacketLossProb
}

// InitSRUDPConnection 初始化 SR 协议的 UDP 连接
func InitSRUDPConnection(port string) (*net.UDPConn, *net.UDPAddr, error) {
	clientAddr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve SR UDP address: %v", err)
	}
	conn, err := net.ListenUDP("udp", clientAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on SR UDP: %v", err)
	}
	return conn, clientAddr, nil
}

// ReceiveSRPacket 接收 SR 协议的数据包
func ReceiveSRPacket(conn *net.UDPConn) (int, string, error) {
	buffer := make([]byte, srBufferSize)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return -1, "", fmt.Errorf("failed to read from SR UDP: %v", err)
	}
	var seqNum int
	var data string

	if _, err := fmt.Sscanf(string(buffer[:n]), "%d:%s", &seqNum, &data); err != nil {
		fmt.Printf("failed to parse SR packet: %v\n", err)
		return -1, "", fmt.Errorf("failed to parse SR packet: %v", err)
	}
	return seqNum, data, nil
}

// SendSRAck 发送 SR 协议的 ACK
func SendSRAck(conn *net.UDPConn, serverAddr *net.UDPAddr, seqNum int) error {
	ackMsg := fmt.Sprintf("SR_ACK:%d", seqNum)
	_, err := conn.WriteToUDP([]byte(ackMsg), serverAddr)
	if err != nil {
		return fmt.Errorf("failed to send SR ACK: %v", err)
	}
	fmt.Printf("Sending SR ACK for packet %d\n", seqNum)
	return nil
}

// RunSRClient 主循环，处理 SR 协议数据包接收和 ACK 发送
func RunSRClient(conn *net.UDPConn, serverAddr *net.UDPAddr) {
	expectedSeqNum := 0            // 期望的序列号
	window := make(map[int]string) // 接收窗口，缓存乱序到达的数据包
	for {
		seqNum, data, err := ReceiveSRPacket(conn)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// 模拟丢包
		if SimulateSRPacketLoss() {
			fmt.Printf("Packet loss: SR SeqNum=%d\n", seqNum)
			continue
		}

		fmt.Printf("Received SR packet: SeqNum=%d, Data=%s\n", seqNum, data)

		// 正常处理数据包
		if seqNum >= expectedSeqNum && seqNum < expectedSeqNum+srWindowSize {
			// 如果是期望的包，处理并发送 ACK
			window[seqNum] = data
			if seqNum == expectedSeqNum {
				// 顺序交付数据包给应用层，并发送 ACK
				for {
					if packetData, ok := window[expectedSeqNum]; ok {
						fmt.Printf("Delivering to application: SR SeqNum=%d, Data=%s\n", expectedSeqNum, packetData)
						delete(window, expectedSeqNum)
						expectedSeqNum = (expectedSeqNum + 1) % srMaxSeqNum
					} else {
						break
					}
				}
			}
			if err := SendSRAck(conn, serverAddr, seqNum); err != nil {
				fmt.Println(err)
			}
		} else if seqNum < expectedSeqNum {
			// 如果收到的包已经被接收并确认过，发送 ACK 重复确认
			if err := SendSRAck(conn, serverAddr, seqNum); err != nil {
				fmt.Println(err)
			}
		}
	}
}
