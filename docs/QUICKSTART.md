# 快速开始指南

## 系统要求

- Go 1.21+ (仅编译时需要)
- MQTT Broker (如 Mosquitto)
- MAVLink 数据源 (PX4/ArduPilot SITL 或真实飞控)

## 安装

### 方式一：下载预编译二进制

```bash
# 树莓派 / Jetson (ARM64)
wget https://github.com/open-uav/telemetry-bridge/releases/download/v0.1.0/outb-linux-arm64
chmod +x outb-linux-arm64

# Linux x86_64
wget https://github.com/open-uav/telemetry-bridge/releases/download/v0.1.0/outb-linux-amd64
chmod +x outb-linux-amd64
```

### 方式二：从源码编译

```bash
git clone https://github.com/open-uav/telemetry-bridge.git
cd telemetry-bridge
make build                  # 本机编译
make build-linux-arm64      # 交叉编译至 ARM64
```

## 配置

1. 复制示例配置文件：

```bash
cp configs/config.example.yaml configs/config.yaml
```

2. 编辑配置文件：

```yaml
# configs/config.yaml
server:
  log_level: info

mavlink:
  enabled: true
  connection_type: udp       # udp | tcp | serial
  address: "0.0.0.0:14550"   # UDP 监听地址

mqtt:
  enabled: true
  broker: "tcp://localhost:1883"
  client_id: "outb-001"
  topic_prefix: "uav/telemetry"
  qos: 1
  lwt:
    enabled: true
    topic: "uav/status"
    message: "offline"

throttle:
  default_rate_hz: 1.0       # 发布频率 (Hz)
```

## 运行

```bash
# 使用默认配置路径
./bin/outb

# 指定配置文件
./bin/outb /path/to/config.yaml
```

启动输出示例：

```
Open-UAV-Telemetry-Bridge v0.1.0
Protocol-agnostic UAV telemetry gateway

2026/01/18 16:50:00 Configuration loaded from configs/config.yaml
2026/01/18 16:50:00 Core engine created (throttle rate: 1.0 Hz)
2026/01/18 16:50:00 MAVLink adapter registered (udp: 0.0.0.0:14550)
2026/01/18 16:50:00 MQTT publisher registered (broker: tcp://localhost:1883)
2026/01/18 16:50:00 [Engine] Publisher started: mqtt
2026/01/18 16:50:00 [Engine] Adapter started: mavlink
2026/01/18 16:50:00 Gateway is running. Press Ctrl+C to stop.
```

## 本地测试

### 1. 启动 MQTT Broker

```bash
# macOS
brew install mosquitto
brew services start mosquitto

# Ubuntu/Debian
sudo apt install mosquitto mosquitto-clients
sudo systemctl start mosquitto
```

### 2. 订阅 MQTT 主题

```bash
mosquitto_sub -h localhost -t "uav/telemetry/#" -v
```

### 3. 启动 ArduPilot SITL (模拟器)

```bash
# 安装 ArduPilot SITL
# https://ardupilot.org/dev/docs/setting-up-sitl-on-linux.html

cd ardupilot
sim_vehicle.py -v ArduCopter --out=udp:127.0.0.1:14550
```

### 4. 启动网关

```bash
./bin/outb configs/config.yaml
```

### 5. 查看输出

在 MQTT 订阅终端中将看到类似输出：

```json
uav/telemetry/mavlink-1/state {
  "device_id": "mavlink-1",
  "timestamp": 1705596600000,
  "protocol_source": "mavlink",
  "location": {
    "lat": -35.3632621,
    "lon": 149.1652374,
    "alt_baro": 10.5,
    "alt_gnss": 584.0,
    "coordinate_system": "WGS84"
  },
  "attitude": {
    "roll": 0.01,
    "pitch": -0.02,
    "yaw": 270.5
  },
  "status": {
    "battery_percent": 100,
    "flight_mode": "LOITER",
    "armed": false,
    "signal_quality": 0
  },
  "velocity": {
    "vx": 0.0,
    "vy": 0.0,
    "vz": 0.0
  }
}
```

## MQTT Topic 结构

| Topic | 描述 |
|-------|------|
| `uav/telemetry/{device_id}/state` | 完整状态 JSON |
| `uav/status/{client_id}` | 在线/离线状态 (LWT) |

## 树莓派部署

```bash
# 1. 复制文件到树莓派
scp bin/outb-linux-arm64 pi@raspberrypi:~/outb
scp configs/config.example.yaml pi@raspberrypi:~/config.yaml

# 2. SSH 登录树莓派
ssh pi@raspberrypi

# 3. 编辑配置 (修改 MQTT broker 地址等)
nano ~/config.yaml

# 4. 运行
chmod +x ~/outb
./outb ~/config.yaml
```

### 串口连接 (直连飞控)

```yaml
mavlink:
  enabled: true
  connection_type: serial
  serial_port: "/dev/ttyUSB0"
  serial_baud: 57600
```

## 常见问题

### Q: MQTT 连接失败

确认 MQTT Broker 正在运行：
```bash
systemctl status mosquitto
```

### Q: 收不到 MAVLink 数据

1. 检查飞控是否正在发送数据到配置的端口
2. 使用 `mavproxy.py --master=udp:0.0.0.0:14550` 验证数据流

### Q: 树莓派权限问题

串口访问需要权限：
```bash
sudo usermod -a -G dialout $USER
# 重新登录生效
```

## 下一步

- 查看 [配置说明](docs/CONFIG.md) 了解所有配置选项
- 查看 [架构设计](README.md) 了解技术细节
