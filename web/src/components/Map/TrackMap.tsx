// Track map component for displaying drone track history

import { useMemo } from 'react';
import { MapContainer, Polyline, CircleMarker, Popup } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import { LayerControl } from './LayerControl';
import type { TrackPoint } from '../../api/types';
import dayjs from 'dayjs';

interface TrackMapProps {
  trackPoints: TrackPoint[];
  deviceId?: string;
  center?: [number, number];
  zoom?: number;
}

export function TrackMap({
  trackPoints,
  deviceId,
  center = [39.9042, 116.4074],
  zoom = 14,
}: TrackMapProps) {
  // Calculate map center from track points
  const mapCenter = useMemo<[number, number]>(() => {
    if (trackPoints.length === 0) return center;

    const latSum = trackPoints.reduce((sum, p) => sum + p.lat, 0);
    const lonSum = trackPoints.reduce((sum, p) => sum + p.lon, 0);
    return [latSum / trackPoints.length, lonSum / trackPoints.length];
  }, [trackPoints, center]);

  // Convert track points to positions for polyline
  const positions = useMemo<[number, number][]>(() => {
    return trackPoints.map((p) => [p.lat, p.lon]);
  }, [trackPoints]);

  // Get start and end points
  const startPoint = trackPoints.length > 0 ? trackPoints[0] : null;
  const endPoint = trackPoints.length > 1 ? trackPoints[trackPoints.length - 1] : null;

  const formatTime = (timestamp: number) => {
    return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss');
  };

  return (
    <MapContainer
      center={mapCenter}
      zoom={zoom}
      className="w-full h-full rounded-lg"
      style={{ minHeight: '400px' }}
    >
      <LayerControl />

      {/* Track line */}
      {positions.length > 1 && (
        <Polyline
          positions={positions}
          pathOptions={{
            color: '#3b82f6',
            weight: 3,
            opacity: 0.8,
          }}
        />
      )}

      {/* Start point (green) */}
      {startPoint && (
        <CircleMarker
          center={[startPoint.lat, startPoint.lon]}
          radius={8}
          pathOptions={{
            fillColor: '#10b981',
            fillOpacity: 1,
            color: '#fff',
            weight: 2,
          }}
        >
          <Popup>
            <div className="text-sm">
              <p className="font-bold text-green-600">Start Point</p>
              {deviceId && <p>Device: {deviceId}</p>}
              <p>Time: {formatTime(startPoint.timestamp)}</p>
              <p>Position: {startPoint.lat.toFixed(6)}, {startPoint.lon.toFixed(6)}</p>
              <p>Altitude: {startPoint.alt.toFixed(1)}m</p>
            </div>
          </Popup>
        </CircleMarker>
      )}

      {/* End point (red) */}
      {endPoint && (
        <CircleMarker
          center={[endPoint.lat, endPoint.lon]}
          radius={8}
          pathOptions={{
            fillColor: '#ef4444',
            fillOpacity: 1,
            color: '#fff',
            weight: 2,
          }}
        >
          <Popup>
            <div className="text-sm">
              <p className="font-bold text-red-600">End Point</p>
              {deviceId && <p>Device: {deviceId}</p>}
              <p>Time: {formatTime(endPoint.timestamp)}</p>
              <p>Position: {endPoint.lat.toFixed(6)}, {endPoint.lon.toFixed(6)}</p>
              <p>Altitude: {endPoint.alt.toFixed(1)}m</p>
            </div>
          </Popup>
        </CircleMarker>
      )}
    </MapContainer>
  );
}

export default TrackMap;
