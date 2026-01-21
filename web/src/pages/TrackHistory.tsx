// Track history page

import { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { TrackMap } from '../components/Map/TrackMap';
import { useDroneStore } from '../store/droneStore';
import { api } from '../api/client';
import type { TrackPoint } from '../api/types';
import dayjs from 'dayjs';

export function TrackHistory() {
  const [searchParams, setSearchParams] = useSearchParams();
  const drones = useDroneStore((state) => Array.from(state.drones.values()));

  const [selectedDroneId, setSelectedDroneId] = useState<string>(
    searchParams.get('drone') || ''
  );
  const [trackPoints, setTrackPoints] = useState<TrackPoint[]>([]);
  const [totalSize, setTotalSize] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [limit, setLimit] = useState(1000);

  const fetchTrack = useCallback(async () => {
    if (!selectedDroneId) {
      setTrackPoints([]);
      setTotalSize(0);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await api.getTrack(selectedDroneId, limit);
      setTrackPoints(response.points);
      setTotalSize(response.total_size);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch track');
      setTrackPoints([]);
    } finally {
      setLoading(false);
    }
  }, [selectedDroneId, limit]);

  useEffect(() => {
    fetchTrack();
  }, [fetchTrack]);

  const handleDroneChange = (droneId: string) => {
    setSelectedDroneId(droneId);
    if (droneId) {
      setSearchParams({ drone: droneId });
    } else {
      setSearchParams({});
    }
  };

  const handleClearTrack = async () => {
    if (!selectedDroneId) return;

    if (!window.confirm(`Are you sure you want to clear the track for ${selectedDroneId}?`)) {
      return;
    }

    try {
      await api.clearTrack(selectedDroneId);
      setTrackPoints([]);
      setTotalSize(0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to clear track');
    }
  };

  const formatDuration = () => {
    if (trackPoints.length < 2) return '-';
    const start = trackPoints[0].timestamp;
    const end = trackPoints[trackPoints.length - 1].timestamp;
    const duration = end - start;
    const minutes = Math.floor(duration / 60000);
    const seconds = Math.floor((duration % 60000) / 1000);
    return `${minutes}m ${seconds}s`;
  };

  const calculateDistance = () => {
    if (trackPoints.length < 2) return 0;

    let distance = 0;
    for (let i = 1; i < trackPoints.length; i++) {
      const prev = trackPoints[i - 1];
      const curr = trackPoints[i];
      distance += haversineDistance(prev.lat, prev.lon, curr.lat, curr.lon);
    }
    return distance;
  };

  // Haversine formula to calculate distance between two points
  const haversineDistance = (lat1: number, lon1: number, lat2: number, lon2: number) => {
    const R = 6371000; // Earth's radius in meters
    const dLat = ((lat2 - lat1) * Math.PI) / 180;
    const dLon = ((lon2 - lon1) * Math.PI) / 180;
    const a =
      Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos((lat1 * Math.PI) / 180) *
        Math.cos((lat2 * Math.PI) / 180) *
        Math.sin(dLon / 2) *
        Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  };

  const distance = calculateDistance();

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-white">Track History</h1>
      </div>

      {/* Controls */}
      <div className="flex items-center gap-4 mb-4">
        <div className="flex-1">
          <select
            value={selectedDroneId}
            onChange={(e) => handleDroneChange(e.target.value)}
            className="w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-blue-500"
          >
            <option value="">Select a drone...</option>
            {drones.map((drone) => (
              <option key={drone.device_id} value={drone.device_id}>
                {drone.device_id} ({drone.protocol_source})
              </option>
            ))}
          </select>
        </div>
        <div>
          <select
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value))}
            className="px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-blue-500"
          >
            <option value={100}>Last 100 points</option>
            <option value={500}>Last 500 points</option>
            <option value={1000}>Last 1000 points</option>
            <option value={5000}>Last 5000 points</option>
            <option value={0}>All points</option>
          </select>
        </div>
        <button
          onClick={fetchTrack}
          disabled={!selectedDroneId || loading}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed rounded-lg text-white"
        >
          {loading ? 'Loading...' : 'Refresh'}
        </button>
        <button
          onClick={handleClearTrack}
          disabled={!selectedDroneId || trackPoints.length === 0}
          className="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-700 disabled:cursor-not-allowed rounded-lg text-white"
        >
          Clear Track
        </button>
      </div>

      {/* Stats */}
      {selectedDroneId && trackPoints.length > 0 && (
        <div className="grid grid-cols-4 gap-4 mb-4">
          <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
            <p className="text-sm text-gray-400">Total Points</p>
            <p className="text-xl font-bold text-white">
              {trackPoints.length} / {totalSize}
            </p>
          </div>
          <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
            <p className="text-sm text-gray-400">Duration</p>
            <p className="text-xl font-bold text-white">{formatDuration()}</p>
          </div>
          <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
            <p className="text-sm text-gray-400">Distance</p>
            <p className="text-xl font-bold text-white">
              {distance > 1000
                ? `${(distance / 1000).toFixed(2)} km`
                : `${distance.toFixed(0)} m`}
            </p>
          </div>
          <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
            <p className="text-sm text-gray-400">Time Range</p>
            <p className="text-sm text-white">
              {trackPoints.length > 0 && (
                <>
                  {dayjs(trackPoints[0].timestamp).format('HH:mm:ss')} -{' '}
                  {dayjs(trackPoints[trackPoints.length - 1].timestamp).format('HH:mm:ss')}
                </>
              )}
            </p>
          </div>
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="mb-4 p-4 bg-red-900/50 border border-red-700 rounded-lg text-red-400">
          {error}
        </div>
      )}

      {/* Map */}
      <div className="flex-1 min-h-[400px]">
        {selectedDroneId ? (
          trackPoints.length > 0 ? (
            <TrackMap trackPoints={trackPoints} deviceId={selectedDroneId} />
          ) : (
            <div className="h-full flex items-center justify-center bg-gray-800 rounded-lg border border-gray-700">
              <div className="text-center text-gray-500">
                {loading ? (
                  <p>Loading track data...</p>
                ) : (
                  <>
                    <p>No track data available</p>
                    <p className="text-sm mt-2">
                      Track data will appear when the drone moves
                    </p>
                  </>
                )}
              </div>
            </div>
          )
        ) : (
          <div className="h-full flex items-center justify-center bg-gray-800 rounded-lg border border-gray-700">
            <div className="text-center text-gray-500">
              <p>Select a drone to view its track history</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default TrackHistory;
