// Drone card component for list view

import type { DroneState } from '../../api/types';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';

dayjs.extend(relativeTime);

interface DroneCardProps {
  drone: DroneState;
  isSelected?: boolean;
  onClick?: () => void;
}

export function DroneCard({ drone, isSelected = false, onClick }: DroneCardProps) {
  const lastUpdate = dayjs(drone.timestamp);
  const isRecent = dayjs().diff(lastUpdate, 'second') < 10;

  const speed = Math.sqrt(
    drone.velocity.vx ** 2 + drone.velocity.vy ** 2
  ).toFixed(1);

  const getBatteryColor = (percent: number) => {
    if (percent > 50) return 'text-green-500';
    if (percent > 20) return 'text-yellow-500';
    return 'text-red-500';
  };

  const getStatusColor = (armed: boolean) => {
    return armed ? 'bg-red-500' : 'bg-green-500';
  };

  return (
    <div
      className={`p-4 rounded-lg border cursor-pointer transition-all ${
        isSelected
          ? 'bg-blue-900/50 border-blue-500'
          : 'bg-gray-800 border-gray-700 hover:border-gray-600'
      }`}
      onClick={onClick}
    >
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center space-x-2">
            <span
              className={`w-2 h-2 rounded-full ${
                isRecent ? 'bg-green-500' : 'bg-gray-500'
              }`}
            />
            <h3 className="font-semibold text-white">{drone.device_id}</h3>
          </div>
          <p className="text-sm text-gray-400 mt-1">{drone.protocol_source}</p>
        </div>
        <div className="flex items-center space-x-2">
          <span
            className={`px-2 py-0.5 rounded text-xs font-medium ${getStatusColor(
              drone.status.armed
            )} text-white`}
          >
            {drone.status.armed ? 'ARMED' : 'DISARMED'}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 mt-4 text-sm">
        <div>
          <p className="text-gray-500">Position</p>
          <p className="text-white">
            {drone.location.lat.toFixed(5)}, {drone.location.lon.toFixed(5)}
          </p>
        </div>
        <div>
          <p className="text-gray-500">Altitude</p>
          <p className="text-white">{drone.location.alt_gnss.toFixed(1)}m</p>
        </div>
        <div>
          <p className="text-gray-500">Speed</p>
          <p className="text-white">{speed} m/s</p>
        </div>
        <div>
          <p className="text-gray-500">Battery</p>
          <p className={getBatteryColor(drone.status.battery_percent)}>
            {drone.status.battery_percent}%
          </p>
        </div>
        <div>
          <p className="text-gray-500">Flight Mode</p>
          <p className="text-white">{drone.status.flight_mode}</p>
        </div>
        <div>
          <p className="text-gray-500">GPS</p>
          <p className="text-white">{drone.status.satellites_visible} sats</p>
        </div>
      </div>

      <div className="mt-4 pt-3 border-t border-gray-700">
        <p className="text-xs text-gray-500">
          Last update: {lastUpdate.fromNow()}
        </p>
      </div>
    </div>
  );
}

export default DroneCard;
