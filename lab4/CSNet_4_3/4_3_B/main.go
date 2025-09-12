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
	IpA        = "192.168.226.129"   // 主机A的IP
	MacA       = "00:11:22:33:44:01" // 主机A的MAC
	IpB1       = "192.168.226.128"   // 主机B的第一个IP
	IpB2       = "192.168.88.128"    // 主机B的第二个IP
	IpC        = "192.168.88.129"    // 主机C的IP
	MacB1      = "00:11:22:33:44:11" // 主机B的第一个MAC
	MacB2      = "00:11:22:33:44:12" // 主机B的第二个MAC
	MacC       = "00:11:22:33:44:03" // 主机C的MAC
	Interface1 = "ens33"             // 主机B的第一个网络接口
	Interface2 = "ens37"             // 主机B的第二个网络接口
	Port       = 8080                // UDP端口号
)

func main() {
	// 解析IP地址
	ipA := net.ParseIP(IpA)
	ipB1 := net.ParseIP(IpB1)
	ipB2 := net.ParseIP(IpB2)
	ipC := net.ParseIP(IpC)

	// 解析MAC地址
	macA, _ := net.ParseMAC(MacA)
	macB1, _ := net.ParseMAC(MacB1)
	macB2, _ := net.ParseMAC(MacB2)
	macC, _ := net.ParseMAC(MacC)

	// 打开网络接口
	handle1, err := pcap.OpenLive(Interface1, 65536, true, time.Second)
	if err != nil {
		fmt.Printf("Error opening input interface: %v\n", err)
		return
	}
	defer handle1.Close()

	handle2, err := pcap.OpenLive(Interface2, 65536, true, time.Second)
	if err != nil {
		fmt.Printf("Error opening output interface: %v\n", err)
		return
	}
	defer handle2.Close()

	// 设置过滤器
	err = handle1.SetBPFFilter("ip and udp")
	if err != nil {
		fmt.Printf("Error setting filter on input interface: %v\n", err)
		return
	}

	err = handle2.SetBPFFilter("ip and udp")
	if err != nil {
		fmt.Printf("Error setting filter on output interface: %v\n", err)
		return
	}

	// 开始捕获数据包
	packetSource1 := gopacket.NewPacketSource(handle1, handle1.LinkType())
	packetSource2 := gopacket.NewPacketSource(handle2, handle2.LinkType())

	go func() {
		for packet := range packetSource1.Packets() {
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

			if ip.DstIP.Equal(ipB1) && udp.DstPort == layers.UDPPort(Port) {
				// 打印接收到的信息
				fmt.Printf("Received message from A:\nEthernet Frame: %s\nIP Packet: %s\nUDP Packet: %s\nPayload: %s\n",
					eth, ip, udp, packet.ApplicationLayer().Payload())

				// 创建新的以太网帧
				newEth := &layers.Ethernet{
					SrcMAC:       macB2,
					DstMAC:       macC,
					EthernetType: layers.EthernetTypeIPv4,
				}

				// 创建新的IP包
				newIP := &layers.IPv4{
					Version:  4,
					IHL:      5,
					TTL:      ip.TTL - 1, // 减少TTL
					SrcIP:    ipA,
					DstIP:    ipC,
					Protocol: layers.IPProtocolUDP,
				}

				// 创建新的UDP包
				newUDP := &layers.UDP{
					SrcPort: udp.SrcPort,
					DstPort: layers.UDPPort(Port),
				}

				// 设置网络层用于计算校验和
				errUDP1 := newUDP.SetNetworkLayerForChecksum(newIP)
				if errUDP1 != nil {
					fmt.Printf("Error setting udp layer for checksum: %v\n", errUDP1)
				}

				// 组合所有层
				buf := gopacket.NewSerializeBuffer()
				opts := gopacket.SerializeOptions{
					FixLengths:       true,
					ComputeChecksums: true,
				}

				err := gopacket.SerializeLayers(buf, opts, newEth, newIP, newUDP, gopacket.Payload(packet.ApplicationLayer().Payload()))
				if err != nil {
					fmt.Printf("Error serializing packet: %v\n", err)
					continue
				}

				// 打印发送的信息
				fmt.Printf("Forwarding to C:\nEthernet Frame: %s\nIP Packet: %s\nUDP Packet: %s\nPayload: %s\n",
					newEth, newIP, newUDP, packet.ApplicationLayer().Payload())

				// 发送数据包
				err = handle2.WritePacketData(buf.Bytes())
				if err != nil {
					fmt.Printf("Error sending packet: %v\n", err)
					continue
				}
			}
		}
	}()

	go func() {
		for packet := range packetSource2.Packets() {
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

			if ip.DstIP.Equal(ipB2) && udp.DstPort == layers.UDPPort(Port) {
				// 打印接收到的信息
				fmt.Printf("Received message from C:\nEthernet Frame: %s\nIP Packet: %s\nUDP Packet: %s\nPayload: %s\n",
					eth, ip, udp, packet.ApplicationLayer().Payload())

				// 创建新的以太网帧
				newEth := &layers.Ethernet{
					SrcMAC:       macB1,
					DstMAC:       macA,
					EthernetType: layers.EthernetTypeIPv4,
				}

				// 创建新的IP包
				newIP := &layers.IPv4{
					Version:  4,
					IHL:      5,
					TTL:      ip.TTL - 1, // 减少TTL
					SrcIP:    ipC,
					DstIP:    ipA,
					Protocol: layers.IPProtocolUDP,
				}

				// 创建新的UDP包
				newUDP := &layers.UDP{
					SrcPort: udp.SrcPort,
					DstPort: layers.UDPPort(Port),
				}

				// 设置网络层用于计算校验和
				errUDP2 := newUDP.SetNetworkLayerForChecksum(newIP)
				if errUDP2 != nil {
					fmt.Printf("Error setting udp layer for checksum: %v\n", errUDP2)
				}

				// 组合所有层
				buf := gopacket.NewSerializeBuffer()
				opts := gopacket.SerializeOptions{
					FixLengths:       true,
					ComputeChecksums: true,
				}

				err := gopacket.SerializeLayers(buf, opts, newEth, newIP, newUDP, gopacket.Payload(packet.ApplicationLayer().Payload()))
				if err != nil {
					fmt.Printf("Error serializing packet: %v\n", err)
					continue
				}

				// 打印发送的信息
				fmt.Printf("Forwarding to A:\nEthernet Frame: %s\nIP Packet: %s\nUDP Packet: %s\nPayload: %s\n",
					newEth, newIP, newUDP, packet.ApplicationLayer().Payload())

				// 发送数据包
				err = handle1.WritePacketData(buf.Bytes())
				if err != nil {
					fmt.Printf("Error sending packet: %v\n", err)
					continue
				}
			}
		}
	}()

	// 阻塞主goroutine以防止程序退出
	select {}
}
