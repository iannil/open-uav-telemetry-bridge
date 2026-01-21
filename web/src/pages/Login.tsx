import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { api } from '../api/client';
import { useAuthStore, useIsAuthenticated, useAuthEnabled } from '../store/authStore';

export function Login() {
  const navigate = useNavigate();
  const location = useLocation();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isCheckingAuth, setIsCheckingAuth] = useState(true);

  const setAuth = useAuthStore((state) => state.setAuth);
  const setAuthEnabled = useAuthStore((state) => state.setAuthEnabled);
  const isAuthenticated = useIsAuthenticated();
  const authEnabled = useAuthEnabled();

  // Get the intended destination
  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/';

  // Check auth status on mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await api.getMe();
        setAuthEnabled(response.auth_enabled);

        if (!response.auth_enabled) {
          // Auth disabled, redirect to app
          navigate(from, { replace: true });
        } else if (isAuthenticated) {
          // Already authenticated, redirect
          navigate(from, { replace: true });
        }
      } catch {
        // If error, assume auth is enabled and show login
        setAuthEnabled(true);
      } finally {
        setIsCheckingAuth(false);
      }
    };

    checkAuth();
  }, []);

  // Redirect if already authenticated
  useEffect(() => {
    if (!isCheckingAuth && authEnabled === false) {
      navigate(from, { replace: true });
    }
    if (!isCheckingAuth && isAuthenticated) {
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, authEnabled, isCheckingAuth, navigate, from]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      const response = await api.login({ username, password });
      setAuth(response.token, response.user, response.expires_at);
      navigate(from, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setIsLoading(false);
    }
  };

  if (isCheckingAuth) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-900">
        <div className="text-slate-400">Checking authentication...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-900 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-white">OUTB</h1>
          <h2 className="mt-2 text-xl text-slate-400">
            Open UAV Telemetry Bridge
          </h2>
          <p className="mt-4 text-sm text-slate-500">
            Sign in to access the management console
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {error && (
            <div className="bg-red-500/10 border border-red-500/50 rounded-lg p-4 text-red-400 text-sm">
              {error}
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-slate-300">
                Username
              </label>
              <input
                id="username"
                name="username"
                type="text"
                autoComplete="username"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="mt-1 block w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter your username"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-slate-300">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 block w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter your password"
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full flex justify-center py-3 px-4 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isLoading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>

        <p className="text-center text-xs text-slate-600">
          Open-UAV-Telemetry-Bridge Management Console
        </p>
      </div>
    </div>
  );
}
