// Tests for droneStore
import { describe, it, expect, beforeEach } from 'vitest';
import { useDroneStore, useDrones, useDroneCount, useIsConnected } from './droneStore';
import { mockDroneState, mockDroneState2, mockGatewayStatus, createMockDrone } from '../test/mocks';

describe('droneStore', () => {
  describe('setDrone', () => {
    it('should add a new drone to the store', () => {
      const { setDrone } = useDroneStore.getState();

      setDrone(mockDroneState);

      const { drones } = useDroneStore.getState();
      expect(drones.size).toBe(1);
      expect(drones.get(mockDroneState.device_id)).toEqual(mockDroneState);
    });

    it('should update an existing drone', () => {
      const { setDrone } = useDroneStore.getState();

      setDrone(mockDroneState);
      const updatedDrone = {
        ...mockDroneState,
        status: { ...mockDroneState.status, battery_percent: 50 },
      };
      setDrone(updatedDrone);

      const { drones } = useDroneStore.getState();
      expect(drones.size).toBe(1);
      expect(drones.get(mockDroneState.device_id)?.status.battery_percent).toBe(50);
    });

    it('should update lastUpdate timestamp', () => {
      const { setDrone } = useDroneStore.getState();
      const beforeTime = Date.now();

      setDrone(mockDroneState);

      const { lastUpdate } = useDroneStore.getState();
      expect(lastUpdate).toBeGreaterThanOrEqual(beforeTime);
    });
  });

  describe('setDrones', () => {
    it('should replace all drones', () => {
      const { setDrone, setDrones } = useDroneStore.getState();

      setDrone(createMockDrone());
      setDrones([mockDroneState, mockDroneState2]);

      const { drones } = useDroneStore.getState();
      expect(drones.size).toBe(2);
      expect(drones.has(mockDroneState.device_id)).toBe(true);
      expect(drones.has(mockDroneState2.device_id)).toBe(true);
    });

    it('should handle empty array', () => {
      const { setDrone, setDrones } = useDroneStore.getState();

      setDrone(mockDroneState);
      setDrones([]);

      const { drones } = useDroneStore.getState();
      expect(drones.size).toBe(0);
    });
  });

  describe('removeDrone', () => {
    it('should remove a drone from the store', () => {
      const { setDrone, removeDrone } = useDroneStore.getState();

      setDrone(mockDroneState);
      removeDrone(mockDroneState.device_id);

      const { drones } = useDroneStore.getState();
      expect(drones.size).toBe(0);
    });

    it('should clear selection if removed drone was selected', () => {
      const { setDrone, selectDrone, removeDrone } = useDroneStore.getState();

      setDrone(mockDroneState);
      selectDrone(mockDroneState.device_id);
      removeDrone(mockDroneState.device_id);

      const { selectedDroneId } = useDroneStore.getState();
      expect(selectedDroneId).toBeNull();
    });

    it('should keep selection if different drone was removed', () => {
      const { setDrone, selectDrone, removeDrone } = useDroneStore.getState();

      setDrone(mockDroneState);
      setDrone(mockDroneState2);
      selectDrone(mockDroneState.device_id);
      removeDrone(mockDroneState2.device_id);

      const { selectedDroneId } = useDroneStore.getState();
      expect(selectedDroneId).toBe(mockDroneState.device_id);
    });
  });

  describe('selectDrone', () => {
    it('should set selectedDroneId', () => {
      const { selectDrone } = useDroneStore.getState();

      selectDrone('drone-001');

      const { selectedDroneId } = useDroneStore.getState();
      expect(selectedDroneId).toBe('drone-001');
    });

    it('should allow setting to null', () => {
      const { selectDrone } = useDroneStore.getState();

      selectDrone('drone-001');
      selectDrone(null);

      const { selectedDroneId } = useDroneStore.getState();
      expect(selectedDroneId).toBeNull();
    });
  });

  describe('setGatewayStatus', () => {
    it('should update gateway status', () => {
      const { setGatewayStatus } = useDroneStore.getState();

      setGatewayStatus(mockGatewayStatus);

      const { gatewayStatus } = useDroneStore.getState();
      expect(gatewayStatus).toEqual(mockGatewayStatus);
    });
  });

  describe('setConnected', () => {
    it('should update connection status', () => {
      const { setConnected } = useDroneStore.getState();

      setConnected(true);
      expect(useDroneStore.getState().isConnected).toBe(true);

      setConnected(false);
      expect(useDroneStore.getState().isConnected).toBe(false);
    });
  });

  describe('clear', () => {
    it('should reset store to initial state', () => {
      const { setDrone, selectDrone, setGatewayStatus, setConnected, clear } = useDroneStore.getState();

      setDrone(mockDroneState);
      selectDrone(mockDroneState.device_id);
      setGatewayStatus(mockGatewayStatus);
      setConnected(true);

      clear();

      const state = useDroneStore.getState();
      expect(state.drones.size).toBe(0);
      expect(state.selectedDroneId).toBeNull();
      expect(state.gatewayStatus).toBeNull();
      expect(state.isConnected).toBe(false);
    });
  });

  describe('selector hooks', () => {
    it('useDroneCount should return correct count', () => {
      const { setDrone } = useDroneStore.getState();

      expect(useDroneStore.getState().drones.size).toBe(0);

      setDrone(mockDroneState);
      setDrone(mockDroneState2);

      expect(useDroneStore.getState().drones.size).toBe(2);
    });
  });
});
