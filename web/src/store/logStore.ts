// Zustand store for log state management

import { create } from 'zustand';
import type { LogEntry, LogLevel } from '../api/types';
import { api } from '../api/client';

interface LogStore {
  // State
  logs: LogEntry[];
  isLoading: boolean;
  error: string | null;
  isStreaming: boolean;
  filter: {
    level: LogLevel;
    source: string;
  };

  // SSE connection
  eventSource: EventSource | null;

  // Actions
  fetchLogs: () => Promise<void>;
  clearLogs: () => Promise<void>;
  addLog: (log: LogEntry) => void;
  setFilter: (filter: Partial<{ level: LogLevel; source: string }>) => void;
  startStreaming: () => void;
  stopStreaming: () => void;
  clear: () => void;
}

const MAX_LOGS = 500; // Keep last 500 logs in memory

export const useLogStore = create<LogStore>((set, get) => ({
  // Initial state
  logs: [],
  isLoading: false,
  error: null,
  isStreaming: false,
  filter: {
    level: 'info',
    source: '',
  },
  eventSource: null,

  // Actions
  fetchLogs: async () => {
    set({ isLoading: true, error: null });
    try {
      const { filter } = get();
      const response = await api.getLogs({
        level: filter.level,
        source: filter.source || undefined,
        limit: 200,
      });
      set({ logs: response.logs, isLoading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load logs',
        isLoading: false,
      });
    }
  },

  clearLogs: async () => {
    try {
      await api.clearLogs();
      set({ logs: [] });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to clear logs',
      });
    }
  },

  addLog: (log: LogEntry) => {
    set((state) => {
      const newLogs = [...state.logs, log];
      // Keep only last MAX_LOGS
      if (newLogs.length > MAX_LOGS) {
        return { logs: newLogs.slice(-MAX_LOGS) };
      }
      return { logs: newLogs };
    });
  },

  setFilter: (filter) => {
    set((state) => ({
      filter: { ...state.filter, ...filter },
    }));
  },

  startStreaming: () => {
    const { eventSource, filter } = get();

    // Close existing connection
    if (eventSource) {
      eventSource.close();
    }

    const url = api.getLogStreamUrl(filter.level);
    const es = new EventSource(url);

    es.addEventListener('connected', () => {
      set({ isStreaming: true, error: null });
    });

    es.addEventListener('log', (event) => {
      try {
        const log = JSON.parse(event.data) as LogEntry;
        const { filter } = get();

        // Apply source filter client-side
        if (filter.source && log.source !== filter.source) {
          return;
        }

        get().addLog(log);
      } catch {
        // Ignore parse errors
      }
    });

    es.addEventListener('ping', () => {
      // Keep-alive, no action needed
    });

    es.onerror = () => {
      set({ isStreaming: false, error: 'Log stream disconnected' });
      es.close();
      set({ eventSource: null });
    };

    set({ eventSource: es });
  },

  stopStreaming: () => {
    const { eventSource } = get();
    if (eventSource) {
      eventSource.close();
      set({ eventSource: null, isStreaming: false });
    }
  },

  clear: () => {
    const { eventSource } = get();
    if (eventSource) {
      eventSource.close();
    }
    set({
      logs: [],
      isLoading: false,
      error: null,
      isStreaming: false,
      eventSource: null,
    });
  },
}));

// Selector hooks
export const useLogs = () => useLogStore((state) => state.logs);
export const useLogsLoading = () => useLogStore((state) => state.isLoading);
export const useLogsError = () => useLogStore((state) => state.error);
export const useLogsStreaming = () => useLogStore((state) => state.isStreaming);
export const useLogFilter = () => useLogStore((state) => state.filter);
