// Drone map component with Leaflet

import { useEffect, useState } from 'react';
import { MapContainer, useMap } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import { DroneMarker } from './DroneMarker';
import { LayerControl } from './LayerControl';
import { useDroneStore } from '../../store/droneStore';
import type { DroneState } from '../../api/types';

interface DroneMapProps {
  drones?: DroneState[];
  selectedDroneId?: string | null;
  onDroneClick?: (droneId: string) => void;
  center?: [number, number];
  zoom?: number;
  followSelected?: boolean;
}

// Component to handle map centering on selected drone
function MapController({
  selectedDrone,
  followSelected,
}: {
  selectedDrone: DroneState | null;
  followSelected: boolean;
}) {
  const map = useMap();

  useEffect(() => {
    if (followSelected && selectedDrone) {
      const pos: [number, number] = [
        selectedDrone.location.lat,
        selectedDrone.location.lon,
      ];
      map.setView(pos, map.getZoom(), { animate: true });
    }
  }, [map, selectedDrone, followSelected]);

  return null;
}

export function DroneMap({
  drones: propDrones,
  selectedDroneId: propSelectedId,
  onDroneClick,
  center = [39.9042, 116.4074], // Default to Beijing
  zoom = 10,
  followSelected = false,
}: DroneMapProps) {
  const storeDrones = useDroneStore((state) => Array.from(state.drones.values()));
  const storeSelectedId = useDroneStore((state) => state.selectedDroneId);
  const selectDrone = useDroneStore((state) => state.selectDrone);

  // Use props if provided, otherwise use store
  const drones = propDrones ?? storeDrones;
  const selectedDroneId = propSelectedId ?? storeSelectedId;

  const [mapCenter, setMapCenter] = useState<[number, number]>(center);

  // Update center based on first drone if no drones visible yet
  useEffect(() => {
    if (drones.length > 0 && mapCenter === center) {
      const firstDrone = drones[0];
      setMapCenter([firstDrone.location.lat, firstDrone.location.lon]);
    }
  }, [drones, mapCenter, center]);

  const handleDroneClick = (droneId: string) => {
    if (onDroneClick) {
      onDroneClick(droneId);
    } else {
      selectDrone(droneId === selectedDroneId ? null : droneId);
    }
  };

  const selectedDrone = selectedDroneId
    ? drones.find((d) => d.device_id === selectedDroneId) ?? null
    : null;

  return (
    <MapContainer
      center={mapCenter}
      zoom={zoom}
      className="w-full h-full rounded-lg"
      style={{ minHeight: '400px' }}
    >
      <LayerControl />
      <MapController
        selectedDrone={selectedDrone}
        followSelected={followSelected}
      />
      {drones.map((drone) => (
        <DroneMarker
          key={drone.device_id}
          drone={drone}
          isSelected={drone.device_id === selectedDroneId}
          onClick={handleDroneClick}
        />
      ))}
    </MapContainer>
  );
}

export default DroneMap;
