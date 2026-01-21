// Zustand store for configuration state management

import { create } from 'zustand';
import type { AppConfig } from '../api/types';
import { api } from '../api/client';

interface ConfigStore {
  // State
  config: AppConfig | null;
  isLoading: boolean;
  error: string | null;
  hasChanges: boolean;

  // Actions
  fetchConfig: () => Promise<void>;
  setConfig: (config: AppConfig) => void;
  setHasChanges: (hasChanges: boolean) => void;
  setError: (error: string | null) => void;
  clear: () => void;
}

export const useConfigStore = create<ConfigStore>((set) => ({
  // Initial state
  config: null,
  isLoading: false,
  error: null,
  hasChanges: false,

  // Actions
  fetchConfig: async () => {
    set({ isLoading: true, error: null });
    try {
      const config = await api.getConfig();
      set({ config, isLoading: false, hasChanges: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load config',
        isLoading: false,
      });
    }
  },

  setConfig: (config: AppConfig) => {
    set({ config, hasChanges: true });
  },

  setHasChanges: (hasChanges: boolean) => {
    set({ hasChanges });
  },

  setError: (error: string | null) => {
    set({ error });
  },

  clear: () => {
    set({ config: null, isLoading: false, error: null, hasChanges: false });
  },
}));

// Selector hooks
export const useConfig = () => useConfigStore((state) => state.config);
export const useConfigLoading = () => useConfigStore((state) => state.isLoading);
export const useConfigError = () => useConfigStore((state) => state.error);
export const useConfigHasChanges = () => useConfigStore((state) => state.hasChanges);
