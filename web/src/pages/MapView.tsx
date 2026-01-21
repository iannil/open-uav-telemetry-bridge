// Map view page with real-time drone positions

import { useState } from 'react';
import { DroneMap } from '../components/Map/DroneMap';
import { DroneDetail } from '../components/Drone/DroneDetail';
import { useDroneStore } from '../store/droneStore';

export function MapView() {
  const drones = useDroneStore((state) => Array.from(state.drones.values()));
  const selectedDroneId = useDroneStore((state) => state.selectedDroneId);
  const selectDrone = useDroneStore((state) => state.selectDrone);

  const [followSelected, setFollowSelected] = useState(false);

  const selectedDrone = selectedDroneId
    ? drones.find((d) => d.device_id === selectedDroneId)
    : null;

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-white">Live Map</h1>
        <div className="flex items-center space-x-4">
          <label className="flex items-center space-x-2 text-sm text-gray-300">
            <input
              type="checkbox"
              checked={followSelected}
              onChange={(e) => setFollowSelected(e.target.checked)}
              className="rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-blue-500"
            />
            <span>Follow selected</span>
          </label>
          <span className="text-sm text-gray-500">
            {drones.length} drone{drones.length !== 1 ? 's' : ''} on map
          </span>
        </div>
      </div>

      <div className="flex-1 flex gap-4 min-h-0">
        {/* Map */}
        <div className={`${selectedDrone ? 'flex-1' : 'w-full'} min-h-[500px]`}>
          <DroneMap
            drones={drones}
            selectedDroneId={selectedDroneId}
            onDroneClick={(id) => selectDrone(id === selectedDroneId ? null : id)}
            followSelected={followSelected}
          />
        </div>

        {/* Detail Panel */}
        {selectedDrone && (
          <div className="w-96 overflow-auto">
            <DroneDetail
              drone={selectedDrone}
              onClose={() => selectDrone(null)}
            />
          </div>
        )}
      </div>
    </div>
  );
}

export default MapView;
