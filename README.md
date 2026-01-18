# OPEN-UAV-TELEMETRY-BRIDGE

这是一个非常扎实且极具商业价值的技术选题。Open-UAV-Telemetry-Bridge（以下简称 OUTB）的核心使命是“屏蔽底层差异，提供统一语言”。

它不仅仅是一个简单的协议转换器，更应该是低空物联网（IoD, Internet of Drones）的边缘计算网关。

以下是为 OUTB 设计的完整技术方案：

---

## 1. 核心架构设计 (Architecture)

采用 微内核 + 插件化 (Microkernel + Plugin) 架构。核心层负责数据路由、缓存和监控，南北向接口通过插件扩展。

### 架构分层图：

```text
+---------------------------------------------------------------+
|                 北向接口层 (Northbound Interfaces)             |
|  [MQTT Publisher]  [WebSocket Server]  [HTTP Webhooks]  [gRPC]|
+---------------------------+-----------------------------------+
                            ^
                            | (Standardized JSON/Protobuf)
+---------------------------+-----------------------------------+
|                    核心处理层 (Core Kernel)                    |
|---------------------------------------------------------------|
|  1. 数据清洗 (Validator)   |  2. 坐标系转换 (WGS84/GCJ02)     |
|  3. 频率控制 (Throttling)  |  4. 状态缓存 (State Store)       |
+---------------------------+-----------------------------------+
                            ^
                            | (Raw Data Event)
+---------------------------+-----------------------------------+
|                 南向适配层 (Southbound Adapters)               |
| [MAVLink Driver] [DJI SDK Bridge] [GB/T 28181] [Private TCP]  |
+---------------------------------------------------------------+
       ^                   ^                   ^
       | (Serial/UDP)      | (SDK Callback)    | (GB Stream)
   [PX4/ArduPilot]     [DJI Drone]       [Gov Platform]
```

---

## 2. 统一数据模型 (Unified Data Model)

这是本项目的灵魂。我们需要定义一套“低空通用语言”（JSON Schema）。无论底层是大疆还是开源飞控，输出给上层的数据结构必须一致。

设计示例 (DroneState.json):

```json
{
  "device_id": "uav-sn-12345678",        // 唯一设备码
  "timestamp": 1709882231000,            // 毫秒级时间戳
  "protocol_source": "mavlink",          // 数据来源
  "location": {
    "lat": 22.5431,
    "lon": 114.0579,
    "alt_baro": 120.5,                   // 气压高度
    "alt_gnss": 125.0,                   // GPS高度
    "coordinate_system": "WGS84"         // 显式声明坐标系，解决地图偏移痛点
  },
  "attitude": {
    "roll": 0.05,
    "pitch": -0.12,
    "yaw": 180.0
  },
  "status": {
    "battery_percent": 85,
    "flight_mode": "AUTO",               // 统一映射：如将 "LOITER" 和 "P-Mode" 都映射为 "HOVER"
    "armed": true,
    "signal_quality": 95
  },
  "velocity": {
    "vx": 10.5,
    "vy": 0.0,
    "vz": -0.5
  }
}
```

---

## 3. 关键模块技术选型

考虑到性能、并发能力和跨平台部署（既要跑在云端服务器，也要跑在树莓派/Jetson上），Golang 是最佳选择。

| 模块 | 技术选型 | 理由 |
| :--- | :--- | :--- |
| 开发语言 | Golang | 天然高并发，跨平台编译（Linux/Arm/Windows），内存安全，适合网络编程。 |
| MAVLink解析 | `gomavlib` | 现成熟的 Go 语言 MAVLink 库，支持生成特定方言代码。 |
| DJI 适配 | CGO / Sidecar | DJI OSDK 是 C++ 的，可以通过 CGO 封装，或者做一个轻量级 C++ Sidecar 进程通过 IPC 与 Go 主进程通信（推荐 Sidecar 模式，避免 C++ 崩溃带崩主程序）。 |
| 国标 GB/T | 自研解析器 | GB/T 28181/35017 多基于 SIP/RTP，Go 处理网络流非常高效。 |
| 消息总线 | Go Channel | 内部高性能异步通信。 |
| 北向输出 | `paho.mqtt.golang` | 标准 MQTT 协议支持。 |
| 配置管理 | YAML/TOML | 易读易改，方便现场部署。 |

---

## 4. 核心功能点详细设计

### A. 坐标系自动纠偏 (The "Coordinate" Pain)

* 痛点： 飞控一般用 WGS84（GPS坐标），但国内政府地图、高德地图用 GCJ02（火星坐标），百度用 BD09。直接上图会飞到河里去。
* 方案： 内置轻量级坐标转换库。
* 配置： 用户在配置文件中设置 `output_coordinate: "GCJ02"`，网关层自动计算转换，上层应用拿到的直接是可上图的坐标。

### B. 频率“削峰填谷” (Frequency Throttling)

* 痛点： 飞控通过串口吐出的姿态数据可能高达 50Hz-100Hz，如果直接透传到 4G 网络推给云端，流量费爆炸且云端处理不过来。
* 方案： 实现一个降采样（Downsampling）过滤器。
  * 设置 `report_rate: 1Hz`（每秒一包）。
  * 网关缓存最新状态，按 1秒 时间窗快照发送。
  * *高级功能：* 动态频率。平稳飞行时 0.5Hz，姿态剧烈变化或报警时自动提升到 10Hz。

### C. 断连与“遗嘱”机制 (Connection & Will)

* 痛点： 飞机飞丢了，云端还在显示“在线”，导致误判。
* 方案： 利用 MQTT 的 Last Will and Testament (LWT) 机制。
  * OUTB 启动时向 MQTT Broker 注册遗嘱消息：“Device Offline”。
  * 一旦 OUTB 崩溃或网络断开，Broker 自动推送离线消息给监控大屏。

### D. 多协议插件化适配 (The Adapter Pattern)

1. MAVLink (PX4/ArduPilot):

* 监听 UDP/TCP 或 Serial 端口。
* 解析 `HEARTBEAT`, `GLOBAL_POSITION_INT`, `SYS_STATUS` 消息。
* 映射到统一 Data Model。

2. DJI (大疆系列):

* 难点： 大疆是私有协议。
* 方案： 提供一个 Android APP 桥接器（运行在遥控器上的手机）或 Onboard SDK 桥接器（运行在机载电脑）。
* 这些桥接器作为 OUTB 的客户端，主动将大疆数据推送到 OUTB 服务端。

3. 国标 (GB/T 28181 - 2016/2022):

* 模拟 SIP Server 行为，接收无人机或机巢注册。
* 解析 XML 格式的设备状态通知。

---

## 5. 部署场景 (Deployment Scenarios)

场景一：机载边缘网关 (Onboard Edge)

* 硬件： Raspberry Pi 4 / NVIDIA Jetson / Rockchip开发板。
* 部署： OUTB 运行在板子上，通过串口直连飞控。
* 优势： 延时最低，支持断网续传。

场景二：地面站网关 (GCS Gateway)

* 硬件： 笔记本电脑 / 安卓平板。
* 部署： OUTB 运行在地面站电脑后台，读取地面站软件转发的数据。
* 优势： 适配存量老旧无人机（不用改飞机硬件）。

场景三：云端协议转换器 (Cloud Aggregator)

* 硬件： 阿里云/腾讯云服务器。
* 部署： 接收来自数千架无人机的原始 TCP/UDP 裸流，集中转码。
* 优势： 适合对接已经建好的私有协议机巢网络。

---

## 6. 开源路线图 (Roadmap)

* v0.1 (MVP): 仅支持 MAVLink 输入，输出标准 JSON over MQTT。在树莓派上跑通。
* v0.2: 增加 DJI Mobile SDK 的 Android 转发端 Demo。
* v0.3: 增加坐标系转换功能，增加 HTTP API 查询当前状态。
* v0.4: 支持 GB/T 28181 A模式（级联），模拟下级平台向上级政府平台推送数据。（杀手级特性）
* v1.0: 发布 Web 管理界面，支持可视化配置。

---

## 7. 商业化延伸 (Business Model)

虽然软件是开源的，但商业模式很清晰：

1. 卖盒子（软硬一体）： 将 OUTB 刷入工业级 4G/5G 模块，作为“低空黑匣子”卖给无人机厂商。
2. 卖企业版（SaaS/私有化）： 开源版只管数据转发，企业版提供历史轨迹存储、多机编队管理、权限控制后台。
3. 技术咨询： 帮政府做“全市无人机统一接入标准”制定。

这个方案技术栈主流（Go），架构清晰（微内核），且直击“协议不通”的行业最痛点，非常适合作为开源切入点。
