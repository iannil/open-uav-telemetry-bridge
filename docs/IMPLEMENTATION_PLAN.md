# IMPLEMENTATION_PLAN.md

Open-UAV-Telemetry-Bridge 实施计划

## 元信息

| 属性 | 值 |
|------|-----|
| 创建时间 | 2026-01-18 16:15 |
| 最后更新 | 2026-01-18 16:25 |
| 当前版本 | v0.0 (规划阶段) |
| 目标版本 | v1.0 |

## 项目总览

### 项目目标

构建一个协议无关的无人机遥测边缘网关，实现：
- 多协议输入（MAVLink、DJI、GB/T 28181）
- 统一数据模型转换
- 多格式输出（MQTT、WebSocket、HTTP、gRPC）

### 当前状态

**阶段**: 预发布规划阶段 (Pre-alpha)

| 组件 | 状态 | 说明 |
|------|------|------|
| 架构设计 | ✅ 已完成 | 见 README.md |
| 项目指南 | ✅ 已完成 | 见 CLAUDE.md |
| 文档规范 | ✅ 已完成 | 见 /docs/standards/ |
| Go 项目初始化 | ⏳ 待开始 | go.mod, 目录结构 |
| 核心模块实现 | ⏳ 待开始 | - |
| 适配器实现 | ⏳ 待开始 | - |
| 发布器实现 | ⏳ 待开始 | - |

## 版本路线图

### v0.1 - MVP (最小可行产品)

**目标**: MAVLink 输入 → JSON/MQTT 输出，在树莓派上可运行

#### 核心任务

1. **Go 项目初始化**
   - [ ] 创建 go.mod
   - [ ] 建立目录结构 (src/core, src/adapters, src/publishers, src/models, src/config)
   - [ ] 配置基础依赖

2. **统一数据模型**
   - [ ] 实现 DroneState 结构体
   - [ ] 实现 JSON 序列化/反序列化
   - [ ] 单元测试

3. **核心处理层**
   - [ ] Validator: 数据校验模块
   - [ ] StateStore: 状态缓存模块
   - [ ] Throttler: 基础频率控制

4. **MAVLink 适配器**
   - [ ] 集成 gomavlib
   - [ ] 解析 HEARTBEAT, GLOBAL_POSITION_INT, SYS_STATUS
   - [ ] 映射到 DroneState

5. **MQTT 发布器**
   - [ ] 集成 paho.mqtt.golang
   - [ ] 实现基础发布功能
   - [ ] 实现 LWT (遗嘱机制)

6. **配置管理**
   - [ ] YAML 配置解析
   - [ ] 基础配置项定义

7. **构建与部署**
   - [ ] ARM64 交叉编译脚本
   - [ ] 树莓派运行验证

### v0.2 - DJI 支持

- [ ] Android 转发端 Demo
- [ ] Sidecar 进程通信机制

### v0.3 - 坐标系与 HTTP API

- [ ] WGS84 → GCJ02/BD09 坐标转换
- [ ] HTTP REST API 查询接口

### v0.4 - 国标支持

- [ ] GB/T 28181 A模式实现
- [ ] SIP/RTP 协议解析

### v1.0 - 正式发布

- [ ] Web 管理界面
- [ ] 可视化配置
- [ ] 完整文档

## 技术债务清单

当前无技术债务（项目尚未开始实现）

## 风险与依赖

| 风险项 | 影响程度 | 缓解措施 |
|--------|----------|----------|
| gomavlib 兼容性 | 中 | 提前验证 MAVLink 方言支持 |
| DJI SDK 闭源限制 | 高 | 采用 Sidecar 模式隔离 |
| GB/T 28181 标准复杂度 | 高 | 分阶段实现，优先 A 模式 |

## 相关文档

- [README.md](/README.md) - 完整技术规格说明
- [CLAUDE.md](/CLAUDE.md) - 项目开发指南
- [文档命名规范](/docs/standards/NAMING_CONVENTION.md)
- [v0.1 MVP 开发计划](/docs/progress/PROGRESS_MVP_V0.1_20260118.md) - 详细任务分解
