# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此代码仓库中工作时提供指导。

## 项目概述

Open-UAV-Telemetry-Bridge (OUTB) 是一个协议无关的无人机遥测边缘网关。它在多种无人机协议（MAVLink、DJI、GB/T 28181）之间进行转换，并通过 MQTT、WebSocket、HTTP 或 gRPC 输出标准化数据。

**当前状态**：v0.3.0-dev 开发中。MAVLink + DJI 双协议支持，坐标转换，HTTP API。

## 构建命令

```bash
# 构建
make build                  # 本机编译
make build-linux-arm64      # 交叉编译至树莓派/Jetson

# 测试
make test                   # 运行所有测试
go test -v ./...            # 详细测试输出

# 运行
./bin/outb configs/config.yaml

# 清理
make clean
```

## 项目结构

```
├── cmd/outb/main.go                    # 程序入口
├── internal/
│   ├── models/                         # 统一数据模型 (DroneState)
│   ├── core/
│   │   ├── interfaces.go               # Adapter/Publisher 接口定义
│   │   ├── engine.go                   # 消息路由引擎
│   │   ├── coordinator/                # 坐标系转换 (WGS84→GCJ02/BD09)
│   │   ├── statestore/                 # 状态缓存
│   │   └── throttler/                  # 频率控制
│   ├── adapters/
│   │   ├── mavlink/                    # MAVLink 南向适配器 (UDP/TCP/Serial)
│   │   └── dji/                        # DJI 南向适配器 (TCP Server)
│   ├── publishers/mqtt/                # MQTT 北向发布器
│   ├── api/                            # HTTP REST API 服务器
│   └── config/                         # YAML 配置管理
├── android/dji-forwarder/              # DJI Android 转发端 (Kotlin)
├── configs/config.example.yaml         # 示例配置
├── scripts/test_dji_client.go          # DJI 协议测试客户端
├── docs/                               # 文档
└── Makefile
```

## 架构

**模式**：微内核 + 插件化架构

```
南向适配层
├── MAVLink Adapter (UDP/TCP/Serial)
└── DJI Adapter (TCP Server ← Android Forwarder)
    ↓ DroneState 事件
核心处理层
├── Engine
├── CoordinateConverter (WGS84 → GCJ02/BD09)
├── Throttler
└── StateStore
    ↓ 频率控制后的 DroneState
北向发布层
├── MQTT Publisher
└── HTTP API (REST)
```

### HTTP API 接口

```
GET /health              # 健康检查
GET /api/v1/status       # 网关状态
GET /api/v1/drones       # 所有无人机列表
GET /api/v1/drones/{id}  # 单个无人机详情
```

### DJI 通信协议

Android 转发端与 Go 网关之间使用长度前缀 JSON 协议：

```
帧格式: [4字节长度 BigEndian][JSON数据]

消息类型:
- hello: {"type":"hello","device_id":"xxx","sdk_version":"5.x"}
- state: {"type":"state","data":{...DroneState...}}
- heartbeat: {"type":"heartbeat","timestamp":123456789}
- ack: {"type":"ack"}
```

### 核心接口

```go
// internal/core/interfaces.go
type Adapter interface {
    Name() string
    Start(ctx context.Context, events chan<- *models.DroneState) error
    Stop() error
}

type Publisher interface {
    Name() string
    Start(ctx context.Context) error
    Publish(state *models.DroneState) error
    Stop() error
}
```

### 数据流

**MAVLink 路径**:
1. MAVLink Adapter 接收 UDP/TCP/Serial 数据
2. 解析 HEARTBEAT、GLOBAL_POSITION_INT、ATTITUDE、SYS_STATUS
3. 转换为 DroneState → Throttler → MQTT Publisher

**DJI 路径**:
1. Android 转发端从 DJI SDK 获取遥测数据
2. 转换为 DroneState JSON，通过 TCP 发送到 Go 网关
3. DJI Adapter 接收解析 → Engine → Throttler → MQTT Publisher

## 技术选型

**Go 网关**:

| 模块 | 库 | 版本 |
|------|-----|------|
| MAVLink | `github.com/bluenviron/gomavlib/v3` | v3.3.0 |
| MQTT | `github.com/eclipse/paho.mqtt.golang` | v1.5.1 |
| HTTP 路由 | `github.com/go-chi/chi/v5` | v5.2.4 |
| CORS | `github.com/go-chi/cors` | v1.2.2 |
| 配置 | `gopkg.in/yaml.v3` | v3.0.1 |

**Android 转发端**:

| 模块 | 技术 |
|------|------|
| 语言 | Kotlin |
| JSON | kotlinx.serialization |
| 异步 | Kotlin Coroutines |
| UI | Material Design 3 |
| 最低版本 | Android 7.0 (API 24) |

## 开发路线图

- [x] v0.1 (MVP)：MAVLink → MQTT，树莓派运行
- [x] v0.2：DJI Mobile SDK Android 转发端 (核心完成，待 SDK 集成)
- [x] v0.3：坐标系转换 (WGS84→GCJ02/BD09) + HTTP API
- [ ] v0.4：GB/T 28181 国标支持
- [ ] v1.0：Web 管理界面

## 项目指南

- 语言约定：交流与文档使用中文；生成的代码使用英文
- 发布约定：发布固定在 `/release` 文件夹
- 文档约定：
  - 未完成的修改：`/docs/progress`
  - 已完成的修改：`/docs/reports/completed`
  - 文档模板：`/docs/templates`

## 备注

- README.md 包含完整的技术规格说明（中文）
- 目标平台包括边缘设备（树莓派 4、Jetson）和云服务器
- 当前已有 14 个单元测试，全部通过
