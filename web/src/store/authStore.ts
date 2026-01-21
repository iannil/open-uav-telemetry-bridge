// Zustand store for authentication state management

import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface User {
  username: string;
  role: string;
}

interface AuthStore {
  // State
  token: string | null;
  user: User | null;
  expiresAt: number | null;
  authEnabled: boolean | null; // null = unknown, fetching
  isAuthenticated: boolean;

  // Actions
  setAuth: (token: string, user: User, expiresAt: number) => void;
  setAuthEnabled: (enabled: boolean) => void;
  logout: () => void;
  isTokenExpired: () => boolean;
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      // Initial state
      token: null,
      user: null,
      expiresAt: null,
      authEnabled: null,
      isAuthenticated: false,

      // Actions
      setAuth: (token: string, user: User, expiresAt: number) =>
        set({
          token,
          user,
          expiresAt,
          isAuthenticated: true,
        }),

      setAuthEnabled: (enabled: boolean) =>
        set({ authEnabled: enabled }),

      logout: () =>
        set({
          token: null,
          user: null,
          expiresAt: null,
          isAuthenticated: false,
        }),

      isTokenExpired: () => {
        const { expiresAt } = get();
        if (!expiresAt) return true;
        return Date.now() / 1000 > expiresAt;
      },
    }),
    {
      name: 'outb-auth',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        expiresAt: state.expiresAt,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

// Selector hooks
export const useToken = () => useAuthStore((state) => state.token);
export const useUser = () => useAuthStore((state) => state.user);
export const useIsAuthenticated = () => useAuthStore((state) => state.isAuthenticated);
export const useAuthEnabled = () => useAuthStore((state) => state.authEnabled);
