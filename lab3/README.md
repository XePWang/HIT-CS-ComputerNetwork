# 实验三：典型协议的抓包分析

## 简介

本实验使用 Wireshark 工具对典型的网络协议进行抓包和分析，以加深对协议工作原理的理解。实验涵盖了应用层、传输层和网络层的多种协议。

## 文件结构

- `alice.txt`: 用于TCP协议分析的文本文件。
- `*.pcapng`: Wireshark抓包数据文件，包含了对不同协议的抓包结果。
  - `ARP.pcapng`: ARP 协议抓包数据。
  - `dns.pcapng`: DNS 协议抓包数据。
  - `HTTP.pcapng`: HTTP 协议抓包数据。
  - `icmp.pcapng`: ICMP 协议抓包数据。
  - `IP.pcapng`: IP 协议抓包数据。
  - `TCP.pcapng`: TCP 协议抓包数据。
  - `udp.pcapng`: UDP 协议抓包数据。

## 分析的协议

本实验主要分析了以下协议：

- **HTTP**: 超文本传输协议，分析了GET请求、响应、条件GET等。
- **TCP**: 传输控制协议，分析了三次握手、数据传输、拥塞控制和连接关闭。
- **IP**: 网际协议，研究了IP包头字段、分片等。
- **DNS**: 域名系统，分析了域名解析的查询和响应过程。
- **UDP**: 用户数据报协议，以QQ消息为例分析了其报文格式和无连接特性。
- **ARP**: 地址解析协议，分析了ARP请求和响应，以及ARP缓存。

## 如何使用

1. 安装 [Wireshark](https://www.wireshark.org/)。
2. 打开对应的 `.pcapng` 文件，使用Wireshark进行分析。可以根据需要使用过滤器（如 `http`, `tcp`, `dns`）来查看特定协议的报文。

## 仓库所有者

[XePWang](https://github.com/XePWang)
