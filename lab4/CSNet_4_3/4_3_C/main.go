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
	IpC        = "192.168.88.129"    // 主机C的IP
	MacC       = "00:11:22:33:44:03" // 主机C的MAC
	MacB2      = "00:11:22:33:44:12" // 主机B的第二个MAC
	Interface2 = "ens33"             // 主机C的第一个网络接口
	Port       = 8080                // UDP端口号
	IpA        = "192.168.226.129"   // 主机A的IP
)

func main() {
	// 解析IP地址
	ipC := net.ParseIP(IpC)
	ipA := net.ParseIP(IpA)

	// 解析MAC地址
	macC, _ := net.ParseMAC(MacC)
	macB2, _ := net.ParseMAC(MacB2)

	// 打开网络接口
	handle, err := pcap.OpenLive(Interface2, 65536, true, time.Second)
	if err != nil {
		fmt.Printf("Error opening interface: %v\n", err)
		return
	}
	defer handle.Close()

	// 设置过滤器
	err = handle.SetBPFFilter("ip and udp and dst host 192.168.88.129")
	if err != nil {
		fmt.Printf("Error setting filter: %v\n", err)
		return
	}

	// 开始捕获数据包
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
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

		if ip.DstIP.Equal(ipC) && udp.DstPort == layers.UDPPort(Port) {
			// 打印接收到的信息
			fmt.Printf("Received message from A via B:\nEthernet Frame: %s\nIP Packet: %s\nUDP Packet: %s\nPayload: %s\n",
				eth, ip, udp, packet.ApplicationLayer().Payload())

			// 构建回复消息
			reply := fmt.Sprintf("Hello! Got your message loud and clear. Your message is %s", packet.ApplicationLayer().Payload())

			// 创建新的以太网帧
			newEth := &layers.Ethernet{
				SrcMAC:       macC,
				DstMAC:       macB2, // 目标MAC地址是主机B2的MAC地址
				EthernetType: layers.EthernetTypeIPv4,
			}

			// 创建新的IP包
			newIP := &layers.IPv4{
				Version:  4,
				IHL:      5,
				TTL:      64,
				SrcIP:    ipC,
				DstIP:    ipA, // 目标IP地址是主机A的IP地址
				Protocol: layers.IPProtocolUDP,
			}

			// 创建新的UDP包
			newUDP := &layers.UDP{
				SrcPort: layers.UDPPort(Port),
				DstPort: udp.SrcPort,
			}

			// 设置网络层用于计算校验和
			errUDP := newUDP.SetNetworkLayerForChecksum(newIP)
			if errUDP != nil {
				fmt.Printf("Error setting udp layer For Checksum: %v\n", errUDP)
			}

			// 组合所有层
			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}

			err := gopacket.SerializeLayers(buf, opts, newEth, newIP, newUDP, gopacket.Payload(reply))
			if err != nil {
				fmt.Printf("Error serializing packet: %v\n", err)
				continue
			}

			// 打印发送的信息
			fmt.Printf("Replying to A via B:\nEthernet Frame: %s\nIP Packet: %s\nUDP Packet: %s\nPayload: %s\n",
				newEth, newIP, newUDP, reply)

			// 发送数据包
			err = handle.WritePacketData(buf.Bytes())
			if err != nil {
				fmt.Printf("Error sending packet: %v\n", err)
				continue
			}
		}
	}

	// 阻塞主goroutine以防止程序退出
	select {}
}
