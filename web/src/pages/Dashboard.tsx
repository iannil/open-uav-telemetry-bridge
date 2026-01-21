// Dashboard page

import { useGatewayStatus } from '../store/droneStore';
import { useDrones } from '../hooks/useDrones';
import { DroneCard } from '../components/Drone/DroneCard';

export function Dashboard() {
  const gatewayStatus = useGatewayStatus();
  const drones = useDrones();

  const formatUptime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hours}h ${minutes}m ${secs}s`;
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-white">Dashboard</h1>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400">Active Drones</p>
              <p className="text-3xl font-bold text-white mt-1">
                {gatewayStatus?.stats.active_drones ?? drones.length}
              </p>
            </div>
            <div className="w-12 h-12 bg-blue-500/20 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 3h14a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15l-5 6h10l-5-6z" />
              </svg>
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400">WebSocket Clients</p>
              <p className="text-3xl font-bold text-white mt-1">
                {gatewayStatus?.stats.websocket_clients ?? 0}
              </p>
            </div>
            <div className="w-12 h-12 bg-green-500/20 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400">Uptime</p>
              <p className="text-xl font-bold text-white mt-1">
                {gatewayStatus ? formatUptime(gatewayStatus.uptime_seconds) : '-'}
              </p>
            </div>
            <div className="w-12 h-12 bg-purple-500/20 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400">Version</p>
              <p className="text-xl font-bold text-white mt-1">
                {gatewayStatus?.version ?? '-'}
              </p>
            </div>
            <div className="w-12 h-12 bg-orange-500/20 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      {/* Adapters & Publishers */}
      {gatewayStatus && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h2 className="text-lg font-semibold text-white mb-4">Adapters</h2>
            {gatewayStatus.adapters.length > 0 ? (
              <ul className="space-y-2">
                {gatewayStatus.adapters.map((adapter) => (
                  <li
                    key={adapter.name}
                    className="flex items-center justify-between py-2 px-3 bg-gray-900 rounded"
                  >
                    <span className="text-gray-300">{adapter.name}</span>
                    <span
                      className={`px-2 py-0.5 rounded text-xs font-medium ${
                        adapter.enabled
                          ? 'bg-green-500/20 text-green-400'
                          : 'bg-gray-600 text-gray-400'
                      }`}
                    >
                      {adapter.enabled ? 'Active' : 'Disabled'}
                    </span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-gray-500">No adapters configured</p>
            )}
          </div>

          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h2 className="text-lg font-semibold text-white mb-4">Publishers</h2>
            {gatewayStatus.publishers.length > 0 ? (
              <ul className="space-y-2">
                {gatewayStatus.publishers.map((publisher) => (
                  <li
                    key={publisher}
                    className="flex items-center py-2 px-3 bg-gray-900 rounded"
                  >
                    <span className="w-2 h-2 bg-green-500 rounded-full mr-3" />
                    <span className="text-gray-300">{publisher}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-gray-500">No publishers configured</p>
            )}
          </div>
        </div>
      )}

      {/* Recent Drones */}
      <div>
        <h2 className="text-lg font-semibold text-white mb-4">Active Drones</h2>
        {drones.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {drones.slice(0, 6).map((drone) => (
              <DroneCard key={drone.device_id} drone={drone} />
            ))}
          </div>
        ) : (
          <div className="bg-gray-800 rounded-lg p-8 border border-gray-700 text-center">
            <p className="text-gray-500">No drones connected</p>
            <p className="text-sm text-gray-600 mt-2">
              Connect a drone via MAVLink or DJI adapter to see it here
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

export default Dashboard;
