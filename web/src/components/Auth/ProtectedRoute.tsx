import { useEffect, useState } from 'react';
import { Navigate, useLocation, Outlet } from 'react-router-dom';
import { api } from '../../api/client';
import { useAuthStore, useIsAuthenticated, useAuthEnabled } from '../../store/authStore';

export function ProtectedRoute() {
  const location = useLocation();
  const [isChecking, setIsChecking] = useState(true);

  const isAuthenticated = useIsAuthenticated();
  const authEnabled = useAuthEnabled();
  const setAuthEnabled = useAuthStore((state) => state.setAuthEnabled);
  const logout = useAuthStore((state) => state.logout);
  const isTokenExpired = useAuthStore((state) => state.isTokenExpired);

  useEffect(() => {
    const checkAuth = async () => {
      // Check if token is expired
      if (isAuthenticated && isTokenExpired()) {
        logout();
        setIsChecking(false);
        return;
      }

      // If we already know auth status, no need to check again
      if (authEnabled !== null) {
        setIsChecking(false);
        return;
      }

      try {
        const response = await api.getMe();
        setAuthEnabled(response.auth_enabled);
      } catch {
        // If we get an error, assume auth is enabled
        setAuthEnabled(true);
      } finally {
        setIsChecking(false);
      }
    };

    checkAuth();
  }, [isAuthenticated, authEnabled]);

  if (isChecking) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-900">
        <div className="text-slate-400">Loading...</div>
      </div>
    );
  }

  // If auth is disabled, allow access
  if (authEnabled === false) {
    return <Outlet />;
  }

  // If auth is enabled and not authenticated, redirect to login
  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // Authenticated, render the protected content
  return <Outlet />;
}
