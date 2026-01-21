// Drone marker component for Leaflet map

import { Marker, Popup, Tooltip } from 'react-leaflet';
import L from 'leaflet';
import type { DroneState } from '../../api/types';

interface DroneMarkerProps {
  drone: DroneState;
  isSelected?: boolean;
  onClick?: (droneId: string) => void;
}

// Create a custom drone icon
function createDroneIcon(yaw: number, isSelected: boolean) {
  const color = isSelected ? '#3b82f6' : '#10b981';
  const size = isSelected ? 32 : 24;

  // SVG drone icon with rotation
  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="${color}" style="transform: rotate(${yaw}deg)">
      <path d="M12 2L4 20h16L12 2zm0 4l5 11H7l5-11z"/>
    </svg>
  `;

  return L.divIcon({
    className: 'drone-marker',
    html: svg,
    iconSize: [size, size],
    iconAnchor: [size / 2, size / 2],
  });
}

export function DroneMarker({ drone, isSelected = false, onClick }: DroneMarkerProps) {
  const position: [number, number] = [drone.location.lat, drone.location.lon];
  const icon = createDroneIcon(drone.attitude.yaw, isSelected);

  const formatAltitude = (alt: number) => `${alt.toFixed(1)}m`;
  const formatSpeed = (vx: number, vy: number) => {
    const speed = Math.sqrt(vx * vx + vy * vy);
    return `${speed.toFixed(1)}m/s`;
  };

  return (
    <Marker
      position={position}
      icon={icon}
      eventHandlers={{
        click: () => onClick?.(drone.device_id),
      }}
    >
      <Tooltip direction="top" offset={[0, -12]} permanent={isSelected}>
        <span className="font-medium">{drone.device_id}</span>
      </Tooltip>
      <Popup>
        <div className="text-sm">
          <h3 className="font-bold text-base mb-2">{drone.device_id}</h3>
          <div className="space-y-1">
            <p>
              <span className="text-gray-500">Protocol:</span>{' '}
              {drone.protocol_source}
            </p>
            <p>
              <span className="text-gray-500">Position:</span>{' '}
              {drone.location.lat.toFixed(6)}, {drone.location.lon.toFixed(6)}
            </p>
            <p>
              <span className="text-gray-500">Altitude:</span>{' '}
              {formatAltitude(drone.location.alt_gnss)}
            </p>
            <p>
              <span className="text-gray-500">Speed:</span>{' '}
              {formatSpeed(drone.velocity.vx, drone.velocity.vy)}
            </p>
            <p>
              <span className="text-gray-500">Battery:</span>{' '}
              {drone.status.battery_percent}%
            </p>
            <p>
              <span className="text-gray-500">Mode:</span>{' '}
              {drone.status.flight_mode}
            </p>
            <p>
              <span className="text-gray-500">GPS:</span>{' '}
              {drone.status.satellites_visible} sats
            </p>
          </div>
        </div>
      </Popup>
    </Marker>
  );
}

export default DroneMarker;
