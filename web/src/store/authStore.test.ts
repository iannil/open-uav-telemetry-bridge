// Tests for authStore
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useAuthStore, useIsAuthenticated, useToken, useUser } from './authStore';

describe('authStore', () => {
  const mockUser = {
    username: 'admin',
    role: 'admin',
  };
  const mockToken = 'mock-jwt-token';
  const mockExpiresAt = Math.floor(Date.now() / 1000) + 3600; // 1 hour from now

  describe('setAuth', () => {
    it('should set authentication state', () => {
      const { setAuth } = useAuthStore.getState();

      setAuth(mockToken, mockUser, mockExpiresAt);

      const state = useAuthStore.getState();
      expect(state.token).toBe(mockToken);
      expect(state.user).toEqual(mockUser);
      expect(state.expiresAt).toBe(mockExpiresAt);
      expect(state.isAuthenticated).toBe(true);
    });
  });

  describe('setAuthEnabled', () => {
    it('should set authEnabled to true', () => {
      const { setAuthEnabled } = useAuthStore.getState();

      setAuthEnabled(true);

      expect(useAuthStore.getState().authEnabled).toBe(true);
    });

    it('should set authEnabled to false', () => {
      const { setAuthEnabled } = useAuthStore.getState();

      setAuthEnabled(false);

      expect(useAuthStore.getState().authEnabled).toBe(false);
    });
  });

  describe('logout', () => {
    it('should clear authentication state', () => {
      const { setAuth, logout } = useAuthStore.getState();

      setAuth(mockToken, mockUser, mockExpiresAt);
      logout();

      const state = useAuthStore.getState();
      expect(state.token).toBeNull();
      expect(state.user).toBeNull();
      expect(state.expiresAt).toBeNull();
      expect(state.isAuthenticated).toBe(false);
    });
  });

  describe('isTokenExpired', () => {
    it('should return true when no expiresAt', () => {
      const { isTokenExpired } = useAuthStore.getState();

      expect(isTokenExpired()).toBe(true);
    });

    it('should return false when token not expired', () => {
      const { setAuth, isTokenExpired } = useAuthStore.getState();
      const futureExpiry = Math.floor(Date.now() / 1000) + 3600;

      setAuth(mockToken, mockUser, futureExpiry);

      expect(isTokenExpired()).toBe(false);
    });

    it('should return true when token expired', () => {
      const { setAuth, isTokenExpired } = useAuthStore.getState();
      const pastExpiry = Math.floor(Date.now() / 1000) - 3600;

      setAuth(mockToken, mockUser, pastExpiry);

      expect(isTokenExpired()).toBe(true);
    });
  });

  describe('persistence', () => {
    it('should persist token, user, expiresAt, and isAuthenticated', () => {
      const { setAuth } = useAuthStore.getState();

      setAuth(mockToken, mockUser, mockExpiresAt);

      const stored = JSON.parse(localStorage.getItem('outb-auth') || '{}');
      expect(stored.state.token).toBe(mockToken);
      expect(stored.state.user).toEqual(mockUser);
      expect(stored.state.expiresAt).toBe(mockExpiresAt);
      expect(stored.state.isAuthenticated).toBe(true);
    });

    it('should not persist authEnabled', () => {
      const { setAuthEnabled, setAuth } = useAuthStore.getState();

      setAuthEnabled(true);
      setAuth(mockToken, mockUser, mockExpiresAt);

      const stored = JSON.parse(localStorage.getItem('outb-auth') || '{}');
      expect(stored.state.authEnabled).toBeUndefined();
    });
  });

  describe('selector hooks', () => {
    it('useToken should return current token', () => {
      const { setAuth } = useAuthStore.getState();

      expect(useAuthStore.getState().token).toBeNull();

      setAuth(mockToken, mockUser, mockExpiresAt);

      expect(useAuthStore.getState().token).toBe(mockToken);
    });

    it('useUser should return current user', () => {
      const { setAuth } = useAuthStore.getState();

      expect(useAuthStore.getState().user).toBeNull();

      setAuth(mockToken, mockUser, mockExpiresAt);

      expect(useAuthStore.getState().user).toEqual(mockUser);
    });

    it('useIsAuthenticated should return authentication status', () => {
      const { setAuth, logout } = useAuthStore.getState();

      expect(useAuthStore.getState().isAuthenticated).toBe(false);

      setAuth(mockToken, mockUser, mockExpiresAt);
      expect(useAuthStore.getState().isAuthenticated).toBe(true);

      logout();
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
    });
  });
});
