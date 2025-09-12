# 实验二：可靠数据传输协议

## 简介

本实验旨在通过Go-Back-N（GBN）和Selective Repeat（SR）协议在不可靠的UDP上实现可靠的数据传输。实验内容包括单向和双向的数据传输，以及一个基于这些协议的文件传输应用。

## 文件结构

- `Client1/`: GBN和SR客户端的源代码。
  - `go.mod`: Go模块文件。
  - `main.go`: 客户端主程序，可通过修改布尔变量`useSR`来切换GBN和SR协议。
  - `funClient/`:
    - `GBNClient.go`: GBN客户端逻辑。
    - `SRClient.go`: SR客户端逻辑。
- `Clients/`: 文件传输应用的客户端。
  - `go.mod`: Go模块文件。
  - `main.go`: 客户端主程序。
  - `doClient/`:
    - `doClient.go`: 客户端功能实现，支持`LIST`、`GET`和`PUSH`命令。
- `Server1/`: GBN和SR服务器的源代码。
  - `go.mod`: Go模块文件。
  - `main.go`: 服务器主程序。
  - `funServer/`:
    - `GBNServer.go`: GBN服务器逻辑。
    - `SRServer.go`: SR服务器逻辑。
- `Server2/`: 文件传输应用的服务器。
  - `go.mod`: Go模块文件。
  - `main.go`: 服务器主程序。
  - `doServer/`:
    - `doServer.go`: 服务器功能实现，处理客户端的文件操作命令。

## 功能

- **GBN协议**: 实现了Go-Back-N协议，支持单向可靠数据传输，并能处理模拟的丢包情况。
- **SR协议**: 实现了Selective Repeat协议，相比GBN更高效，仅重传丢失的数据包。
- **协议切换**: 在`Client1`中，可以方便地通过修改代码中的`useSR`变量来切换使用GBN还是SR协议。
- **文件传输应用**: 一个C/S结构的应用，支持：
  - `LIST`: 查看服务器上的文件列表。
  - `GET <filename>`: 从服务器下载文件。
  - `PUSH <filename>`: 上传文件到服务器。

## 如何运行

### GBN/SR 数据传输

1. 运行服务器:
   ```bash
   cd lab2/Server1
   go run main.go
   ```
2. 运行客户端 (可选择GBN或SR):
   ```bash
   cd lab2/Client1
   # 在 main.go 中设置 useSR 的值
   go run main.go
   ```

### 文件传输应用

1. 运行服务器:
   ```bash
   cd lab2/Server2
   go run main.go
   ```
2. 运行客户端并使用命令:
   ```bash
   cd lab2/Clients
   go run main.go
   ```

## 仓库所有者

[XePWang](https://github.com/XePWang)
