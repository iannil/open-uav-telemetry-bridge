// Main layout component

import { Outlet } from 'react-router-dom';
import { Header } from './Header';
import { Sidebar } from './Sidebar';
import { useWebSocket } from '../../hooks/useWebSocket';
import { useDronesData } from '../../hooks/useDrones';

export function MainLayout() {
  // Initialize WebSocket connection and initial data fetch
  useWebSocket({ autoConnect: true });
  useDronesData({ autoFetch: true, pollInterval: 30000 });

  return (
    <div className="flex flex-col h-screen bg-gray-900 text-white">
      <Header />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}

export default MainLayout;
