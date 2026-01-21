import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { MainLayout } from './components/Layout/MainLayout';
import { ProtectedRoute } from './components/Auth/ProtectedRoute';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { MapView } from './pages/MapView';
import { DroneList } from './pages/DroneList';
import { TrackHistory } from './pages/TrackHistory';
import { SystemStatus } from './pages/SystemStatus';
import { Settings } from './pages/Settings';
import { LogViewer } from './pages/LogViewer';
import { Alerts, Geofences } from './pages';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public route */}
        <Route path="/login" element={<Login />} />

        {/* Protected routes */}
        <Route element={<ProtectedRoute />}>
          <Route element={<MainLayout />}>
            <Route path="/" element={<Dashboard />} />
            <Route path="/map" element={<MapView />} />
            <Route path="/drones" element={<DroneList />} />
            <Route path="/tracks" element={<TrackHistory />} />
            <Route path="/status" element={<SystemStatus />} />
            <Route path="/logs" element={<LogViewer />} />
            <Route path="/alerts" element={<Alerts />} />
            <Route path="/geofences" element={<Geofences />} />
            <Route path="/settings" element={<Settings />} />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
