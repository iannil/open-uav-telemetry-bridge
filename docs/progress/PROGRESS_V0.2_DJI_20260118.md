# v0.2 DJI Android 转发端开发计划

## 元信息

| 属性 | 值 |
|------|-----|
| 功能模块 | DJI Mobile SDK Android 转发端 |
| 创建时间 | 2026-01-18 17:10 |
| 最后更新 | 2026-01-18 18:45 |
| 作者 | Claude Code |
| 状态 | **核心实现完成** |
| 依赖版本 | v0.1.0 |
| Git Commits | `6d52c74`, `6a347d4` |

## 目标

**v0.2**: 开发 Android 应用，从 DJI 遥控器获取无人机遥测数据，转发至 OUTB 网关。

## 背景分析

### DJI 数据获取方案

DJI 是私有协议，官方不公开通信细节。获取数据的可行方案：

| 方案 | 说明 | 优缺点 |
|------|------|--------|
| **Mobile SDK** | 在遥控器连接的手机/平板上运行 APP | ✅ 官方支持，稳定；❌ 需要 Android 开发 |
| **Onboard SDK** | 在机载电脑上运行 | ✅ 直接访问；❌ 仅限特定机型，需硬件改装 |
| **Payload SDK** | 开发负载设备 | ❌ 复杂度高，不适合遥测场景 |

**选择方案**: Mobile SDK Android 应用

### 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        DJI 无人机                            │
└─────────────────────────────────────────────────────────────┘
                              │ 私有协议
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      DJI 遥控器                              │
└─────────────────────────────────────────────────────────────┘
                              │ USB / WiFi
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Android 手机/平板 (DJI Pilot APP)               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │           OUTB DJI Forwarder APP                      │  │
│  │  ┌─────────────┐    ┌─────────────┐    ┌───────────┐  │  │
│  │  │ DJI Mobile  │───▶│   数据转换   │───▶│  TCP/UDP  │  │  │
│  │  │    SDK      │    │  DroneState │    │  Client   │  │  │
│  │  └─────────────┘    └─────────────┘    └───────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │ TCP/UDP over WiFi/4G
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    OUTB 网关 (Go)                            │
│  ┌─────────────┐    ┌─────────────┐    ┌───────────────┐   │
│  │ DJI Adapter │───▶│    Core     │───▶│ MQTT Publisher│   │
│  │ (TCP Server)│    │   Engine    │    │               │   │
│  └─────────────┘    └─────────────┘    └───────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 实现进展汇总

> 核心实现已完成，等待 DJI 开发者账号注册后集成真实 SDK。

### 已完成 (2026-01-18 18:30)

| 阶段 | 完成情况 | 备注 |
|------|----------|------|
| 阶段 5: Go DJI Adapter | ✅ 完成 | `internal/adapters/dji/` |
| 阶段 3: 数据格式转换 | ✅ 完成 | Kotlin DroneState |
| 阶段 4: 网络传输层 | ✅ 完成 | TCP Client + 重连机制 |
| 阶段 6: Android UI | ✅ 完成 | Material Design 配置界面 |
| 阶段 1: 环境准备 | ✅ 项目框架完成 | DJI SDK placeholder |
| 阶段 2: 遥测数据采集 | ⏳ 等待 SDK | 模拟模式可用 |
| 阶段 7: 测试与发布 | ⏳ 进行中 | 端到端测试中 |

### Go 端实现 (`internal/adapters/dji/`)

```
internal/adapters/dji/
├── adapter.go    # TCP Server + 多客户端管理
├── protocol.go   # 消息协议定义
└── client.go     # 客户端连接处理
```

**关键特性**:
- 长度前缀 JSON 协议 (4 字节长度 + JSON 数据)
- 多客户端并发处理
- 心跳超时检测 (90 秒)
- 自动 ACK 响应

### Android 端实现 (`android/dji-forwarder/`)

```
app/src/main/java/com/outb/dji/
├── model/
│   ├── DroneState.kt        # 统一数据模型
│   └── Message.kt           # 协议消息
├── network/
│   └── TcpClient.kt         # TCP 客户端 + 自动重连
├── dji/
│   └── DJIManager.kt        # DJI SDK 封装 + 模拟模式
├── service/
│   └── ForwarderService.kt  # 前台服务
└── ui/
    └── MainActivity.kt      # 配置界面
```

**关键特性**:
- Kotlin + Coroutines 异步处理
- kotlinx.serialization JSON 序列化
- 自动重连机制 (指数退避)
- 心跳保活 (30 秒)
- 前台服务防止后台杀死
- 模拟模式便于无硬件测试

### 待完成

1. **DJI SDK 真实集成**: 需注册 DJI 开发者账号获取 App Key
2. **端到端测试**: Go 网关 + Android 模拟模式通信验证
3. **APK 签名发布**: 生成 release APK

---

## 开发阶段划分

### 阶段 1: 环境准备与 SDK 集成

**状态**: ✅ 项目框架完成 | ⏳ SDK 集成等待开发者账号

**预期产出**: 可运行的 Android 项目，能连接 DJI SDK

| 序号 | 任务 | 优先级 |
|------|------|--------|
| 1.1 | 注册 DJI 开发者账号，获取 App Key | P0 |
| 1.2 | 创建 Android Studio 项目 | P0 |
| 1.3 | 集成 DJI Mobile SDK v5 (MSDK) | P0 |
| 1.4 | 实现 SDK 初始化与注册 | P0 |
| 1.5 | 实现飞机连接状态监听 | P0 |
| 1.6 | 在模拟器/真机验证 SDK 连接 | P0 |

**技术要点**:
- DJI Mobile SDK v5 (最新版本)
- minSdkVersion: 24 (Android 7.0)
- targetSdkVersion: 34 (Android 14)
- Kotlin 优先，兼容 Java

---

### 阶段 2: 遥测数据采集

**预期产出**: 能从 DJI SDK 获取所有遥测数据

| 序号 | 任务 | 优先级 |
|------|------|--------|
| 2.1 | 获取 GPS 位置 (FlightControllerState) | P0 |
| 2.2 | 获取姿态数据 (Attitude) | P0 |
| 2.3 | 获取电池状态 (BatteryState) | P0 |
| 2.4 | 获取飞行模式 (FlightMode) | P0 |
| 2.5 | 获取遥控器信号强度 | P1 |
| 2.6 | 获取返航点信息 | P1 |
| 2.7 | 实现数据变化监听器 | P0 |

**DJI SDK 数据映射**:

| DJI SDK 字段 | DroneState 字段 |
|--------------|-----------------|
| `location.latitude` | `location.lat` |
| `location.longitude` | `location.lon` |
| `altitude` | `location.alt_gnss` |
| `attitude.pitch/roll/yaw` | `attitude.*` |
| `velocityX/Y/Z` | `velocity.*` |
| `batteryPercentage` | `status.battery_percent` |
| `flightMode` | `status.flight_mode` |
| `areMotorsOn` | `status.armed` |
| `signalQuality` | `status.signal_quality` |

---

### 阶段 3: 数据格式转换

**预期产出**: DJI 数据转换为 OUTB DroneState JSON

| 序号 | 任务 | 优先级 |
|------|------|--------|
| 3.1 | 定义 DroneState Kotlin 数据类 | P0 |
| 3.2 | 实现 DJI → DroneState 转换器 | P0 |
| 3.3 | 实现 FlightMode 映射表 | P0 |
| 3.4 | 实现 JSON 序列化 (kotlinx.serialization) | P0 |
| 3.5 | 单元测试 | P0 |

**DJI FlightMode 映射**:

| DJI FlightMode | OUTB FlightMode |
|----------------|-----------------|
| `MANUAL` | `MANUAL` |
| `ATTI` | `STABILIZE` |
| `GPS_NORMAL` | `LOITER` |
| `GPS_SPORT` | `MANUAL` |
| `GO_HOME` | `RTL` |
| `AUTO_LANDING` | `LAND` |
| `WAYPOINT` | `AUTO` |
| `FOLLOW_ME` | `GUIDED` |

---

### 阶段 4: 网络传输层

**预期产出**: 能将数据通过 TCP/UDP 发送到 OUTB

| 序号 | 任务 | 优先级 |
|------|------|--------|
| 4.1 | 实现 TCP Client | P0 |
| 4.2 | 实现 UDP Client (备选) | P1 |
| 4.3 | 实现连接管理 (重连、心跳) | P0 |
| 4.4 | 实现消息队列与发送频率控制 | P0 |
| 4.5 | 网络状态监听与切换 | P1 |

**通信协议设计**:

```
消息格式: JSON over TCP
帧格式: [4字节长度][JSON数据]

连接流程:
1. Client 连接 Server
2. Client 发送 HELLO 消息 (含设备信息)
3. Server 返回 ACK
4. Client 持续发送 DroneState 消息
5. 每 30 秒发送 HEARTBEAT

消息类型:
- HELLO: {"type":"hello","device_id":"dji-xxx","sdk_version":"5.x"}
- STATE: {"type":"state","data":{...DroneState...}}
- HEARTBEAT: {"type":"heartbeat","timestamp":123456789}
```

---

### 阶段 5: OUTB DJI Adapter (Go)

**预期产出**: OUTB 支持接收 DJI 转发端数据

| 序号 | 任务 | 优先级 |
|------|------|--------|
| 5.1 | 创建 `internal/adapters/dji/` 目录 | P0 |
| 5.2 | 实现 TCP Server | P0 |
| 5.3 | 实现消息解析 | P0 |
| 5.4 | 实现多客户端管理 | P0 |
| 5.5 | 实现 Adapter 接口 | P0 |
| 5.6 | 配置文件支持 | P0 |
| 5.7 | 集成测试 | P0 |

**配置示例**:

```yaml
dji:
  enabled: true
  listen_address: "0.0.0.0:14560"
  protocol: tcp  # tcp | udp
  max_clients: 10
```

---

### 阶段 6: Android UI 与用户体验

**预期产出**: 完整可用的 Android 应用

| 序号 | 任务 | 优先级 |
|------|------|--------|
| 6.1 | 主界面设计 (连接状态、数据预览) | P0 |
| 6.2 | 设置界面 (服务器地址、发送频率) | P0 |
| 6.3 | 连接状态指示 (DJI、网络、服务器) | P0 |
| 6.4 | 后台服务 (Service) | P0 |
| 6.5 | 通知栏常驻 | P1 |
| 6.6 | 数据统计与日志 | P1 |

---

### 阶段 7: 测试与发布

| 序号 | 任务 | 优先级 |
|------|------|--------|
| 7.1 | 模拟器测试 (DJI Assistant) | P0 |
| 7.2 | 真机测试 (实际 DJI 飞机) | P0 |
| 7.3 | 网络异常测试 | P0 |
| 7.4 | 生成签名 APK | P0 |
| 7.5 | 编写用户文档 | P0 |

---

## 技术选型

### Android 应用

| 组件 | 技术 |
|------|------|
| 语言 | Kotlin |
| DJI SDK | Mobile SDK v5 |
| JSON | kotlinx.serialization |
| 网络 | OkHttp / Ktor Client |
| 依赖注入 | Hilt |
| UI | Jetpack Compose |
| 架构 | MVVM + Clean Architecture |

### OUTB Go 端

| 组件 | 技术 |
|------|------|
| TCP Server | net 标准库 |
| JSON | encoding/json |

---

## 风险与缓解

| 风险 | 概率 | 缓解措施 |
|------|------|----------|
| DJI SDK 注册需要企业资质 | 中 | 使用个人开发者账号，部分功能受限 |
| SDK 对特定机型不支持 | 中 | 先支持主流机型 (Mini/Air/Mavic 系列) |
| Android 后台服务被杀死 | 高 | 使用前台服务 + 通知栏常驻 |
| 网络不稳定 | 高 | 本地缓存 + 断线重连 + 消息队列 |

---

## 资源需求

| 资源 | 说明 |
|------|------|
| DJI 开发者账号 | 需注册 DJI Developer |
| Android 测试设备 | Android 7.0+ 手机/平板 |
| DJI 飞机 | 用于真机测试 (可选用模拟器) |
| DJI 遥控器 | 连接手机使用 |

---

## 时间估算

| 阶段 | 工作量 |
|------|--------|
| 阶段 1: 环境准备 | 1 天 |
| 阶段 2: 数据采集 | 2 天 |
| 阶段 3: 格式转换 | 1 天 |
| 阶段 4: 网络传输 | 2 天 |
| 阶段 5: Go Adapter | 1 天 |
| 阶段 6: Android UI | 2 天 |
| 阶段 7: 测试发布 | 2 天 |

---

## 目录结构规划

```
open-uav-telemetry-bridge/
├── cmd/outb/                    # 现有 Go 网关
├── internal/
│   ├── adapters/
│   │   ├── mavlink/             # 现有
│   │   └── dji/                 # 新增：DJI TCP Adapter
│   └── ...
├── android/                     # 新增：Android 项目
│   └── dji-forwarder/
│       ├── app/
│       │   └── src/main/
│       │       ├── java/
│       │       │   └── com/outb/dji/
│       │       │       ├── MainActivity.kt
│       │       │       ├── service/
│       │       │       │   └── ForwarderService.kt
│       │       │       ├── dji/
│       │       │       │   ├── DJIManager.kt
│       │       │       │   └── TelemetryCollector.kt
│       │       │       ├── network/
│       │       │       │   └── TcpClient.kt
│       │       │       └── model/
│       │       │           └── DroneState.kt
│       │       └── res/
│       ├── build.gradle.kts
│       └── settings.gradle.kts
└── docs/
    └── progress/
        └── PROGRESS_V0.2_DJI_20260118.md
```

---

## 下一步行动

**立即开始**: 阶段 5 - OUTB DJI Adapter (Go)

先在 Go 端实现 DJI TCP Adapter，便于后续 Android 开发时有服务端可调试。

```bash
# 创建 DJI Adapter 目录
mkdir -p internal/adapters/dji
```

---

## 修改记录

| 时间 | 修改内容 | 修改者 |
|------|----------|--------|
| 2026-01-18 17:10 | 创建 v0.2 开发计划 | Claude Code |
| 2026-01-18 17:30 | 完成 Go DJI Adapter (阶段 5) | Claude Code |
| 2026-01-18 18:00 | 完成 Android 项目框架 (阶段 1, 3, 4, 6) | Claude Code |
| 2026-01-18 18:30 | 完成 Android 完整实现，提交 Git | Claude Code |
| 2026-01-18 18:45 | 更新进度文档，准备端到端测试 | Claude Code |
