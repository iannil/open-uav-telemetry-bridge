// Drone detail panel component

import type { DroneState } from '../../api/types';
import dayjs from 'dayjs';

interface DroneDetailProps {
  drone: DroneState;
  onClose?: () => void;
}

export function DroneDetail({ drone, onClose }: DroneDetailProps) {
  const formatTime = (timestamp: number) => {
    return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  const speed = Math.sqrt(
    drone.velocity.vx ** 2 + drone.velocity.vy ** 2 + drone.velocity.vz ** 2
  );

  const direction = Math.atan2(drone.velocity.vy, drone.velocity.vx) * (180 / Math.PI);

  return (
    <div className="bg-gray-800 rounded-lg border border-gray-700 p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-white">{drone.device_id}</h2>
        {onClose && (
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>

      {/* Status Badge */}
      <div className="flex items-center space-x-3 mb-6">
        <span
          className={`px-3 py-1 rounded-full text-sm font-medium ${
            drone.status.armed
              ? 'bg-red-500/20 text-red-400'
              : 'bg-green-500/20 text-green-400'
          }`}
        >
          {drone.status.armed ? 'Armed' : 'Disarmed'}
        </span>
        <span className="px-3 py-1 rounded-full text-sm font-medium bg-blue-500/20 text-blue-400">
          {drone.status.flight_mode}
        </span>
        <span className="px-3 py-1 rounded-full text-sm font-medium bg-gray-600 text-gray-300">
          {drone.protocol_source}
        </span>
      </div>

      {/* Location Section */}
      <section className="mb-6">
        <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">
          Location
        </h3>
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Latitude</p>
            <p className="text-lg font-mono text-white">{drone.location.lat.toFixed(6)}</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Longitude</p>
            <p className="text-lg font-mono text-white">{drone.location.lon.toFixed(6)}</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Altitude (GNSS)</p>
            <p className="text-lg font-mono text-white">{drone.location.alt_gnss.toFixed(1)} m</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Altitude (Baro)</p>
            <p className="text-lg font-mono text-white">{drone.location.alt_baro.toFixed(1)} m</p>
          </div>
        </div>
        {drone.location.gcj02 && (
          <div className="mt-3 text-xs text-gray-500">
            GCJ02: {drone.location.gcj02.lat.toFixed(6)}, {drone.location.gcj02.lon.toFixed(6)}
          </div>
        )}
      </section>

      {/* Attitude Section */}
      <section className="mb-6">
        <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">
          Attitude
        </h3>
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Roll</p>
            <p className="text-lg font-mono text-white">{drone.attitude.roll.toFixed(1)}</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Pitch</p>
            <p className="text-lg font-mono text-white">{drone.attitude.pitch.toFixed(1)}</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Yaw</p>
            <p className="text-lg font-mono text-white">{drone.attitude.yaw.toFixed(1)}</p>
          </div>
        </div>
      </section>

      {/* Velocity Section */}
      <section className="mb-6">
        <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">
          Velocity
        </h3>
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Ground Speed</p>
            <p className="text-lg font-mono text-white">{speed.toFixed(2)} m/s</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Vertical Speed</p>
            <p className="text-lg font-mono text-white">{drone.velocity.vz.toFixed(2)} m/s</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Direction</p>
            <p className="text-lg font-mono text-white">{direction.toFixed(1)}</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Vx / Vy / Vz</p>
            <p className="text-sm font-mono text-white">
              {drone.velocity.vx.toFixed(1)} / {drone.velocity.vy.toFixed(1)} / {drone.velocity.vz.toFixed(1)}
            </p>
          </div>
        </div>
      </section>

      {/* Status Section */}
      <section className="mb-6">
        <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">
          System Status
        </h3>
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Battery</p>
            <p className={`text-lg font-mono ${
              drone.status.battery_percent > 50 ? 'text-green-400' :
              drone.status.battery_percent > 20 ? 'text-yellow-400' : 'text-red-400'
            }`}>
              {drone.status.battery_percent}%
            </p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Voltage</p>
            <p className="text-lg font-mono text-white">{drone.status.battery_voltage.toFixed(1)} V</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">GPS Fix</p>
            <p className="text-lg font-mono text-white">{drone.status.gps_fix_type}</p>
          </div>
          <div className="bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-500">Satellites</p>
            <p className="text-lg font-mono text-white">{drone.status.satellites_visible}</p>
          </div>
        </div>
      </section>

      {/* Timestamp */}
      <div className="pt-4 border-t border-gray-700">
        <p className="text-xs text-gray-500">
          Last Update: {formatTime(drone.timestamp)}
        </p>
      </div>
    </div>
  );
}

export default DroneDetail;
