// Mock data for tests
import type { DroneState, StatusResponse } from '../api/types';

export const mockDroneState: DroneState = {
  device_id: 'drone-001',
  timestamp: Date.now(),
  protocol_source: 'mavlink',
  location: {
    lat: 39.9042,
    lon: 116.4074,
    alt_baro: 100,
    alt_gnss: 105,
    coordinate_system: 'WGS84',
  },
  attitude: {
    roll: 2.5,
    pitch: -1.5,
    yaw: 45.0,
  },
  velocity: {
    vx: 5.0,
    vy: 3.0,
    vz: -1.0,
  },
  status: {
    armed: true,
    flight_mode: 'LOITER',
    battery_percent: 75,
    battery_voltage: 22.4,
    gps_fix_type: 3,
    satellites_visible: 12,
  },
};

export const mockDroneState2: DroneState = {
  device_id: 'drone-002',
  timestamp: Date.now(),
  protocol_source: 'dji_mobile_sdk',
  location: {
    lat: 31.2304,
    lon: 121.4737,
    alt_baro: 50,
    alt_gnss: 55,
    coordinate_system: 'WGS84',
  },
  attitude: {
    roll: 0,
    pitch: 0,
    yaw: 180.0,
  },
  velocity: {
    vx: 0,
    vy: 0,
    vz: 0,
  },
  status: {
    armed: false,
    flight_mode: 'LANDING',
    battery_percent: 25,
    battery_voltage: 20.1,
    gps_fix_type: 3,
    satellites_visible: 8,
  },
};

export const mockGatewayStatus: StatusResponse = {
  version: '0.4.0-dev',
  uptime_seconds: 3600,
  adapters: [
    { name: 'mavlink', enabled: true },
    { name: 'dji', enabled: true },
  ],
  publishers: ['mqtt', 'gb28181'],
  stats: {
    active_drones: 2,
    websocket_clients: 1,
  },
};

export function createMockDrone(overrides: Partial<DroneState> = {}): DroneState {
  return {
    ...mockDroneState,
    device_id: `drone-${Date.now()}`,
    timestamp: Date.now(),
    ...overrides,
  };
}
