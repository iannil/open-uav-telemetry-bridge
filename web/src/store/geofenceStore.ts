// Zustand store for geofence state management

import { create } from 'zustand';
import type { Geofence, GeofenceBreach, GeofenceStats } from '../api/types';
import { api } from '../api/client';

interface GeofenceStore {
  // State
  geofences: Geofence[];
  breaches: GeofenceBreach[];
  stats: GeofenceStats | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchGeofences: () => Promise<void>;
  fetchBreaches: (options?: { deviceId?: string; geofenceId?: string; limit?: number }) => Promise<void>;
  fetchStats: () => Promise<void>;
  createGeofence: (geofence: Partial<Geofence>) => Promise<Geofence>;
  updateGeofence: (id: string, geofence: Partial<Geofence>) => Promise<void>;
  deleteGeofence: (id: string) => Promise<void>;
  clearBreaches: () => Promise<void>;
  clear: () => void;
}

export const useGeofenceStore = create<GeofenceStore>((set) => ({
  // Initial state
  geofences: [],
  breaches: [],
  stats: null,
  isLoading: false,
  error: null,

  // Actions
  fetchGeofences: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await api.getGeofences();
      set({
        geofences: response.geofences,
        isLoading: false,
      });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load geofences',
        isLoading: false,
      });
    }
  },

  fetchBreaches: async (options) => {
    try {
      const response = await api.getGeofenceBreaches(options);
      set({ breaches: response.breaches });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load breaches',
      });
    }
  },

  fetchStats: async () => {
    try {
      const stats = await api.getGeofenceStats();
      set({ stats });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load stats',
      });
    }
  },

  createGeofence: async (geofence: Partial<Geofence>) => {
    try {
      const newGeofence = await api.createGeofence(geofence);
      set((state) => ({
        geofences: [...state.geofences, newGeofence],
      }));
      return newGeofence;
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to create geofence',
      });
      throw err;
    }
  },

  updateGeofence: async (id: string, geofence: Partial<Geofence>) => {
    try {
      const updated = await api.updateGeofence(id, geofence);
      set((state) => ({
        geofences: state.geofences.map((g) => (g.id === id ? updated : g)),
      }));
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update geofence',
      });
      throw err;
    }
  },

  deleteGeofence: async (id: string) => {
    try {
      await api.deleteGeofence(id);
      set((state) => ({
        geofences: state.geofences.filter((g) => g.id !== id),
      }));
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to delete geofence',
      });
      throw err;
    }
  },

  clearBreaches: async () => {
    try {
      await api.clearGeofenceBreaches();
      set({ breaches: [] });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to clear breaches',
      });
    }
  },

  clear: () => {
    set({
      geofences: [],
      breaches: [],
      stats: null,
      isLoading: false,
      error: null,
    });
  },
}));

// Selector hooks
export const useGeofences = () => useGeofenceStore((state) => state.geofences);
export const useGeofenceBreaches = () => useGeofenceStore((state) => state.breaches);
export const useGeofenceStats = () => useGeofenceStore((state) => state.stats);
export const useGeofencesLoading = () => useGeofenceStore((state) => state.isLoading);
export const useGeofencesError = () => useGeofenceStore((state) => state.error);
