package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

const (
	HostAIP      = "192.168.245.133"
	HostBIP      = "192.168.245.134"
	HostCIP      = "192.168.245.135"
	HostDIP      = "192.168.245.136"
	HostEIP      = "192.168.245.137"
	HostAMAC     = "00:11:22:33:44:01"
	HostBMAC     = "00:11:22:33:44:02"
	HostCMAC     = "00:11:22:33:44:03"
	HostDMAC     = "00:11:22:33:44:04"
	HostEMAC     = "00:11:22:33:44:05"
	EthernetType = 0x0800 // IPv4
	Port         = "8080"
)

type EthernetFrame struct {
	DestinationMAC [6]byte
	SourceMAC      [6]byte
	Type           uint16
	Payload        []byte
}

type IPPacket struct {
	VersionIHL     uint8  // 版本 + 头部长度
	TypeOfService  uint8  // 服务类型
	TotalLength    uint16 // 总长度
	Identification uint16 // 标识符
	FlagsAndOffset uint16 // 标志 + 片偏移
	TTL            uint8  // 存活时间
	Protocol       uint8  // 协议类型（如：TCP为6）
	HeaderChecksum uint16 // 头部校验和
	SourceIP       [4]byte
	DestinationIP  [4]byte
	Payload        []byte
}

func printEthernetFrame(frame EthernetFrame) {
	fmt.Printf("Ethernet Frame:\n")
	fmt.Printf("Destination MAC: %s\n", macToString(frame.DestinationMAC))
	fmt.Printf("Source MAC: %s\n", macToString(frame.SourceMAC))
	fmt.Printf("Type: 0x%X\n", frame.Type)
	fmt.Printf("Payload: %x\n\n", frame.Payload)
}

func printIPPacket(packet IPPacket) {
	fmt.Printf("IP Packet:\n")
	fmt.Printf("Version & IHL: 0x%X\n", packet.VersionIHL)
	fmt.Printf("Type of Service: 0x%X\n", packet.TypeOfService)
	fmt.Printf("Total Length: %d\n", packet.TotalLength)
	fmt.Printf("Identification: 0x%X\n", packet.Identification)
	fmt.Printf("Flags & Fragment Offset: 0x%X\n", packet.FlagsAndOffset)
	fmt.Printf("TTL: %d\n", packet.TTL)
	fmt.Printf("Protocol: %d\n", packet.Protocol)
	fmt.Printf("Header Checksum: 0x%X\n", packet.HeaderChecksum)
	fmt.Printf("Source IP: %s\n", ipToString(packet.SourceIP))
	fmt.Printf("Destination IP: %s\n", ipToString(packet.DestinationIP))
	fmt.Printf("Payload: %s\n\n", packet.Payload)
}

func macToString(mac [6]byte) string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

func ipToString(ip [4]byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func createEthernetFrame(sourceMAC, destMAC [6]byte, ipPacket IPPacket) EthernetFrame {
	payload := serializeIPPacket(ipPacket)
	return EthernetFrame{
		DestinationMAC: destMAC,
		SourceMAC:      sourceMAC,
		Type:           EthernetType,
		Payload:        payload,
	}
}

func serializeIPPacket(packet IPPacket) []byte {
	payloadSize := int(packet.TotalLength)
	ipPacketBytes := make([]byte, payloadSize)

	offset := 0
	binary.BigEndian.PutUint16(ipPacketBytes[offset:], uint16(packet.VersionIHL)<<8|uint16(packet.TypeOfService))
	offset += 2
	binary.BigEndian.PutUint16(ipPacketBytes[offset:], packet.TotalLength)
	offset += 2
	binary.BigEndian.PutUint16(ipPacketBytes[offset:], packet.Identification)
	offset += 2
	binary.BigEndian.PutUint16(ipPacketBytes[offset:], packet.FlagsAndOffset)
	offset += 2
	ipPacketBytes[offset] = packet.TTL
	offset++
	ipPacketBytes[offset] = packet.Protocol
	offset++
	binary.BigEndian.PutUint16(ipPacketBytes[offset:], packet.HeaderChecksum)
	offset += 2
	copy(ipPacketBytes[offset:], packet.SourceIP[:])
	offset += 4
	copy(ipPacketBytes[offset:], packet.DestinationIP[:])
	offset += 4

	copy(ipPacketBytes[offset:], packet.Payload)

	return ipPacketBytes
}

func parseEthernetFrame(frame EthernetFrame) (IPPacket, error) {
	ipPacket := IPPacket{}
	payload := frame.Payload

	offset := 0
	versionIHL := binary.BigEndian.Uint16(payload[offset:])
	ipPacket.VersionIHL = uint8(versionIHL >> 8)
	ipPacket.TypeOfService = uint8(versionIHL & 0xFF)
	offset += 2
	ipPacket.TotalLength = binary.BigEndian.Uint16(payload[offset:])
	offset += 2
	ipPacket.Identification = binary.BigEndian.Uint16(payload[offset:])
	offset += 2
	ipPacket.FlagsAndOffset = binary.BigEndian.Uint16(payload[offset:])
	offset += 2
	ipPacket.TTL = payload[offset]
	offset++
	ipPacket.Protocol = payload[offset]
	offset++
	ipPacket.HeaderChecksum = binary.BigEndian.Uint16(payload[offset:])
	offset += 2
	copy(ipPacket.SourceIP[:], payload[offset:])
	offset += 4
	copy(ipPacket.DestinationIP[:], payload[offset:])
	offset += 4

	ipPacket.Payload = payload[offset:]

	return ipPacket, nil
}

func createIPPacket(sourceIP, destIP [4]byte, payload []byte) IPPacket {
	return IPPacket{
		VersionIHL:     0x45, // IPv4, IHL=5
		TypeOfService:  0x00, // 默认
		TotalLength:    uint16(20 + len(payload)),
		Identification: 0x1234,
		FlagsAndOffset: 0x4000, // 不分片
		TTL:            64,     // TTL
		Protocol:       17,     // UDP协议
		HeaderChecksum: 0xFFFF, // 假设校验和
		SourceIP:       sourceIP,
		DestinationIP:  destIP,
		Payload:        payload,
	}
}

func stringToIPBytes(ipStr string) [4]byte {
	var result [4]byte
	_, err := fmt.Sscanf(ipStr, "%d.%d.%d.%d", &result[0], &result[1], &result[2], &result[3])
	if err != nil {
		fmt.Println("Invalid IP address format")
		os.Exit(1)
	}
	return result
}

func stringToMACBytes(macStr string) [6]byte {
	var result [6]byte
	_, err := fmt.Sscanf(macStr, "%02X:%02X:%02X:%02X:%02X:%02X", &result[0], &result[1], &result[2], &result[3], &result[4], &result[5])
	if err != nil {
		fmt.Println("Invalid MAC address format")
		os.Exit(1)
	}
	return result
}

func main() {
	// 监听UDP端口
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

	for {
		buffer := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading UDP message:", err)
			continue
		}

		// 解析以太网帧
		ethernetFrame := EthernetFrame{
			Payload: buffer[:n],
		}
		fmt.Println("C主机接收到消息:")
		printEthernetFrame(ethernetFrame)

		// 解析IP包
		ipPacket, err := parseEthernetFrame(ethernetFrame)
		if err != nil {
			fmt.Println("Error parsing IP packet:", err)
			continue
		}
		fmt.Println("C主机解析IP包:")
		printIPPacket(ipPacket)

		// 转发到D主机
		if ipToString(ipPacket.DestinationIP) == HostEIP {
			ethernetFrame = createEthernetFrame(stringToMACBytes(HostCMAC), stringToMACBytes(HostDMAC), ipPacket)
			fmt.Println("C主机转发消息到D主机:")
			printEthernetFrame(ethernetFrame)

			// 发送以太网帧到D主机
			sendFrameToHost(ethernetFrame, HostDIP)
		}
	}
}

func sendFrameToHost(frame EthernetFrame, hostIP string) {
	udpAddr, err := net.ResolveUDPAddr("udp", hostIP+":"+Port)
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
			fmt.Println("Error closing connection:", err)
		}
	}()

	_, err = conn.Write(frame.Payload)
	if err != nil {
		fmt.Println("Error sending UDP message:", err)
	}
}
