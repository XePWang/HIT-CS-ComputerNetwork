package main

import (
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	IpA  = "192.168.226.129"   // 主机A的IP
	MacA = "00:11:22:33:44:01" // 主机A的MAC
	// IpB1       = "192.168.226.128"   // 主机B的第一个IP
	MacB1      = "00:11:22:33:44:11" // 主机B的第一个MAC
	IpC        = "192.168.88.129"    // 主机C的IP
	Interface1 = "ens33"             // 主机A的第一个网络接口
	Port       = 8080                // UDP端口号
)

func main() {
	// 解析IP地址
	ipA := net.ParseIP(IpA)
	ipC := net.ParseIP(IpC)
	// ipB1 := net.ParseIP(IpB1)

	// 解析MAC地址
	macA, _ := net.ParseMAC(MacA)
	macB1, _ := net.ParseMAC(MacB1)

	// 创建以太网帧
	eth := &layers.Ethernet{
		SrcMAC:       macA,
		DstMAC:       macB1, // 目标MAC地址是主机B1的MAC地址
		EthernetType: layers.EthernetTypeIPv4,
	}

	// 创建IP包
	ip := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		SrcIP:    ipA,
		DstIP:    ipC, // 目标IP地址是主机C的IP地址
		Protocol: layers.IPProtocolUDP,
	}

	// 创建UDP包
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(Port),
		DstPort: layers.UDPPort(Port),
	}

	// 打开网络接口
	handle, err := pcap.OpenLive(Interface1, 65536, true, time.Second)
	if err != nil {
		fmt.Printf("Error opening interface: %v\n", err)
		return
	}
	defer handle.Close()

	// 设置过滤器
	err = handle.SetBPFFilter("ip and udp and dst host 192.168.226.129")
	if err != nil {
		fmt.Printf("Error setting filter: %v\n", err)
		return
	}

	// 开始捕获数据包
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	go func() {
		for packet := range packetSource.Packets() {
			ethLayer := packet.Layer(layers.LayerTypeEthernet)
			if ethLayer == nil {
				continue
			}
			eth, _ := ethLayer.(*layers.Ethernet)

			ipLayer := packet.Layer(layers.LayerTypeIPv4)
			if ipLayer == nil {
				continue
			}
			ip, _ := ipLayer.(*layers.IPv4)

			udpLayer := packet.Layer(layers.LayerTypeUDP)
			if udpLayer == nil {
				continue
			}
			udp, _ := udpLayer.(*layers.UDP)

			if ip.DstIP.Equal(ipA) && udp.DstPort == layers.UDPPort(Port) {
				// 打印接收到的信息
				fmt.Printf("Received message from C via B:\nEthernet Frame: %s\nIP Packet: %s\nUDP Packet: %s\nPayload: %s\n\n",
					eth, ip, udp, packet.ApplicationLayer().Payload())
			}
		}
	}()

	// 主循环，允许用户多次发送消息
	for {
		// 用户输入消息
		var message string
		fmt.Print("Enter message to send to C: ")
		_, err := fmt.Scanln(&message)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// 创建应用层数据
		payload := []byte(message)

		// 组合所有层
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		}

		// 设置网络层用于计算校验和
		err2 := udp.SetNetworkLayerForChecksum(ip)
		if err2 != nil {
			fmt.Printf("Error setting network layer for checksum: %v\n", err2)
		}
		err1 := gopacket.SerializeLayers(buf, opts, eth, ip, udp, gopacket.Payload(payload))
		if err1 != nil {
			fmt.Printf("Error serializing packet: %v\n", err1)
			continue
		}

		// 打印发送的信息
		fmt.Printf("Sending:\nEthernet Frame: %s\nIP Packet: %s\nUDP Packet: %s\nPayload: %s\n",
			eth, ip, udp, payload)

		// 发送数据包
		err = handle.WritePacketData(buf.Bytes())
		if err != nil {
			fmt.Printf("Error sending packet: %v\n", err)
			continue
		}
	}
}
