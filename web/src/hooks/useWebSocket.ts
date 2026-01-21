// WebSocket hook for real-time drone updates

import { useEffect, useRef, useCallback } from 'react';
import { useDroneStore } from '../store/droneStore';
import type { WSMessage, DroneState } from '../api/types';

const RECONNECT_INTERVAL = 3000;
const MAX_RECONNECT_ATTEMPTS = 10;

interface UseWebSocketOptions {
  autoConnect?: boolean;
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const { autoConnect = true } = options;
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<number | null>(null);

  const setDrone = useDroneStore((state) => state.setDrone);
  const removeDrone = useDroneStore((state) => state.removeDrone);
  const setConnected = useDroneStore((state) => state.setConnected);

  const getWebSocketUrl = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    return `${protocol}//${host}/api/v1/ws`;
  }, []);

  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message: WSMessage = JSON.parse(event.data);

        switch (message.type) {
          case 'state_update':
            if (message.data) {
              setDrone(message.data as DroneState);
            }
            break;
          case 'drone_online':
            // Drone came online, data might be included
            if (message.data) {
              setDrone(message.data as DroneState);
            }
            break;
          case 'drone_offline':
            if (message.device_id) {
              removeDrone(message.device_id);
            }
            break;
          default:
            console.warn('Unknown WebSocket message type:', message.type);
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    },
    [setDrone, removeDrone]
  );

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const url = getWebSocketUrl();
    console.log('[WebSocket] Connecting to', url);

    const ws = new WebSocket(url);

    ws.onopen = () => {
      console.log('[WebSocket] Connected');
      setConnected(true);
      reconnectAttemptsRef.current = 0;
    };

    ws.onmessage = handleMessage;

    ws.onclose = (event) => {
      console.log('[WebSocket] Disconnected:', event.code, event.reason);
      setConnected(false);
      wsRef.current = null;

      // Auto reconnect
      if (reconnectAttemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
        reconnectAttemptsRef.current++;
        console.log(
          `[WebSocket] Reconnecting in ${RECONNECT_INTERVAL}ms (attempt ${reconnectAttemptsRef.current}/${MAX_RECONNECT_ATTEMPTS})`
        );
        reconnectTimeoutRef.current = window.setTimeout(connect, RECONNECT_INTERVAL);
      } else {
        console.error('[WebSocket] Max reconnection attempts reached');
      }
    };

    ws.onerror = (error) => {
      console.error('[WebSocket] Error:', error);
    };

    wsRef.current = ws;
  }, [getWebSocketUrl, handleMessage, setConnected]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setConnected(false);
  }, [setConnected]);

  const reconnect = useCallback(() => {
    disconnect();
    reconnectAttemptsRef.current = 0;
    connect();
  }, [connect, disconnect]);

  useEffect(() => {
    if (autoConnect) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [autoConnect, connect, disconnect]);

  return {
    connect,
    disconnect,
    reconnect,
    isConnected: wsRef.current?.readyState === WebSocket.OPEN,
  };
}

export default useWebSocket;
