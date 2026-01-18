# v0.1 MVP 开发计划

## 元信息

| 属性 | 值 |
|------|-----|
| 功能模块 | v0.1 MVP 全模块 |
| 创建时间 | 2026-01-18 16:25 |
| 最后更新 | 2026-01-18 16:50 |
| 作者 | Claude Code |
| 状态 | 进行中 |

## 目标

**v0.1 MVP**: MAVLink 输入 → JSON/MQTT 输出，在树莓派上可运行

## 当前进展摘要

阶段 1-7 核心功能已完成。项目可编译运行，支持 MAVLink 数据接收和 MQTT 发布。

## 开发阶段划分

### 阶段 1: 项目基础设施 (Foundation) ✅ 已完成

| 序号 | 任务 | 状态 |
|------|------|------|
| 1.1 | 初始化 Go 模块 | ✅ 已完成 |
| 1.2 | 创建目录结构 | ✅ 已完成 |
| 1.3 | 添加 .gitignore | ✅ 已完成 |
| 1.4 | 添加基础依赖 | ✅ 已完成 |
| 1.5 | 创建 Makefile | ✅ 已完成 |
| 1.6 | 创建 ARM64 交叉编译脚本 | ✅ 已完成 |

---

### 阶段 2: 统一数据模型 (Data Model) ✅ 已完成

| 序号 | 任务 | 状态 |
|------|------|------|
| 2.1-2.5 | DroneState 及子结构体 | ✅ 已完成 |
| 2.6 | JSON 序列化 | ✅ 已完成 |
| 2.7 | FlightMode 枚举 | ✅ 已完成 |
| 2.8 | 单元测试 | ✅ 已完成 (3 tests) |

---

### 阶段 3: 配置管理 (Configuration) ✅ 已完成

| 序号 | 任务 | 状态 |
|------|------|------|
| 3.1-3.4 | Config 结构体与 YAML 加载 | ✅ 已完成 |
| 3.5 | 单元测试 | ✅ 已完成 (4 tests) |

---

### 阶段 4: 核心处理层 (Core Kernel) ✅ 已完成

| 序号 | 任务 | 状态 |
|------|------|------|
| 4.1 | StateStore (内存缓存) | ✅ 已完成 |
| 4.3 | Throttler (频率控制) | ✅ 已完成 |
| 4.4 | Adapter 接口 | ✅ 已完成 |
| 4.5 | Publisher 接口 | ✅ 已完成 |
| 4.6 | Core Engine (消息路由) | ✅ 已完成 |
| 4.7 | 单元测试 | ✅ 已完成 (7 tests) |

---

### 阶段 5: MAVLink 适配器 (Southbound) ✅ 已完成

| 序号 | 任务 | 状态 |
|------|------|------|
| 5.1 | 集成 gomavlib v3.3.0 | ✅ 已完成 |
| 5.2 | UDP/TCP/Serial 连接管理 | ✅ 已完成 |
| 5.3 | 解析 HEARTBEAT | ✅ 已完成 |
| 5.4 | 解析 GLOBAL_POSITION_INT | ✅ 已完成 |
| 5.5 | 解析 ATTITUDE | ✅ 已完成 |
| 5.6 | 解析 SYS_STATUS | ✅ 已完成 |
| 5.7 | FlightMode 映射 (Copter/Plane) | ✅ 已完成 |
| 5.8 | 实现 Adapter 接口 | ✅ 已完成 |

---

### 阶段 6: MQTT 发布器 (Northbound) ✅ 已完成

| 序号 | 任务 | 状态 |
|------|------|------|
| 6.1 | 集成 paho.mqtt.golang v1.5.1 | ✅ 已完成 |
| 6.2 | 连接管理 (含自动重连) | ✅ 已完成 |
| 6.3 | LWT 遗嘱机制 | ✅ 已完成 |
| 6.4 | 消息发布 (JSON 格式) | ✅ 已完成 |
| 6.5 | 实现 Publisher 接口 | ✅ 已完成 |

---

### 阶段 7: 主程序集成 (Integration) ✅ 已完成

| 序号 | 任务 | 状态 |
|------|------|------|
| 7.1 | main.go 完整启动流程 | ✅ 已完成 |
| 7.2 | 优雅关闭 (signal handling) | ✅ 已完成 |
| 7.3 | 启动日志 | ✅ 已完成 |
| 7.5 | ARM64 交叉编译 | ✅ 已完成 |

---

### 阶段 8: 文档与发布 (Release) ⏳ 待开始

| 序号 | 任务 | 状态 |
|------|------|------|
| 8.1 | README 快速开始指南 | ⏳ 待开始 |
| 8.2 | 配置说明文档 | ⏳ 待开始 |
| 8.3 | release/v0.1/ 发布包 | ⏳ 待开始 |
| 8.4 | CHANGELOG | ⏳ 待开始 |
| 8.5 | Git tag v0.1.0 | ⏳ 待开始 |

---

## 项目文件结构

```
open-uav-telemetry-bridge/
├── cmd/outb/main.go                          # 程序入口
├── internal/
│   ├── models/
│   │   ├── drone_state.go                    # DroneState 数据模型
│   │   └── drone_state_test.go
│   ├── core/
│   │   ├── interfaces.go                     # Adapter/Publisher 接口
│   │   ├── engine.go                         # 消息路由引擎
│   │   ├── statestore/statestore.go          # 状态缓存
│   │   └── throttler/throttler.go            # 频率控制
│   ├── adapters/mavlink/
│   │   ├── adapter.go                        # MAVLink 适配器
│   │   └── flightmode.go                     # FlightMode 映射
│   ├── publishers/mqtt/
│   │   └── publisher.go                      # MQTT 发布器
│   └── config/
│       ├── config.go                         # 配置加载
│       └── config_test.go
├── configs/config.example.yaml
├── scripts/build-arm64.sh
├── bin/
│   ├── outb                                  # macOS 二进制 (7.4M)
│   └── outb-linux-arm64                      # 树莓派二进制 (7.2M)
├── go.mod
├── go.sum
└── Makefile
```

## 构建结果

```bash
$ make build
Building outb...
# bin/outb (7.4M)

$ make build-linux-arm64
Building outb for Linux ARM64...
# bin/outb-linux-arm64 (7.2M)
```

## 测试结果

```
=== 测试执行时间: 2026-01-18 16:50 ===

ok  github.com/open-uav/telemetry-bridge/internal/config        (4 tests)
ok  github.com/open-uav/telemetry-bridge/internal/core/statestore (2 tests)
ok  github.com/open-uav/telemetry-bridge/internal/core/throttler  (5 tests)
ok  github.com/open-uav/telemetry-bridge/internal/models          (3 tests)

总计: 14 tests, 全部通过
```

## 依赖版本

| 依赖 | 版本 |
|------|------|
| Go | 1.25.1 |
| gomavlib | v3.3.0 |
| paho.mqtt.golang | v1.5.1 |
| yaml.v3 | v3.0.1 |

## 下一步行动

**待完成**: 阶段 8 - 文档与发布

1. 编写 README 快速开始指南
2. 创建 release/v0.1/ 发布包
3. 在真实环境测试 (SITL + Mosquitto)

## 修改记录

| 时间 | 修改内容 | 修改者 |
|------|----------|--------|
| 2026-01-18 16:25 | 创建 v0.1 开发计划 | Claude Code |
| 2026-01-18 16:45 | 完成阶段 1-4 | Claude Code |
| 2026-01-18 16:50 | 完成阶段 5-7，核心功能实现 | Claude Code |
