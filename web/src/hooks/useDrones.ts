// Hook for drone data fetching and management

import { useEffect, useCallback } from 'react';
import { useDroneStore, useDrones, useGatewayStatus } from '../store/droneStore';
import { api } from '../api/client';

interface UseDronesOptions {
  autoFetch?: boolean;
  pollInterval?: number;
}

export function useDronesData(options: UseDronesOptions = {}) {
  const { autoFetch = true, pollInterval = 0 } = options;

  const drones = useDrones();
  const gatewayStatus = useGatewayStatus();
  const setDrones = useDroneStore((state) => state.setDrones);
  const setGatewayStatus = useDroneStore((state) => state.setGatewayStatus);

  const fetchDrones = useCallback(async () => {
    try {
      const response = await api.getDrones();
      setDrones(response.drones);
    } catch (error) {
      console.error('Failed to fetch drones:', error);
    }
  }, [setDrones]);

  const fetchStatus = useCallback(async () => {
    try {
      const status = await api.getStatus();
      setGatewayStatus(status);
    } catch (error) {
      console.error('Failed to fetch status:', error);
    }
  }, [setGatewayStatus]);

  const fetchAll = useCallback(async () => {
    await Promise.all([fetchDrones(), fetchStatus()]);
  }, [fetchDrones, fetchStatus]);

  useEffect(() => {
    if (autoFetch) {
      fetchAll();
    }
  }, [autoFetch, fetchAll]);

  useEffect(() => {
    if (pollInterval > 0) {
      const interval = setInterval(fetchAll, pollInterval);
      return () => clearInterval(interval);
    }
  }, [pollInterval, fetchAll]);

  return {
    drones,
    gatewayStatus,
    fetchDrones,
    fetchStatus,
    fetchAll,
  };
}

export { useDrones, useGatewayStatus };
export default useDronesData;
