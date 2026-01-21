// System status page

import { useEffect, useState } from 'react';
import { useGatewayStatus } from '../store/droneStore';
import { api } from '../api/client';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';

dayjs.extend(relativeTime);

export function SystemStatus() {
  const gatewayStatus = useGatewayStatus();
  const [isHealthy, setIsHealthy] = useState<boolean | null>(null);
  const [lastCheck, setLastCheck] = useState<Date | null>(null);

  useEffect(() => {
    const checkHealth = async () => {
      const healthy = await api.checkHealth();
      setIsHealthy(healthy);
      setLastCheck(new Date());
    };

    checkHealth();
    const interval = setInterval(checkHealth, 10000);
    return () => clearInterval(interval);
  }, []);

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    const parts = [];
    if (days > 0) parts.push(`${days}d`);
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    parts.push(`${secs}s`);

    return parts.join(' ');
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-white">System Status</h1>

      {/* Health Status */}
      <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <div
              className={`w-4 h-4 rounded-full ${
                isHealthy === null
                  ? 'bg-gray-500'
                  : isHealthy
                  ? 'bg-green-500'
                  : 'bg-red-500'
              }`}
            />
            <div>
              <h2 className="text-lg font-semibold text-white">Gateway Health</h2>
              <p className="text-sm text-gray-400">
                {isHealthy === null
                  ? 'Checking...'
                  : isHealthy
                  ? 'Healthy - All systems operational'
                  : 'Unhealthy - Connection issues detected'}
              </p>
            </div>
          </div>
          {lastCheck && (
            <p className="text-sm text-gray-500">
              Last checked: {dayjs(lastCheck).fromNow()}
            </p>
          )}
        </div>
      </div>

      {/* Gateway Info */}
      {gatewayStatus && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
              <p className="text-sm text-gray-400">Version</p>
              <p className="text-2xl font-bold text-white mt-1">
                {gatewayStatus.version}
              </p>
            </div>
            <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
              <p className="text-sm text-gray-400">Uptime</p>
              <p className="text-2xl font-bold text-white mt-1">
                {formatUptime(gatewayStatus.uptime_seconds)}
              </p>
            </div>
            <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
              <p className="text-sm text-gray-400">Active Drones</p>
              <p className="text-2xl font-bold text-white mt-1">
                {gatewayStatus.stats.active_drones}
              </p>
            </div>
            <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
              <p className="text-sm text-gray-400">WebSocket Clients</p>
              <p className="text-2xl font-bold text-white mt-1">
                {gatewayStatus.stats.websocket_clients}
              </p>
            </div>
          </div>

          {/* Adapters */}
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h2 className="text-lg font-semibold text-white mb-4">
              Adapters (Data Sources)
            </h2>
            <p className="text-sm text-gray-400 mb-4">
              Adapters receive telemetry data from drones via various protocols.
            </p>
            {gatewayStatus.adapters.length > 0 ? (
              <div className="space-y-3">
                {gatewayStatus.adapters.map((adapter) => (
                  <div
                    key={adapter.name}
                    className="flex items-center justify-between p-4 bg-gray-900 rounded-lg"
                  >
                    <div className="flex items-center space-x-4">
                      <div
                        className={`w-3 h-3 rounded-full ${
                          adapter.enabled ? 'bg-green-500' : 'bg-gray-500'
                        }`}
                      />
                      <div>
                        <p className="font-medium text-white">{adapter.name}</p>
                        <p className="text-sm text-gray-400">
                          {adapter.name.toLowerCase().includes('mavlink')
                            ? 'MAVLink Protocol (UDP/TCP/Serial)'
                            : adapter.name.toLowerCase().includes('dji')
                            ? 'DJI Mobile SDK Forwarder'
                            : 'Custom Adapter'}
                        </p>
                      </div>
                    </div>
                    <span
                      className={`px-3 py-1 rounded-full text-sm ${
                        adapter.enabled
                          ? 'bg-green-500/20 text-green-400'
                          : 'bg-gray-600 text-gray-400'
                      }`}
                    >
                      {adapter.enabled ? 'Active' : 'Disabled'}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-gray-500">No adapters configured</p>
            )}
          </div>

          {/* Publishers */}
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h2 className="text-lg font-semibold text-white mb-4">
              Publishers (Data Outputs)
            </h2>
            <p className="text-sm text-gray-400 mb-4">
              Publishers send processed telemetry data to external systems.
            </p>
            {gatewayStatus.publishers.length > 0 ? (
              <div className="space-y-3">
                {gatewayStatus.publishers.map((publisher) => (
                  <div
                    key={publisher}
                    className="flex items-center justify-between p-4 bg-gray-900 rounded-lg"
                  >
                    <div className="flex items-center space-x-4">
                      <div className="w-3 h-3 rounded-full bg-green-500" />
                      <div>
                        <p className="font-medium text-white">{publisher}</p>
                        <p className="text-sm text-gray-400">
                          {publisher.toLowerCase().includes('mqtt')
                            ? 'MQTT Message Broker'
                            : publisher.toLowerCase().includes('gb28181')
                            ? 'GB/T 28181 National Standard (SIP)'
                            : 'Custom Publisher'}
                        </p>
                      </div>
                    </div>
                    <span className="px-3 py-1 rounded-full text-sm bg-green-500/20 text-green-400">
                      Active
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-gray-500">No publishers configured</p>
            )}
          </div>

          {/* API Endpoints */}
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h2 className="text-lg font-semibold text-white mb-4">
              API Endpoints
            </h2>
            <div className="space-y-2">
              <div className="flex items-center justify-between p-3 bg-gray-900 rounded">
                <code className="text-blue-400">GET /health</code>
                <span className="text-gray-400 text-sm">Health check</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-900 rounded">
                <code className="text-blue-400">GET /api/v1/status</code>
                <span className="text-gray-400 text-sm">Gateway status</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-900 rounded">
                <code className="text-blue-400">GET /api/v1/drones</code>
                <span className="text-gray-400 text-sm">List all drones</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-900 rounded">
                <code className="text-blue-400">GET /api/v1/drones/:id</code>
                <span className="text-gray-400 text-sm">Get drone details</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-900 rounded">
                <code className="text-blue-400">GET /api/v1/drones/:id/track</code>
                <span className="text-gray-400 text-sm">Get drone track</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-900 rounded">
                <code className="text-green-400">WS /api/v1/ws</code>
                <span className="text-gray-400 text-sm">Real-time updates</span>
              </div>
            </div>
          </div>
        </>
      )}

      {!gatewayStatus && (
        <div className="bg-gray-800 rounded-lg p-8 border border-gray-700 text-center">
          <p className="text-gray-500">Loading gateway status...</p>
        </div>
      )}
    </div>
  );
}

export default SystemStatus;
