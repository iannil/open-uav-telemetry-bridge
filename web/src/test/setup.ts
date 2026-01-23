// Test setup file for Vitest
import '@testing-library/jest-dom';

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  }),
});

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Reset stores between tests
import { afterEach } from 'vitest';
import { useDroneStore } from '../store/droneStore';
import { useAuthStore } from '../store/authStore';

afterEach(() => {
  // Reset drone store
  useDroneStore.setState({
    drones: new Map(),
    selectedDroneId: null,
    gatewayStatus: null,
    isConnected: false,
    lastUpdate: 0,
  });

  // Reset auth store (need to clear localStorage too)
  useAuthStore.setState({
    token: null,
    user: null,
    expiresAt: null,
    authEnabled: null,
    isAuthenticated: false,
  });
  localStorage.clear();
});
