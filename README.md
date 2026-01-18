# Open-UAV-Telemetry-Bridge

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)]()

**Protocol-agnostic UAV telemetry edge gateway** - Bridge the gap between drone protocols and cloud platforms.

[中文文档](README_zh-CN.md)

---

## Overview

Open-UAV-Telemetry-Bridge (OUTB) is a lightweight, high-performance telemetry gateway designed for the Internet of Drones (IoD). It translates between various drone protocols (MAVLink, DJI, GB/T 28181) and outputs standardized data via MQTT, WebSocket, HTTP, or gRPC.

### Why OUTB?

- **Protocol Fragmentation**: PX4 uses MAVLink, DJI uses proprietary protocols, government platforms require GB/T 28181
- **Coordinate System Chaos**: GPS outputs WGS84, but Chinese maps require GCJ02/BD09 offsets
- **Bandwidth Constraints**: Raw telemetry at 50-100Hz is too much for 4G networks
- **Integration Complexity**: Each platform needs a custom adapter

OUTB solves all these problems with a unified, pluggable architecture.

---

## Features

### Core Features

- **Multi-Protocol Support**: MAVLink (UDP/TCP/Serial), DJI (via Android Forwarder), GB/T 28181 (planned)
- **Unified Data Model**: Standardized JSON output regardless of source protocol
- **Coordinate Conversion**: Automatic WGS84 → GCJ02/BD09 transformation for China maps
- **Frequency Throttling**: Configurable downsampling (e.g., 50Hz → 1Hz) to save bandwidth
- **State Caching**: In-memory state store with historical track storage

### Output Interfaces

- **MQTT Publisher**: Standard MQTT 3.1.1 with LWT (Last Will and Testament) support
- **HTTP REST API**: Query drone states, health checks, gateway status
- **WebSocket**: Real-time push notifications for state updates
- **Track Storage**: Historical trajectory with ring buffer (configurable retention)

### Operational Features

- **Edge-Ready**: Runs on Raspberry Pi 4, Jetson Nano, or cloud servers
- **Zero Dependencies**: Single binary, no external runtime required
- **Hot Configuration**: YAML-based configuration

---

## Quick Start

### Prerequisites

- Go 1.21 or higher
- (Optional) MQTT Broker (e.g., Mosquitto)

### Installation

```bash
# Clone the repository
git clone https://github.com/iannil/open-uav-telemetry-bridge.git
cd open-uav-telemetry-bridge

# Build
make build

# Or build for Raspberry Pi / Jetson
make build-linux-arm64
```

### Configuration

```bash
# Copy example configuration
cp configs/config.example.yaml configs/config.yaml

# Edit as needed
vim configs/config.yaml
```

### Run

```bash
# Run with configuration file
./bin/outb configs/config.yaml
```

### Verify

```bash
# Check health
curl http://localhost:8080/health

# Get gateway status
curl http://localhost:8080/api/v1/status

# List connected drones
curl http://localhost:8080/api/v1/drones
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   Northbound Interfaces                      │
│  ┌──────────┐  ┌───────────┐  ┌──────────┐  ┌────────────┐  │
│  │   MQTT   │  │ WebSocket │  │   HTTP   │  │    gRPC    │  │
│  └────┬─────┘  └─────┬─────┘  └────┬─────┘  └─────┬──────┘  │
└───────┼──────────────┼─────────────┼──────────────┼─────────┘
        │              │             │              │
        └──────────────┴──────┬──────┴──────────────┘
                              │
┌─────────────────────────────┼───────────────────────────────┐
│                        Core Engine                           │
│  ┌─────────────┐  ┌─────────┴────────┐  ┌────────────────┐  │
│  │  Throttler  │  │   State Store    │  │ Track Storage  │  │
│  └─────────────┘  └──────────────────┘  └────────────────┘  │
│  ┌─────────────────────────────────────────────────────────┐│
│  │           Coordinate Converter (WGS84→GCJ02/BD09)       ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────┼───────────────────────────────┐
│                    Southbound Adapters                       │
│  ┌──────────────┐  ┌────────────────┐  ┌─────────────────┐  │
│  │   MAVLink    │  │  DJI Forwarder │  │   GB/T 28181    │  │
│  │ (UDP/TCP/Ser)│  │  (TCP Server)  │  │    (Planned)    │  │
│  └──────┬───────┘  └───────┬────────┘  └────────┬────────┘  │
└─────────┼──────────────────┼────────────────────┼───────────┘
          │                  │                    │
     ┌────┴────┐      ┌──────┴──────┐      ┌──────┴──────┐
     │PX4/Ardu │      │ DJI Drone   │      │ Gov Platform│
     │ Pilot   │      │ (via App)   │      │             │
     └─────────┘      └─────────────┘      └─────────────┘
```

---

## API Reference

### HTTP Endpoints

| Method | Endpoint | Description |
| -------- | ---------- | ------------- |
| GET | `/health` | Health check |
| GET | `/api/v1/status` | Gateway status and statistics |
| GET | `/api/v1/drones` | List all connected drones |
| GET | `/api/v1/drones/{id}` | Get specific drone state |
| GET | `/api/v1/drones/{id}/track` | Get historical track points |
| DELETE | `/api/v1/drones/{id}/track` | Clear track history |

### WebSocket

Connect to `ws://localhost:8080/api/v1/ws` for real-time updates.

**Message Types:**

```json
// State update (server → client)
{
  "type": "state_update",
  "data": { /* DroneState */ }
}

// Subscribe to specific drones (client → server)
{
  "type": "subscribe",
  "device_ids": ["drone-001", "drone-002"]
}

// Unsubscribe (client → server)
{
  "type": "unsubscribe",
  "device_ids": ["drone-001"]
}
```

### Unified Data Model (DroneState)

```json
{
  "device_id": "mavlink-001",
  "timestamp": 1709882231000,
  "protocol_source": "mavlink",
  "location": {
    "lat": 39.9042,
    "lon": 116.4074,
    "lat_gcj02": 39.9066,
    "lon_gcj02": 116.4136,
    "alt": 120.5,
    "coordinate_system": "WGS84"
  },
  "attitude": {
    "roll": 0.05,
    "pitch": -0.12,
    "yaw": 180.0
  },
  "velocity": {
    "vx": 10.5,
    "vy": 0.0,
    "vz": -0.5
  },
  "status": {
    "battery_percent": 85,
    "flight_mode": "AUTO",
    "armed": true,
    "signal_quality": 95
  }
}
```

---

## Configuration

```yaml
# Server settings
server:
  log_level: info  # debug, info, warn, error

# MAVLink Adapter
mavlink:
  enabled: true
  connection_type: udp   # udp | tcp | serial
  address: "0.0.0.0:14550"

# DJI Forwarder Adapter
dji:
  enabled: false
  listen_address: "0.0.0.0:14560"
  max_clients: 10

# MQTT Publisher
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

# HTTP API
http:
  enabled: true
  address: "0.0.0.0:8080"
  cors_enabled: true
  cors_origins: ["*"]

# Frequency Throttling
throttle:
  default_rate_hz: 1.0
  min_rate_hz: 0.5
  max_rate_hz: 10.0

# Coordinate Conversion (China maps)
coordinate:
  convert_gcj02: true   # For Amap, Tencent, Google China
  convert_bd09: false   # For Baidu Maps

# Track Storage
track:
  enabled: true
  max_points_per_drone: 10000
  sample_interval_ms: 1000
```

---

## Deployment Scenarios

### 1. Onboard Edge Gateway

Deploy on Raspberry Pi or Jetson mounted on the drone.

```
[Flight Controller] --Serial--> [OUTB on Pi] --4G--> [Cloud]
```

### 2. Ground Control Station

Run alongside your GCS software.

```
[Drone] --Radio--> [GCS + OUTB] --WiFi/4G--> [Cloud]
```

### 3. Cloud Aggregator

Centralized protocol conversion for fleet management.

```
[Drone Fleet] --TCP/UDP--> [OUTB on Cloud] --> [Backend Services]
```

---

## Roadmap

- [x] **v0.1** - MAVLink → MQTT basic pipeline
- [x] **v0.2** - DJI Android Forwarder app
- [x] **v0.3** - Coordinate conversion + HTTP API
- [x] **v0.3.1** - WebSocket + Track storage
- [ ] **v0.4** - GB/T 28181 national standard support
- [ ] **v1.0** - Web management dashboard

---

## Project Structure

```
├── cmd/outb/               # Application entry point
├── internal/
│   ├── adapters/           # Southbound protocol adapters
│   │   ├── mavlink/        # MAVLink (UDP/TCP/Serial)
│   │   └── dji/            # DJI Forwarder (TCP Server)
│   ├── api/                # HTTP/WebSocket server
│   ├── config/             # YAML configuration
│   ├── core/               # Core engine
│   │   ├── coordinator/    # Coordinate conversion
│   │   ├── statestore/     # State caching
│   │   ├── throttler/      # Frequency control
│   │   └── trackstore/     # Historical tracks
│   ├── models/             # Unified data models
│   └── publishers/         # Northbound publishers
│       └── mqtt/           # MQTT publisher
├── android/                # DJI Android Forwarder (Kotlin)
├── configs/                # Configuration examples
├── scripts/                # Test utilities
└── docs/                   # Documentation
```

---

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [gomavlib](https://github.com/bluenviron/gomavlib) - Go MAVLink library
- [paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) - Eclipse Paho MQTT client
- [chi](https://github.com/go-chi/chi) - Lightweight HTTP router
