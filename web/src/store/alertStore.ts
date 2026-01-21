// Zustand store for alert state management

import { create } from 'zustand';
import type { Alert, AlertRule } from '../api/types';
import { api } from '../api/client';

interface AlertStats {
  total_alerts: number;
  unacknowledged: number;
  rules_count: number;
  devices_with_alerts: number;
}

interface AlertStore {
  // State
  alerts: Alert[];
  rules: AlertRule[];
  stats: AlertStats | null;
  isLoading: boolean;
  error: string | null;
  filter: {
    acknowledged: boolean | null;
    deviceId: string;
  };

  // Actions
  fetchAlerts: () => Promise<void>;
  fetchRules: () => Promise<void>;
  acknowledgeAlert: (alertId: string) => Promise<void>;
  clearAlerts: () => Promise<void>;
  createRule: (rule: Partial<AlertRule>) => Promise<void>;
  updateRule: (ruleId: string, rule: Partial<AlertRule>) => Promise<void>;
  deleteRule: (ruleId: string) => Promise<void>;
  setFilter: (filter: Partial<{ acknowledged: boolean | null; deviceId: string }>) => void;
  clear: () => void;
}

export const useAlertStore = create<AlertStore>((set, get) => ({
  // Initial state
  alerts: [],
  rules: [],
  stats: null,
  isLoading: false,
  error: null,
  filter: {
    acknowledged: null,
    deviceId: '',
  },

  // Actions
  fetchAlerts: async () => {
    set({ isLoading: true, error: null });
    try {
      const { filter } = get();
      const response = await api.getAlerts({
        acknowledged: filter.acknowledged ?? undefined,
        deviceId: filter.deviceId || undefined,
        limit: 200,
      });
      set({
        alerts: response.alerts,
        stats: response.stats,
        isLoading: false,
      });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load alerts',
        isLoading: false,
      });
    }
  },

  fetchRules: async () => {
    try {
      const response = await api.getAlertRules();
      set({ rules: response.rules });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load rules',
      });
    }
  },

  acknowledgeAlert: async (alertId: string) => {
    try {
      await api.acknowledgeAlert(alertId);
      // Update local state
      set((state) => ({
        alerts: state.alerts.map((a) =>
          a.id === alertId ? { ...a, acknowledged: true, acked_at: Date.now() } : a
        ),
        stats: state.stats
          ? { ...state.stats, unacknowledged: Math.max(0, state.stats.unacknowledged - 1) }
          : null,
      }));
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to acknowledge alert',
      });
    }
  },

  clearAlerts: async () => {
    try {
      await api.clearAlerts();
      set({ alerts: [], stats: null });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to clear alerts',
      });
    }
  },

  createRule: async (rule: Partial<AlertRule>) => {
    try {
      const newRule = await api.createAlertRule(rule);
      set((state) => ({
        rules: [...state.rules, newRule],
      }));
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to create rule',
      });
      throw err;
    }
  },

  updateRule: async (ruleId: string, rule: Partial<AlertRule>) => {
    try {
      const updatedRule = await api.updateAlertRule(ruleId, rule);
      set((state) => ({
        rules: state.rules.map((r) => (r.id === ruleId ? updatedRule : r)),
      }));
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update rule',
      });
      throw err;
    }
  },

  deleteRule: async (ruleId: string) => {
    try {
      await api.deleteAlertRule(ruleId);
      set((state) => ({
        rules: state.rules.filter((r) => r.id !== ruleId),
      }));
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to delete rule',
      });
      throw err;
    }
  },

  setFilter: (filter) => {
    set((state) => ({
      filter: { ...state.filter, ...filter },
    }));
  },

  clear: () => {
    set({
      alerts: [],
      rules: [],
      stats: null,
      isLoading: false,
      error: null,
    });
  },
}));

// Selector hooks
export const useAlerts = () => useAlertStore((state) => state.alerts);
export const useAlertRules = () => useAlertStore((state) => state.rules);
export const useAlertStats = () => useAlertStore((state) => state.stats);
export const useAlertsLoading = () => useAlertStore((state) => state.isLoading);
export const useAlertsError = () => useAlertStore((state) => state.error);
export const useAlertFilter = () => useAlertStore((state) => state.filter);
