// API Types for Open-UAV-Telemetry-Bridge

// Authentication Types
export interface User {
  username: string;
  role: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  expires_at: number;
  user: User;
}

export interface AuthStatusResponse {
  auth_enabled: boolean;
  user?: User;
  message?: string;
}

export interface Location {
  lat: number;
  lon: number;
  alt_baro: number;
  alt_gnss: number;
  coordinate_system: string;
  gcj02?: {
    lat: number;
    lon: number;
  };
  bd09?: {
    lat: number;
    lon: number;
  };
}

export interface Attitude {
  roll: number;
  pitch: number;
  yaw: number;
}

export interface Velocity {
  vx: number;
  vy: number;
  vz: number;
}

export interface Status {
  armed: boolean;
  flight_mode: string;
  battery_percent: number;
  battery_voltage: number;
  gps_fix_type: number;
  satellites_visible: number;
}

export interface DroneState {
  device_id: string;
  timestamp: number;
  protocol_source: string;
  location: Location;
  attitude: Attitude;
  velocity: Velocity;
  status: Status;
}

export interface TrackPoint {
  timestamp: number;
  lat: number;
  lon: number;
  alt: number;
}

export interface Stats {
  active_drones: number;
  websocket_clients: number;
}

export interface AdapterStatus {
  name: string;
  enabled: boolean;
}

export interface StatusResponse {
  version: string;
  uptime_seconds: number;
  adapters: AdapterStatus[];
  publishers: string[];
  stats: Stats;
}

export interface DronesResponse {
  count: number;
  drones: DroneState[];
}

export interface TrackResponse {
  device_id: string;
  count: number;
  points: TrackPoint[];
  total_size: number;
}

export interface ErrorResponse {
  error: string;
  device_id?: string;
}

// WebSocket message types
export type WSMessageType = 'state_update' | 'drone_online' | 'drone_offline';

export interface WSMessage {
  type: WSMessageType;
  device_id?: string;
  data?: DroneState;
}

// Configuration Types
export interface ServerConfig {
  log_level: string;
}

export interface MAVLinkConfig {
  enabled: boolean;
  connection_type: string;
  address: string;
  serial_port: string;
  serial_baud: number;
}

export interface DJIConfig {
  enabled: boolean;
  listen_address: string;
  max_clients: number;
}

export interface LWTConfig {
  enabled: boolean;
  topic: string;
  message: string;
}

export interface MQTTConfig {
  enabled: boolean;
  broker: string;
  client_id: string;
  topic_prefix: string;
  qos: number;
  username: string;
  has_password: boolean;
  lwt: LWTConfig;
}

export interface GB28181Config {
  enabled: boolean;
  device_id: string;
  device_name: string;
  local_ip: string;
  local_port: number;
  server_id: string;
  server_ip: string;
  server_port: number;
  server_domain: string;
  username: string;
  has_password: boolean;
  transport: string;
  register_expires: number;
  heartbeat_interval: number;
  position_interval: number;
}

export interface AuthConfig {
  enabled: boolean;
  username: string;
  has_password_hash: boolean;
  has_jwt_secret: boolean;
  token_expiry_hours: number;
}

export interface HTTPConfig {
  enabled: boolean;
  address: string;
  cors_enabled: boolean;
  cors_origins: string[];
  webui_enabled: boolean;
  auth: AuthConfig;
}

export interface ThrottleConfig {
  default_rate_hz: number;
  min_rate_hz: number;
  max_rate_hz: number;
}

export interface CoordinateConfig {
  convert_gcj02: boolean;
  convert_bd09: boolean;
}

export interface TrackConfig {
  enabled: boolean;
  max_points_per_drone: number;
  sample_interval_ms: number;
}

export interface AppConfig {
  server: ServerConfig;
  mavlink: MAVLinkConfig;
  dji: DJIConfig;
  mqtt: MQTTConfig;
  gb28181: GB28181Config;
  http: HTTPConfig;
  throttle: ThrottleConfig;
  coordinate: CoordinateConfig;
  track: TrackConfig;
}

// Log Types
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export interface LogEntry {
  id: number;
  timestamp: number;
  level: LogLevel;
  source: string;
  message: string;
}

export interface LogsResponse {
  logs: LogEntry[];
  count: number;
  total: number;
}

export interface LogStreamEvent {
  type: 'connected' | 'log' | 'ping';
  data: LogEntry | { subscriber_id: string } | { time: number };
}

// Alert Types
export type AlertType = 'battery_low' | 'connection_lost' | 'signal_weak' | 'geofence_breach' | 'custom';
export type AlertSeverity = 'info' | 'warning' | 'critical';

export interface Alert {
  id: string;
  rule_id: string;
  type: AlertType;
  severity: AlertSeverity;
  device_id: string;
  message: string;
  value?: number;
  threshold?: number;
  timestamp: number;
  acknowledged: boolean;
  acked_at?: number;
  acked_by?: string;
}

export interface AlertCondition {
  field: string;
  operator: string;
  threshold: number;
}

export interface AlertRule {
  id: string;
  name: string;
  type: AlertType;
  severity: AlertSeverity;
  enabled: boolean;
  condition: AlertCondition;
  cooldown_ms: number;
  created_at: number;
  updated_at: number;
}

export interface AlertsResponse {
  alerts: Alert[];
  count: number;
  stats: {
    total_alerts: number;
    unacknowledged: number;
    rules_count: number;
    devices_with_alerts: number;
  };
}

export interface AlertRulesResponse {
  rules: AlertRule[];
  count: number;
}

// Geofence Types
export type GeofenceType = 'polygon' | 'circle';
export type BreachType = 'enter' | 'exit';

export interface Geofence {
  id: string;
  name: string;
  type: GeofenceType;
  coordinates?: number[][];  // [[lat, lon], ...] for polygon
  center?: number[];         // [lat, lon] for circle
  radius?: number;           // in meters for circle
  min_altitude?: number;
  max_altitude?: number;
  alert_on_enter: boolean;
  alert_on_exit: boolean;
  enabled: boolean;
  created_at: number;
  updated_at: number;
}

export interface GeofenceBreach {
  id: string;
  geofence_id: string;
  device_id: string;
  type: BreachType;
  lat: number;
  lon: number;
  alt: number;
  timestamp: number;
}

export interface GeofencesResponse {
  geofences: Geofence[];
  count: number;
}

export interface BreachesResponse {
  breaches: GeofenceBreach[];
  count: number;
}

export interface GeofenceStats {
  total_geofences: number;
  enabled_geofences: number;
  total_breaches: number;
  tracked_devices: number;
}
