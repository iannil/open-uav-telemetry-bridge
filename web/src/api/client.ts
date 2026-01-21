// HTTP API Client for Open-UAV-Telemetry-Bridge

import type {
  StatusResponse,
  DronesResponse,
  DroneState,
  TrackResponse,
  LoginRequest,
  LoginResponse,
  AuthStatusResponse,
  AppConfig,
  MAVLinkConfig,
  DJIConfig,
  ThrottleConfig,
  CoordinateConfig,
  TrackConfig,
  LogsResponse,
  LogLevel,
  AlertsResponse,
  AlertRulesResponse,
  AlertRule,
  Geofence,
  GeofencesResponse,
  BreachesResponse,
  GeofenceStats,
} from './types';
import { useAuthStore } from '../store/authStore';

const API_BASE = '/api/v1';

// Get token from auth store
function getAuthToken(): string | null {
  return useAuthStore.getState().token;
}

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const token = getAuthToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options?.headers as Record<string, string>) || {}),
  };

  // Add Authorization header if token exists
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  });

  // Handle 401 Unauthorized - logout user
  if (response.status === 401) {
    const authEnabled = useAuthStore.getState().authEnabled;
    if (authEnabled) {
      useAuthStore.getState().logout();
    }
  }

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

export const api = {
  // Authentication
  login: (credentials: LoginRequest): Promise<LoginResponse> => {
    return fetchAPI<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
  },

  logout: (): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/auth/logout', {
      method: 'POST',
    });
  },

  getMe: (): Promise<AuthStatusResponse> => {
    return fetchAPI<AuthStatusResponse>('/auth/me');
  },

  // Get gateway status
  getStatus: (): Promise<StatusResponse> => {
    return fetchAPI<StatusResponse>('/status');
  },

  // Get all drones
  getDrones: (): Promise<DronesResponse> => {
    return fetchAPI<DronesResponse>('/drones');
  },

  // Get single drone by ID
  getDrone: (deviceId: string): Promise<DroneState> => {
    return fetchAPI<DroneState>(`/drones/${encodeURIComponent(deviceId)}`);
  },

  // Get drone track history
  getTrack: (deviceId: string, limit?: number, since?: number): Promise<TrackResponse> => {
    const params = new URLSearchParams();
    if (limit !== undefined) params.set('limit', String(limit));
    if (since !== undefined) params.set('since', String(since));
    const query = params.toString();
    return fetchAPI<TrackResponse>(`/drones/${encodeURIComponent(deviceId)}/track${query ? `?${query}` : ''}`);
  },

  // Clear drone track
  clearTrack: (deviceId: string): Promise<void> => {
    return fetchAPI<void>(`/drones/${encodeURIComponent(deviceId)}/track`, {
      method: 'DELETE',
    });
  },

  // Health check
  checkHealth: async (): Promise<boolean> => {
    try {
      const response = await fetch('/health');
      return response.ok;
    } catch {
      return false;
    }
  },

  // Configuration
  getConfig: (): Promise<AppConfig> => {
    return fetchAPI<AppConfig>('/config');
  },

  updateMAVLinkConfig: (config: Partial<MAVLinkConfig>): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/config/adapters/mavlink', {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  },

  updateDJIConfig: (config: Partial<DJIConfig>): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/config/adapters/dji', {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  },

  updateMQTTConfig: (config: Record<string, unknown>): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/config/publishers/mqtt', {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  },

  updateGB28181Config: (config: Record<string, unknown>): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/config/publishers/gb28181', {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  },

  updateThrottleConfig: (config: ThrottleConfig): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/config/throttle', {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  },

  updateCoordinateConfig: (config: CoordinateConfig): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/config/coordinate', {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  },

  updateTrackConfig: (config: TrackConfig): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/config/track', {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  },

  applyConfig: (): Promise<{ message: string; restart_required: boolean }> => {
    return fetchAPI<{ message: string; restart_required: boolean }>('/config/apply', {
      method: 'POST',
    });
  },

  exportConfig: async (): Promise<string> => {
    const token = getAuthToken();
    const headers: Record<string, string> = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE}/config/export`, {
      method: 'POST',
      headers,
    });

    if (!response.ok) {
      throw new Error('Failed to export config');
    }

    return response.text();
  },

  // Logs
  getLogs: (options?: { level?: LogLevel; source?: string; limit?: number; sinceId?: number }): Promise<LogsResponse> => {
    const params = new URLSearchParams();
    if (options?.level) params.set('level', options.level);
    if (options?.source) params.set('source', options.source);
    if (options?.limit) params.set('limit', String(options.limit));
    if (options?.sinceId) params.set('since_id', String(options.sinceId));
    const query = params.toString();
    return fetchAPI<LogsResponse>(`/logs${query ? `?${query}` : ''}`);
  },

  clearLogs: (): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/logs', {
      method: 'DELETE',
    });
  },

  // Get log stream URL with auth token
  getLogStreamUrl: (level?: LogLevel): string => {
    const token = getAuthToken();
    const params = new URLSearchParams();
    if (level) params.set('level', level);
    if (token) params.set('token', token);
    const query = params.toString();
    return `${API_BASE}/logs/stream${query ? `?${query}` : ''}`;
  },

  // Alerts
  getAlerts: (options?: { deviceId?: string; acknowledged?: boolean; limit?: number }): Promise<AlertsResponse> => {
    const params = new URLSearchParams();
    if (options?.deviceId) params.set('device_id', options.deviceId);
    if (options?.acknowledged !== undefined) params.set('acknowledged', String(options.acknowledged));
    if (options?.limit) params.set('limit', String(options.limit));
    const query = params.toString();
    return fetchAPI<AlertsResponse>(`/alerts${query ? `?${query}` : ''}`);
  },

  acknowledgeAlert: (alertId: string): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>(`/alerts/${alertId}/ack`, {
      method: 'POST',
    });
  },

  clearAlerts: (): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/alerts', {
      method: 'DELETE',
    });
  },

  getAlertRules: (): Promise<AlertRulesResponse> => {
    return fetchAPI<AlertRulesResponse>('/alerts/rules');
  },

  createAlertRule: (rule: Partial<AlertRule>): Promise<AlertRule> => {
    return fetchAPI<AlertRule>('/alerts/rules', {
      method: 'POST',
      body: JSON.stringify(rule),
    });
  },

  updateAlertRule: (ruleId: string, rule: Partial<AlertRule>): Promise<AlertRule> => {
    return fetchAPI<AlertRule>(`/alerts/rules/${ruleId}`, {
      method: 'PUT',
      body: JSON.stringify(rule),
    });
  },

  deleteAlertRule: (ruleId: string): Promise<void> => {
    return fetchAPI<void>(`/alerts/rules/${ruleId}`, {
      method: 'DELETE',
    });
  },

  // Geofences
  getGeofences: (): Promise<GeofencesResponse> => {
    return fetchAPI<GeofencesResponse>('/geofences');
  },

  getGeofence: (id: string): Promise<Geofence> => {
    return fetchAPI<Geofence>(`/geofences/${id}`);
  },

  createGeofence: (geofence: Partial<Geofence>): Promise<Geofence> => {
    return fetchAPI<Geofence>('/geofences', {
      method: 'POST',
      body: JSON.stringify(geofence),
    });
  },

  updateGeofence: (id: string, geofence: Partial<Geofence>): Promise<Geofence> => {
    return fetchAPI<Geofence>(`/geofences/${id}`, {
      method: 'PUT',
      body: JSON.stringify(geofence),
    });
  },

  deleteGeofence: (id: string): Promise<void> => {
    return fetchAPI<void>(`/geofences/${id}`, {
      method: 'DELETE',
    });
  },

  getGeofenceBreaches: (options?: { deviceId?: string; geofenceId?: string; limit?: number }): Promise<BreachesResponse> => {
    const params = new URLSearchParams();
    if (options?.deviceId) params.set('device_id', options.deviceId);
    if (options?.geofenceId) params.set('geofence_id', options.geofenceId);
    if (options?.limit) params.set('limit', String(options.limit));
    const query = params.toString();
    return fetchAPI<BreachesResponse>(`/geofences/breaches${query ? `?${query}` : ''}`);
  },

  clearGeofenceBreaches: (): Promise<{ message: string }> => {
    return fetchAPI<{ message: string }>('/geofences/breaches', {
      method: 'DELETE',
    });
  },

  getGeofenceStats: (): Promise<GeofenceStats> => {
    return fetchAPI<GeofenceStats>('/geofences/stats');
  },
};

export default api;
