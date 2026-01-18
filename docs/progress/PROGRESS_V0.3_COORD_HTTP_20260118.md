# v0.3 坐标系转换 + HTTP API 开发计划

## 元信息

| 属性 | 值 |
|------|-----|
| 功能模块 | 坐标系转换、HTTP API |
| 创建时间 | 2026-01-18 17:15 |
| 最后更新 | 2026-01-18 17:25 |
| 作者 | Claude Code |
| 状态 | **核心实现完成** |
| 依赖版本 | v0.2.0 |
| Git Commit | 待提交 |

## 实现进展汇总

> v0.3 核心功能已完成：坐标系转换 + HTTP API

### 已完成 (2026-01-18 17:25)

| 功能 | 完成情况 | 备注 |
|------|----------|------|
| 坐标转换模块 | ✅ 完成 | `internal/core/coordinator/` |
| WGS84 → GCJ02 | ✅ 完成 | 中国区坐标转换 |
| WGS84 → BD09 | ✅ 完成 | 百度坐标转换 |
| HTTP API Server | ✅ 完成 | `internal/api/` |
| GET /api/v1/status | ✅ 完成 | 网关状态查询 |
| GET /api/v1/drones | ✅ 完成 | 全部无人机列表 |
| GET /api/v1/drones/{id} | ✅ 完成 | 单个无人机查询 |
| CORS 支持 | ✅ 完成 | 可配置跨域 |
| 单元测试 | ✅ 完成 | 坐标转换测试通过 |
| 端到端测试 | ✅ 完成 | 全链路验证通过 |

### 新增文件

```
internal/core/coordinator/
├── converter.go          # 坐标转换算法
└── converter_test.go     # 单元测试

internal/api/
└── server.go             # HTTP API 服务器
```

### 修改文件

- `internal/models/drone_state.go` - 添加 GCJ02/BD09 坐标字段
- `internal/config/config.go` - 添加 HTTP/Coordinate 配置
- `internal/core/engine.go` - 集成坐标转换器
- `cmd/outb/main.go` - 集成 HTTP 服务器
- `configs/config.example.yaml` - 添加新配置项

### 测试验证

**坐标转换验证** (天安门):
```
输入 WGS84:  39.9046, 116.4078
输出 GCJ02: 39.9060, 116.4140
偏移约: ~130m (符合预期)
```

**HTTP API 验证**:
```
GET /health             → 200 OK
GET /api/v1/status      → 200 {"version":"0.3.0-dev",...}
GET /api/v1/drones      → 200 {"count":1,"drones":[...]}
GET /api/v1/drones/{id} → 200 {完整 DroneState}
GET /api/v1/drones/xxx  → 404 {"error":"drone not found"}
```

---

## 功能一：坐标系转换

### 背景

中国地图服务（高德、百度、腾讯）使用加密坐标系，直接显示 GPS 原始坐标（WGS84）会产生偏移：

| 坐标系 | 使用场景 | 偏移 |
|--------|----------|------|
| **WGS84** | GPS 原始坐标、国际地图 | 基准 |
| **GCJ02** (火星坐标) | 高德、腾讯、Google中国 | ~100-700m |
| **BD09** | 百度地图 | GCJ02 基础上再偏移 |

### 设计方案

```
DroneState (WGS84 原始坐标)
    ↓
CoordinateConverter
    ↓
DroneState (含 WGS84 + GCJ02 + BD09)
```

### 数据模型扩展

```go
// internal/models/drone_state.go
type Location struct {
    // WGS84 原始坐标 (GPS)
    Lat         float64 `json:"lat"`
    Lon         float64 `json:"lon"`
    AltGNSS     float64 `json:"alt_gnss"`
    AltRelative float64 `json:"alt_relative"`

    // 转换后的坐标 (可选，按需填充)
    LatGCJ02    *float64 `json:"lat_gcj02,omitempty"`
    LonGCJ02    *float64 `json:"lon_gcj02,omitempty"`
    LatBD09     *float64 `json:"lat_bd09,omitempty"`
    LonBD09     *float64 `json:"lon_bd09,omitempty"`
}
```

### 转换算法

GCJ02 加密算法是公开的（虽然官方不承认），核心是对经纬度加偏移：

```go
// internal/core/coordinator/converter.go
package coordinator

import "math"

const (
    a  = 6378245.0              // 长半轴
    ee = 0.00669342162296594323 // 偏心率平方
)

// WGS84ToGCJ02 converts WGS84 coordinates to GCJ02 (Mars coordinates)
func WGS84ToGCJ02(lat, lon float64) (float64, float64) {
    if outOfChina(lat, lon) {
        return lat, lon // 国外不转换
    }

    dLat := transformLat(lon-105.0, lat-35.0)
    dLon := transformLon(lon-105.0, lat-35.0)

    radLat := lat / 180.0 * math.Pi
    magic := math.Sin(radLat)
    magic = 1 - ee*magic*magic
    sqrtMagic := math.Sqrt(magic)

    dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * math.Pi)
    dLon = (dLon * 180.0) / (a / sqrtMagic * math.Cos(radLat) * math.Pi)

    return lat + dLat, lon + dLon
}

// GCJ02ToBD09 converts GCJ02 to BD09 (Baidu coordinates)
func GCJ02ToBD09(lat, lon float64) (float64, float64) {
    z := math.Sqrt(lon*lon+lat*lat) + 0.00002*math.Sin(lat*math.Pi*3000.0/180.0)
    theta := math.Atan2(lat, lon) + 0.000003*math.Cos(lon*math.Pi*3000.0/180.0)
    return z*math.Sin(theta) + 0.006, z*math.Cos(theta) + 0.0065
}

// WGS84ToBD09 converts WGS84 directly to BD09
func WGS84ToBD09(lat, lon float64) (float64, float64) {
    gcjLat, gcjLon := WGS84ToGCJ02(lat, lon)
    return GCJ02ToBD09(gcjLat, gcjLon)
}

func outOfChina(lat, lon float64) bool {
    return lon < 72.004 || lon > 137.8347 || lat < 0.8293 || lat > 55.8271
}

func transformLat(x, y float64) float64 {
    ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
    ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
    ret += (20.0*math.Sin(y*math.Pi) + 40.0*math.Sin(y/3.0*math.Pi)) * 2.0 / 3.0
    ret += (160.0*math.Sin(y/12.0*math.Pi) + 320*math.Sin(y*math.Pi/30.0)) * 2.0 / 3.0
    return ret
}

func transformLon(x, y float64) float64 {
    ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
    ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
    ret += (20.0*math.Sin(x*math.Pi) + 40.0*math.Sin(x/3.0*math.Pi)) * 2.0 / 3.0
    ret += (150.0*math.Sin(x/12.0*math.Pi) + 300.0*math.Sin(x/30.0*math.Pi)) * 2.0 / 3.0
    return ret
}
```

### 配置选项

```yaml
# configs/config.yaml
coordinate:
  convert_gcj02: true   # 是否转换为 GCJ02
  convert_bd09: false   # 是否转换为 BD09
```

### 开发任务

| 序号 | 任务 | 优先级 |
|------|------|--------|
| 1.1 | 创建 `internal/core/coordinator/` 目录 | P0 |
| 1.2 | 实现 WGS84 → GCJ02 转换 | P0 |
| 1.3 | 实现 GCJ02 → BD09 转换 | P1 |
| 1.4 | 扩展 Location 数据模型 | P0 |
| 1.5 | 在 Engine 中集成转换器 | P0 |
| 1.6 | 添加配置开关 | P0 |
| 1.7 | 单元测试 (已知坐标点验证) | P0 |

---

## 功能二：HTTP API

### 设计目标

提供 RESTful HTTP API，便于：
- Web 前端查询无人机状态
- 第三方系统集成
- 调试和监控

### API 设计

**基础信息**:
- Base URL: `http://localhost:8080/api/v1`
- 响应格式: JSON
- 认证: 可选 (v0.3 暂不实现)

### 接口列表

#### 1. 获取所有无人机状态

```http
GET /api/v1/drones
```

响应:
```json
{
  "count": 2,
  "drones": [
    {
      "device_id": "mavlink-001",
      "timestamp": 1705579200000,
      "protocol_source": "mavlink",
      "location": {
        "lat": 39.9042,
        "lon": 116.4074,
        "lat_gcj02": 39.9087,
        "lon_gcj02": 116.4142,
        "alt_gnss": 100.5,
        "alt_relative": 50.2
      },
      "attitude": {
        "pitch": 2.5,
        "roll": -1.2,
        "yaw": 45.0
      },
      "status": {
        "armed": true,
        "flight_mode": "LOITER",
        "battery_percent": 85.0,
        "signal_quality": 95
      },
      "velocity": {
        "vx": 5.0,
        "vy": 3.0,
        "vz": -1.0
      }
    }
  ]
}
```

#### 2. 获取单个无人机状态

```http
GET /api/v1/drones/{device_id}
```

响应: 单个 DroneState 对象

错误响应 (404):
```json
{
  "error": "drone not found",
  "device_id": "unknown-001"
}
```

#### 3. 获取网关状态

```http
GET /api/v1/status
```

响应:
```json
{
  "version": "0.3.0",
  "uptime_seconds": 3600,
  "adapters": [
    {"name": "mavlink", "enabled": true, "connected": true},
    {"name": "dji", "enabled": true, "clients": 1}
  ],
  "publishers": [
    {"name": "mqtt", "enabled": true, "connected": true}
  ],
  "stats": {
    "messages_received": 10000,
    "messages_published": 5000,
    "active_drones": 2
  }
}
```

#### 4. 获取历史轨迹 (可选)

```http
GET /api/v1/drones/{device_id}/track?limit=100&since=1705579200000
```

响应:
```json
{
  "device_id": "mavlink-001",
  "points": [
    {"timestamp": 1705579200000, "lat": 39.9042, "lon": 116.4074, "alt": 100.5},
    {"timestamp": 1705579201000, "lat": 39.9043, "lon": 116.4075, "alt": 101.0}
  ]
}
```

### 技术选型

| 方案 | 优点 | 缺点 |
|------|------|------|
| **net/http** (标准库) | 无依赖，简单 | 路由不便 |
| **chi** | 轻量，兼容标准库 | 需额外依赖 |
| **gin** | 功能丰富，性能好 | 依赖较多 |
| **echo** | 性能好，API 清晰 | 依赖较多 |

**推荐**: `chi` - 轻量级路由器，与标准库兼容

```go
import "github.com/go-chi/chi/v5"
```

### 目录结构

```
internal/
├── api/
│   ├── server.go          # HTTP 服务器
│   ├── handlers/
│   │   ├── drones.go      # /drones 相关处理
│   │   └── status.go      # /status 相关处理
│   └── middleware/
│       └── logging.go     # 请求日志
└── core/
    └── statestore/        # 状态存储 (已有)
```

### 配置选项

```yaml
# configs/config.yaml
http:
  enabled: true
  address: "0.0.0.0:8080"
  cors_enabled: true
  cors_origins: ["*"]
```

### 开发任务

| 序号 | 任务 | 优先级 |
|------|------|--------|
| 2.1 | 添加 chi 依赖 | P0 |
| 2.2 | 创建 `internal/api/` 目录结构 | P0 |
| 2.3 | 实现 HTTP Server 启动/停止 | P0 |
| 2.4 | 实现 GET /api/v1/drones | P0 |
| 2.5 | 实现 GET /api/v1/drones/{id} | P0 |
| 2.6 | 实现 GET /api/v1/status | P0 |
| 2.7 | 添加请求日志中间件 | P1 |
| 2.8 | 添加 CORS 支持 | P1 |
| 2.9 | 实现轨迹查询 (可选) | P2 |
| 2.10 | 集成测试 | P0 |

---

## 系统集成

### 架构更新

```
南向适配层
├── MAVLink Adapter
└── DJI Adapter
    ↓ DroneState (WGS84)
核心处理层
├── Engine
├── CoordinateConverter (新增)  ← 坐标转换
├── Throttler
└── StateStore
    ↓ DroneState (WGS84 + GCJ02 + BD09)
北向发布层
├── MQTT Publisher
└── HTTP API (新增)  ← REST 接口
```

### 数据流

1. Adapter 产生 DroneState (WGS84 原始坐标)
2. Engine 传递给 CoordinateConverter
3. Converter 添加 GCJ02/BD09 坐标
4. Throttler 频率控制
5. StateStore 存储最新状态
6. MQTT Publisher 发布到 Broker
7. HTTP API 从 StateStore 读取响应请求

---

## 开发阶段划分

### 阶段 1: 坐标转换模块

**预期产出**: 可正确转换坐标的独立模块

| 步骤 | 内容 |
|------|------|
| 1 | 创建 coordinator 包 |
| 2 | 实现转换算法 |
| 3 | 单元测试验证 |
| 4 | 集成到 Engine |

### 阶段 2: HTTP API 基础

**预期产出**: 可查询无人机状态的 HTTP 服务

| 步骤 | 内容 |
|------|------|
| 1 | 添加 chi 路由器 |
| 2 | 实现 Server 结构 |
| 3 | 实现 drones 接口 |
| 4 | 实现 status 接口 |

### 阶段 3: 集成测试

**预期产出**: 端到端验证

| 步骤 | 内容 |
|------|------|
| 1 | 启动网关 |
| 2 | 发送模拟数据 |
| 3 | 验证 HTTP 响应坐标正确 |
| 4 | 验证 MQTT 消息坐标正确 |

---

## 测试数据

### 坐标转换验证点

| 地点 | WGS84 | GCJ02 (预期) |
|------|-------|--------------|
| 天安门 | 39.9087, 116.3975 | 39.9111, 116.4037 |
| 东方明珠 | 31.2397, 121.4998 | 31.2378, 121.5058 |
| 广州塔 | 23.1066, 113.3245 | 23.1044, 113.3312 |

---

## 风险与缓解

| 风险 | 概率 | 缓解措施 |
|------|------|----------|
| 坐标转换精度问题 | 低 | 使用已验证的算法，对照在线工具验证 |
| HTTP 并发性能 | 中 | chi 性能足够，必要时可加缓存 |
| StateStore 并发访问 | 中 | 已有 RWMutex，需确保读写安全 |

---

## 下一步行动

**立即开始**: 阶段 1 - 坐标转换模块

```bash
# 创建 coordinator 目录
mkdir -p internal/core/coordinator
```

---

## 修改记录

| 时间 | 修改内容 | 修改者 |
|------|----------|--------|
| 2026-01-18 17:15 | 创建 v0.3 开发计划 | Claude Code |
