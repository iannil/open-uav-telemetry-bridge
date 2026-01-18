# Open-UAV-Telemetry-Bridge

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)]()

**协议无关的无人机遥测边缘网关** - 连接无人机协议与云平台的桥梁

[English](README.md)

---

## 项目简介

Open-UAV-Telemetry-Bridge（简称 OUTB）是一个轻量级、高性能的遥测网关，专为低空物联网（IoD, Internet of Drones）设计。它在多种无人机协议（MAVLink、DJI、GB/T 28181）之间进行转换，并通过 MQTT、WebSocket、HTTP 或 gRPC 输出标准化数据。

### 为什么需要 OUTB？

- **协议碎片化**：PX4 使用 MAVLink，大疆使用私有协议，政府平台要求 GB/T 28181
- **坐标系混乱**：GPS 输出 WGS84，但国内地图需要 GCJ02/BD09 偏移
- **带宽限制**：原始遥测数据 50-100Hz 对 4G 网络来说太多了
- **集成复杂**：每个平台都需要定制适配器

OUTB 通过统一的插件化架构解决所有这些问题。

---

## 功能特性

### 核心功能

- **多协议支持**：MAVLink（UDP/TCP/串口）、DJI（通过 Android 转发端）、GB/T 28181（计划中）
- **统一数据模型**：无论源协议如何，都输出标准化 JSON
- **坐标转换**：自动 WGS84 → GCJ02/BD09 转换，适配国内地图
- **频率控制**：可配置降采样（如 50Hz → 1Hz），节省带宽
- **状态缓存**：内存状态存储 + 历史轨迹存储

### 输出接口

- **MQTT 发布器**：标准 MQTT 3.1.1，支持遗嘱消息（LWT）
- **HTTP REST API**：查询无人机状态、健康检查、网关状态
- **WebSocket**：实时状态推送
- **轨迹存储**：环形缓冲区历史轨迹（可配置保留数量）

### 运维特性

- **边缘就绪**：可运行在树莓派 4、Jetson Nano 或云服务器
- **零依赖**：单一二进制文件，无需外部运行时
- **热配置**：基于 YAML 的配置文件

---

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- （可选）MQTT Broker（如 Mosquitto）

### 安装

```bash
# 克隆仓库
git clone https://github.com/iannil/open-uav-telemetry-bridge.git
cd open-uav-telemetry-bridge

# 构建
make build

# 或交叉编译到树莓派/Jetson
make build-linux-arm64
```

### 配置

```bash
# 复制示例配置
cp configs/config.example.yaml configs/config.yaml

# 按需编辑
vim configs/config.yaml
```

### 运行

```bash
# 使用配置文件运行
./bin/outb configs/config.yaml
```

### 验证

```bash
# 健康检查
curl http://localhost:8080/health

# 获取网关状态
curl http://localhost:8080/api/v1/status

# 列出已连接的无人机
curl http://localhost:8080/api/v1/drones
```

---

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      北向接口层                              │
│  ┌──────────┐  ┌───────────┐  ┌──────────┐  ┌────────────┐  │
│  │   MQTT   │  │ WebSocket │  │   HTTP   │  │    gRPC    │  │
│  └────┬─────┘  └─────┬─────┘  └────┬─────┘  └─────┬──────┘  │
└───────┼──────────────┼─────────────┼──────────────┼─────────┘
        │              │             │              │
        └──────────────┴──────┬──────┴──────────────┘
                              │
┌─────────────────────────────┼───────────────────────────────┐
│                        核心引擎                              │
│  ┌─────────────┐  ┌─────────┴────────┐  ┌────────────────┐  │
│  │  频率控制器  │  │     状态存储      │  │   轨迹存储     │  │
│  └─────────────┘  └──────────────────┘  └────────────────┘  │
│  ┌─────────────────────────────────────────────────────────┐│
│  │            坐标转换器 (WGS84→GCJ02/BD09)                ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────┼───────────────────────────────┐
│                      南向适配层                              │
│  ┌──────────────┐  ┌────────────────┐  ┌─────────────────┐  │
│  │   MAVLink    │  │   DJI 转发端    │  │   GB/T 28181   │  │
│  │(UDP/TCP/串口)│  │  (TCP 服务器)   │  │    (计划中)    │  │
│  └──────┬───────┘  └───────┬────────┘  └────────┬────────┘  │
└─────────┼──────────────────┼────────────────────┼───────────┘
          │                  │                    │
     ┌────┴────┐      ┌──────┴──────┐      ┌──────┴──────┐
     │PX4/Ardu │      │   大疆无人机  │      │  政府平台   │
     │ Pilot   │      │  (通过 APP)  │      │            │
     └─────────┘      └─────────────┘      └────────────┘
```

---

## API 参考

### HTTP 接口

| 方法 | 端点 | 描述 |
| ------ | ------ | ------ |
| GET | `/health` | 健康检查 |
| GET | `/api/v1/status` | 网关状态和统计信息 |
| GET | `/api/v1/drones` | 列出所有已连接的无人机 |
| GET | `/api/v1/drones/{id}` | 获取指定无人机状态 |
| GET | `/api/v1/drones/{id}/track` | 获取历史轨迹点 |
| DELETE | `/api/v1/drones/{id}/track` | 清除轨迹历史 |

### WebSocket

连接 `ws://localhost:8080/api/v1/ws` 获取实时更新。

**消息类型：**

```json
// 状态更新（服务端 → 客户端）
{
  "type": "state_update",
  "data": { /* DroneState */ }
}

// 订阅特定无人机（客户端 → 服务端）
{
  "type": "subscribe",
  "device_ids": ["drone-001", "drone-002"]
}

// 取消订阅（客户端 → 服务端）
{
  "type": "unsubscribe",
  "device_ids": ["drone-001"]
}
```

### 统一数据模型（DroneState）

```json
{
  "device_id": "mavlink-001",
  "timestamp": 1709882231000,
  "protocol_source": "mavlink",
  "location": {
    "lat": 39.9042,
    "lon": 116.4074,
    "lat_gcj02": 39.9066,
    "lon_gcj02": 116.4136,
    "alt": 120.5,
    "coordinate_system": "WGS84"
  },
  "attitude": {
    "roll": 0.05,
    "pitch": -0.12,
    "yaw": 180.0
  },
  "velocity": {
    "vx": 10.5,
    "vy": 0.0,
    "vz": -0.5
  },
  "status": {
    "battery_percent": 85,
    "flight_mode": "AUTO",
    "armed": true,
    "signal_quality": 95
  }
}
```

---

## 配置说明

```yaml
# 服务器设置
server:
  log_level: info  # debug, info, warn, error

# MAVLink 适配器
mavlink:
  enabled: true
  connection_type: udp   # udp | tcp | serial
  address: "0.0.0.0:14550"

# DJI 转发端适配器
dji:
  enabled: false
  listen_address: "0.0.0.0:14560"
  max_clients: 10

# MQTT 发布器
mqtt:
  enabled: true
  broker: "tcp://localhost:1883"
  client_id: "outb-001"
  topic_prefix: "uav/telemetry"
  qos: 1
  lwt:
    enabled: true
    topic: "uav/status"
    message: "offline"

# HTTP API
http:
  enabled: true
  address: "0.0.0.0:8080"
  cors_enabled: true
  cors_origins: ["*"]

# 频率控制
throttle:
  default_rate_hz: 1.0
  min_rate_hz: 0.5
  max_rate_hz: 10.0

# 坐标转换（国内地图）
coordinate:
  convert_gcj02: true   # 高德、腾讯、谷歌中国
  convert_bd09: false   # 百度地图

# 轨迹存储
track:
  enabled: true
  max_points_per_drone: 10000
  sample_interval_ms: 1000
```

---

## 部署场景

### 1. 机载边缘网关

部署在挂载于无人机上的树莓派或 Jetson。

```
[飞控] --串口--> [树莓派上的 OUTB] --4G--> [云端]
```

### 2. 地面站网关

与地面站软件一起运行。

```
[无人机] --无线电--> [地面站 + OUTB] --WiFi/4G--> [云端]
```

### 3. 云端聚合器

用于机队管理的集中式协议转换。

```
[无人机机队] --TCP/UDP--> [云端 OUTB] --> [后端服务]
```

---

## 开发路线图

- [x] **v0.1** - MAVLink → MQTT 基础管道
- [x] **v0.2** - DJI Android 转发端应用
- [x] **v0.3** - 坐标转换 + HTTP API
- [x] **v0.3.1** - WebSocket + 轨迹存储
- [x] **v0.4** - GB/T 28181 国标支持
- [ ] **v1.0** - Web 管理界面

---

## 项目结构

```
├── cmd/outb/               # 应用程序入口
├── internal/
│   ├── adapters/           # 南向协议适配器
│   │   ├── mavlink/        # MAVLink (UDP/TCP/串口)
│   │   └── dji/            # DJI 转发端 (TCP 服务器)
│   ├── api/                # HTTP/WebSocket 服务器
│   ├── config/             # YAML 配置
│   ├── core/               # 核心引擎
│   │   ├── coordinator/    # 坐标转换
│   │   ├── statestore/     # 状态缓存
│   │   ├── throttler/      # 频率控制
│   │   └── trackstore/     # 历史轨迹
│   ├── models/             # 统一数据模型
│   └── publishers/         # 北向发布器
│       └── mqtt/           # MQTT 发布器
├── android/                # DJI Android 转发端 (Kotlin)
├── configs/                # 配置示例
├── scripts/                # 测试工具
└── docs/                   # 文档
```

---

## 技术选型

| 模块 | 技术 | 版本 |
| ------ | ------ | ------ |
| MAVLink 解析 | gomavlib | v3.3.0 |
| MQTT 客户端 | paho.mqtt.golang | v1.5.1 |
| HTTP 路由 | chi | v5.2.4 |
| WebSocket | gorilla/websocket | v1.5.3 |
| 配置管理 | yaml.v3 | v3.0.1 |

---

## 参与贡献

欢迎贡献代码！提交 PR 前请阅读贡献指南。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

---

## 开源协议

本项目采用 Apache License 2.0 协议 - 详见 [LICENSE](LICENSE) 文件。

---

## 致谢

- [gomavlib](https://github.com/bluenviron/gomavlib) - Go MAVLink 库
- [paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) - Eclipse Paho MQTT 客户端
- [chi](https://github.com/go-chi/chi) - 轻量级 HTTP 路由
