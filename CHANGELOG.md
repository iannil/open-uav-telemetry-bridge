# Changelog

本文件记录 Open-UAV-Telemetry-Bridge 的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [0.1.0] - 2026-01-18

### 新增

- **MAVLink 适配器**: 支持 UDP/TCP/Serial 连接，解析 HEARTBEAT、GLOBAL_POSITION_INT、ATTITUDE、SYS_STATUS 消息
- **MQTT 发布器**: 支持连接管理、自动重连、LWT 遗嘱机制
- **统一数据模型**: DroneState JSON 结构，包含位置、姿态、状态、速度信息
- **频率控制**: 可配置的发布频率 (0.5-10 Hz)
- **状态缓存**: 内存缓存最新设备状态
- **FlightMode 映射**: ArduCopter/ArduPlane 飞行模式到统一枚举的映射
- **配置管理**: YAML 配置文件支持
- **交叉编译**: 支持 Linux ARM64 (树莓派/Jetson)

### 技术细节

- Go 1.25.1
- gomavlib v3.3.0
- paho.mqtt.golang v1.5.1

### 已知限制

- 仅支持 MAVLink 协议输入
- 仅支持 MQTT 协议输出
- 未实现坐标系转换 (WGS84 → GCJ02)
- 未实现动态频率调整

---

## [未发布]

### 计划中

- v0.2: DJI Mobile SDK Android 转发端
- v0.3: 坐标系转换 + HTTP API
- v0.4: GB/T 28181 国标支持
- v1.0: Web 管理界面
