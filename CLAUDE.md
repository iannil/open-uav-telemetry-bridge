# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此代码仓库中工作时提供指导。

## 项目概述

Open-UAV-Telemetry-Bridge (OUTB) 是一个协议无关的无人机遥测边缘网关。它在多种无人机协议（MAVLink、DJI、GB/T 28181）之间进行转换，并通过 MQTT、WebSocket、HTTP 或 gRPC 输出标准化数据。

**当前状态**：v0.1.0 已发布。MAVLink → MQTT 核心功能已实现。

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
│   │   ├── statestore/                 # 状态缓存
│   │   └── throttler/                  # 频率控制
│   ├── adapters/mavlink/               # MAVLink 南向适配器
│   ├── publishers/mqtt/                # MQTT 北向发布器
│   └── config/                         # YAML 配置管理
├── configs/config.example.yaml         # 示例配置
├── release/v0.1/                       # 发布包
├── docs/                               # 文档
│   ├── QUICKSTART.md                   # 快速开始指南
│   └── progress/                       # 开发进度文档
└── Makefile
```

## 架构

**模式**：微内核 + 插件化架构

```
南向适配层 (MAVLink Adapter)
    ↓ DroneState 事件
核心处理层 (Engine → Throttler → StateStore)
    ↓ 频率控制后的 DroneState
北向发布层 (MQTT Publisher)
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

1. MAVLink Adapter 接收 UDP/TCP/Serial 数据
2. 解析 HEARTBEAT、GLOBAL_POSITION_INT、ATTITUDE、SYS_STATUS
3. 转换为 DroneState 结构体
4. 经过 Throttler 频率控制
5. MQTT Publisher 发布到 Broker

## 技术选型

| 模块 | 库 | 版本 |
|------|-----|------|
| MAVLink | `github.com/bluenviron/gomavlib/v3` | v3.3.0 |
| MQTT | `github.com/eclipse/paho.mqtt.golang` | v1.5.1 |
| 配置 | `gopkg.in/yaml.v3` | v3.0.1 |

## 开发路线图

- [x] v0.1 (MVP)：MAVLink → MQTT，树莓派运行
- [ ] v0.2：DJI Mobile SDK Android 转发端
- [ ] v0.3：坐标系转换 + HTTP API
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
