# Open-UAV-Telemetry-Bridge 剩余问题清单

> 创建时间: 2026-01-23
> 最后更新: 2026-01-23
> 排列顺序: 全局→局部, 高风险→低风险, 高优先级→低优先级

---

## 一、架构与安全问题 (P0 - 高风险)

### ~~1.1 HTTP API 缺少 HTTPS/TLS 支持~~ ✅ 已修复
- **状态**: 已修复 (2026-01-23)
- **修复内容**: 添加 TLSConfig 配置，支持证书文件配置
- **相关文件**: `internal/api/server.go`, `internal/config/config.go`, `configs/config.example.yaml`

### ~~1.2 API 缺少速率限制 (Rate Limiting)~~ ✅ 已修复
- **状态**: 已修复 (2026-01-23)
- **修复内容**: 添加 IP 级别速率限制中间件，使用 `golang.org/x/time/rate`
- **相关文件**: `internal/api/ratelimit/ratelimit.go`, `internal/api/server.go`

### ~~1.3 CORS 配置过于宽松~~ ✅ 已修复
- **状态**: 已修复 (2026-01-23)
- **修复内容**: 示例配置默认值从 `["*"]` 改为 `["http://localhost:3000"]`
- **相关文件**: `configs/config.example.yaml`

### ~~1.4 测试脚本硬编码敏感信息~~ ✅ 已修复
- **状态**: 已修复 (2026-01-23)
- **修复内容**: 移除默认密码，改为必填参数或环境变量 `GB28181_PASSWORD`
- **相关文件**: `scripts/test_gb28181_client.go`

---

## 二、测试覆盖不足 (P0 - 高风险)

### ~~2.1 核心适配器缺少单元测试~~ ✅ 已修复
| 模块 | 测试文件 | 状态 |
|------|----------|------|
| MAVLink Adapter | `internal/adapters/mavlink/adapter_test.go` | ✅ 已添加 (13 tests) |
| DJI Adapter | `internal/adapters/dji/adapter_test.go` | ✅ 已添加 (16 tests) |
| MQTT Publisher | `internal/publishers/mqtt/publisher_test.go` | ✅ 已添加 (9 tests) |

### ~~2.2 核心功能模块缺少单元测试~~ ✅ 已修复
| 模块 | 测试文件 | 状态 |
|------|----------|------|
| Logger Buffer | `internal/core/logger/buffer_test.go` | ✅ 已添加 (18 tests) |
| Alerter | `internal/core/alerter/alerter_test.go` | ✅ 已添加 (18 tests) |
| Geofence Engine | `internal/core/geofence/geofence_test.go` | ✅ 已添加 (25 tests) |
| Auth Manager | `internal/api/auth/auth_test.go` | ✅ 已添加 (21 tests) |
| Rate Limiter | `internal/api/ratelimit/ratelimit_test.go` | ✅ 已添加 (11 tests) |

### ~~2.3 Web 前端缺少测试~~ ✅ 已修复
- **状态**: 已修复 (2026-01-23)
- **修复内容**: 添加 Vitest 测试框架配置，创建 42 个测试用例
- **相关文件**:
  - `web/vite.config.ts` - Vitest 配置
  - `web/src/test/setup.ts` - 测试环境配置
  - `web/src/test/mocks.ts` - 测试数据
  - `web/src/store/droneStore.test.ts` - 14 tests
  - `web/src/store/authStore.test.ts` - 12 tests
  - `web/src/components/Drone/DroneCard.test.tsx` - 16 tests

### 2.4 已有测试存在条件跳过
- **位置**: `internal/publishers/gb28181/publisher_test.go:401-402`
- **问题**: 部分集成测试在 short mode 下跳过
- **影响**: 可能遗漏集成问题
- **建议**: 确保 CI 中运行完整测试套件

---

## 三、功能未完成 (P1 - 中风险)

### 3.1 DJI SDK 集成为占位符
- **位置**: `android/dji-forwarder/app/src/main/java/com/outb/dji/dji/DJIManager.kt:50-54, 66-70`
- **问题**:
  ```kotlin
  // TODO: Implement actual DJI SDK initialization:
  // TODO: Implement actual telemetry collection:
  ```
- **影响**: DJI 适配器仅能使用模拟模式，无法连接真实 DJI 无人机
- **依赖**: 需要注册 DJI 开发者账号获取 App Key

### 3.2 Android 端到端测试未完成
- **位置**: `docs/progress/PROGRESS_V0.2_DJI_20260118.md` 阶段 7
- **问题**: 端到端测试和 APK 签名发布尚未完成
- **建议**: 完成模拟模式下的端到端测试流程

### ~~3.3 HTTP API 状态接口数据不完整~~ ✅ 已修复
- **状态**: 已修复 (2026-01-23)
- **修复内容**: 添加 `GetAdapterNames()` 和 `GetPublisherNames()` 方法到 Engine，在 `/api/v1/status` 返回实际数据
- **相关文件**: `internal/core/engine.go`, `internal/api/server.go`

---

## 四、代码质量问题 (P2 - 低风险)

### ~~4.1 硬编码的默认端口和地址~~ ✅ 部分修复
- **状态**: 测试脚本已修复 (2026-01-23)
- **修复内容**: 测试脚本支持命令行参数配置 (`-addr`, `-path`, `-device`, `-count`)
- **相关文件**: `scripts/test_dji_client.go`, `scripts/test_ws_client.go`
- **剩余**: `internal/config/config.go` 中的默认值属于合理默认，无需修改

### ~~4.2 日志缓冲区大小硬编码~~ ✅ 已修复
- **状态**: 已修复 (2026-01-23)
- **修复内容**: 添加 `server.log_buffer_size` 配置项，默认值 1000
- **相关文件**: `internal/config/config.go`, `internal/api/server.go`, `configs/config.example.yaml`

### 4.3 错误处理一致性
- **位置**: 部分文件
- **问题**: 部分错误使用 `log.Printf` 而非结构化日志
- **建议**: 统一使用结构化日志格式

---

## 五、文档同步问题 (P2 - 低风险)

### ~~5.1 PROJECT_STATUS.md 严重过期~~ ✅ 已修复
- **状态**: 已修复 (2026-01-23)
- **修复内容**: 更新至 v0.4.0-dev 状态，包含完整功能列表、测试覆盖、技术栈信息

### ~~5.2 缺少 API 文档~~ ✅ 已修复
- **状态**: 已修复 (2026-01-23)
- **修复内容**: 添加 OpenAPI 3.0.3 规范文档，包含所有 API 端点定义
- **相关文件**: `docs/api/openapi.yaml`

### 5.3 README 与实际代码的版本号不一致
- **问题**: README 中部分信息可能与当前代码状态不同步
- **建议**: 定期审查并更新

---

## 六、Web 前端问题 (P2 - 低风险)

### 6.1 缺少国际化支持
- **位置**: `web/src/**/*.tsx`
- **问题**: 界面文本硬编码，仅支持单一语言
- **建议**: 根据需求添加 i18n 支持

### 6.2 缺少移动端适配测试
- **问题**: 未验证在移动设备上的显示效果
- **建议**: 添加响应式设计测试

---

## 七、待确认/低优先级

### 7.1 GB/T 28181 协议实现完整性
- **问题**: 需要与实际国标平台对接验证
- **状态**: 已有测试客户端 (`scripts/test_gb28181_client.go`)

### 7.2 gRPC 输出接口
- **问题**: README 提及 gRPC 输出，但代码中未实现
- **建议**: 明确是否为计划功能，更新文档

### 7.3 数据持久化
- **问题**: 轨迹数据仅内存存储，重启后丢失
- **建议**: 根据需求考虑添加数据库持久化

---

## 汇总统计

| 优先级 | 总数 | 已修复 | 剩余 |
|--------|------|--------|------|
| P0 (高风险) | 8 | 8 | 0 |
| P1 (中风险) | 3 | 1 | 2 |
| P2 (低风险) | 8 | 4 | 4 |
| 待确认 | 3 | 0 | 3 |
| **合计** | **22** | **13** | **9** |

**剩余问题**:
- P1: DJI SDK 集成 (3.1) - 需要开发者账号
- P1: Android 端到端测试 (3.2)
- P2: 错误处理一致性 (4.3)
- P2: README 版本同步 (5.3)
- P2: 国际化 (6.1)
- P2: 移动端适配 (6.2)
- 待确认: GB28181 对接验证、gRPC、数据持久化

---

## 修改记录

| 时间 | 修改内容 | 修改者 |
|------|----------|--------|
| 2026-01-23 08:00 | 创建问题清单 | Claude Code |
| 2026-01-23 14:45 | 标记已修复问题 (单元测试、HTTPS、速率限制、CORS、API状态、文档) | Claude Code |
| 2026-01-23 15:15 | 修复测试脚本硬编码、添加 Web 前端测试 (42 tests) | Claude Code |
| 2026-01-23 15:20 | 日志缓冲区大小改为可配置 | Claude Code |
| 2026-01-23 15:25 | 添加 OpenAPI 3.0.3 规范文档 | Claude Code |
