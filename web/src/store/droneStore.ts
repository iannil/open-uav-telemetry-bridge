// Zustand store for drone state management

import { create } from 'zustand';
import type { DroneState, StatusResponse } from '../api/types';

interface DroneStore {
  // State
  drones: Map<string, DroneState>;
  selectedDroneId: string | null;
  gatewayStatus: StatusResponse | null;
  isConnected: boolean;
  lastUpdate: number;

  // Actions
  setDrone: (state: DroneState) => void;
  setDrones: (states: DroneState[]) => void;
  removeDrone: (deviceId: string) => void;
  selectDrone: (deviceId: string | null) => void;
  setGatewayStatus: (status: StatusResponse) => void;
  setConnected: (connected: boolean) => void;
  clear: () => void;
}

export const useDroneStore = create<DroneStore>((set) => ({
  // Initial state
  drones: new Map(),
  selectedDroneId: null,
  gatewayStatus: null,
  isConnected: false,
  lastUpdate: 0,

  // Actions
  setDrone: (state: DroneState) =>
    set((prev) => {
      const newDrones = new Map(prev.drones);
      newDrones.set(state.device_id, state);
      return { drones: newDrones, lastUpdate: Date.now() };
    }),

  setDrones: (states: DroneState[]) =>
    set(() => {
      const newDrones = new Map<string, DroneState>();
      states.forEach((state) => newDrones.set(state.device_id, state));
      return { drones: newDrones, lastUpdate: Date.now() };
    }),

  removeDrone: (deviceId: string) =>
    set((prev) => {
      const newDrones = new Map(prev.drones);
      newDrones.delete(deviceId);
      return {
        drones: newDrones,
        selectedDroneId: prev.selectedDroneId === deviceId ? null : prev.selectedDroneId,
      };
    }),

  selectDrone: (deviceId: string | null) =>
    set({ selectedDroneId: deviceId }),

  setGatewayStatus: (status: StatusResponse) =>
    set({ gatewayStatus: status }),

  setConnected: (connected: boolean) =>
    set({ isConnected: connected }),

  clear: () =>
    set({
      drones: new Map(),
      selectedDroneId: null,
      gatewayStatus: null,
      isConnected: false,
    }),
}));

// Selector hooks for convenience
export const useDrones = () => useDroneStore((state) => Array.from(state.drones.values()));
export const useSelectedDrone = () => {
  const drones = useDroneStore((state) => state.drones);
  const selectedId = useDroneStore((state) => state.selectedDroneId);
  return selectedId ? drones.get(selectedId) : null;
};
export const useDroneById = (id: string) => useDroneStore((state) => state.drones.get(id));
export const useDroneCount = () => useDroneStore((state) => state.drones.size);
export const useIsConnected = () => useDroneStore((state) => state.isConnected);
export const useGatewayStatus = () => useDroneStore((state) => state.gatewayStatus);
