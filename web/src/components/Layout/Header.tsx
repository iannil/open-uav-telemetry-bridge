// Header component

import { useIsConnected, useGatewayStatus } from '../../store/droneStore';

export function Header() {
  const isConnected = useIsConnected();
  const gatewayStatus = useGatewayStatus();

  return (
    <header className="bg-gray-800 border-b border-gray-700 px-4 py-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <h1 className="text-xl font-bold text-white">
            Open-UAV Telemetry Bridge
          </h1>
          {gatewayStatus && (
            <span className="text-sm text-gray-400">
              v{gatewayStatus.version}
            </span>
          )}
        </div>

        <div className="flex items-center space-x-4">
          {/* Connection Status */}
          <div className="flex items-center space-x-2">
            <span
              className={`w-2 h-2 rounded-full ${
                isConnected ? 'bg-green-500' : 'bg-red-500'
              }`}
            />
            <span className="text-sm text-gray-300">
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>

          {/* Stats */}
          {gatewayStatus && (
            <div className="text-sm text-gray-400">
              <span className="mr-4">
                Drones: {gatewayStatus.stats.active_drones}
              </span>
              <span>
                Clients: {gatewayStatus.stats.websocket_clients}
              </span>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}

export default Header;
