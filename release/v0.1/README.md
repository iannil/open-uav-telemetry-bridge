# Open-UAV-Telemetry-Bridge v0.1.0

## 发布内容

| 文件 | 描述 | 平台 |
|------|------|------|
| `outb-linux-arm64` | 预编译二进制 | Linux ARM64 (树莓派/Jetson) |
| `config.example.yaml` | 示例配置文件 | - |

## 快速开始

```bash
# 1. 下载并赋予执行权限
chmod +x outb-linux-arm64

# 2. 复制并编辑配置
cp config.example.yaml config.yaml
nano config.yaml

# 3. 运行
./outb-linux-arm64 config.yaml
```

## 系统要求

- Linux ARM64 (树莓派 4、Jetson Nano 等)
- MQTT Broker (Mosquitto 等)
- MAVLink 数据源

## 配置说明

```yaml
mavlink:
  enabled: true
  connection_type: udp       # udp | tcp | serial
  address: "0.0.0.0:14550"   # UDP 监听地址
  # serial_port: "/dev/ttyUSB0"
  # serial_baud: 57600

mqtt:
  enabled: true
  broker: "tcp://localhost:1883"
  client_id: "outb-001"
  topic_prefix: "uav/telemetry"

throttle:
  default_rate_hz: 1.0       # 发布频率
```

## MQTT 输出格式

Topic: `uav/telemetry/{device_id}/state`

```json
{
  "device_id": "mavlink-1",
  "timestamp": 1705596600000,
  "protocol_source": "mavlink",
  "location": {
    "lat": 22.5431,
    "lon": 114.0579,
    "alt_baro": 120.5,
    "alt_gnss": 125.0,
    "coordinate_system": "WGS84"
  },
  "attitude": {
    "roll": 0.05,
    "pitch": -0.12,
    "yaw": 180.0
  },
  "status": {
    "battery_percent": 85,
    "flight_mode": "AUTO",
    "armed": true,
    "signal_quality": 0
  },
  "velocity": {
    "vx": 10.5,
    "vy": 0.0,
    "vz": -0.5
  }
}
```

## 更多信息

- [完整文档](../../docs/QUICKSTART.md)
- [项目主页](../../README.md)
- [更新日志](../../CHANGELOG.md)
