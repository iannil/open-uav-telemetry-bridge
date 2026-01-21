// Geofences page - Geofence management with map visualization

import { useEffect, useState, useRef } from 'react';
import { MapContainer, Polygon, Circle, useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import 'leaflet-draw/dist/leaflet.draw.css';
import 'leaflet-draw';
import { LayerControl } from '../../components/Map';
import {
  useGeofenceStore,
  useGeofences,
  useGeofenceBreaches,
  useGeofenceStats,
  useGeofencesLoading,
  useGeofencesError,
} from '../../store/geofenceStore';
import type { Geofence, GeofenceType } from '../../api/types';

type TabType = 'map' | 'list' | 'breaches';

export default function GeofencesPage() {
  const [activeTab, setActiveTab] = useState<TabType>('map');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingGeofence, setEditingGeofence] = useState<Geofence | null>(null);
  const [drawingMode, setDrawingMode] = useState<'polygon' | 'circle' | null>(null);
  const [drawnCoordinates, setDrawnCoordinates] = useState<number[][]>([]);
  const [drawnCenter, setDrawnCenter] = useState<number[] | null>(null);
  const [drawnRadius, setDrawnRadius] = useState<number>(0);

  const geofences = useGeofences();
  const breaches = useGeofenceBreaches();
  const stats = useGeofenceStats();
  const isLoading = useGeofencesLoading();
  const error = useGeofencesError();

  const {
    fetchGeofences,
    fetchBreaches,
    fetchStats,
    createGeofence,
    updateGeofence,
    deleteGeofence,
    clearBreaches,
  } = useGeofenceStore();

  useEffect(() => {
    fetchGeofences();
    fetchBreaches();
    fetchStats();
  }, [fetchGeofences, fetchBreaches, fetchStats]);

  const handleEdit = (geofence: Geofence) => {
    setEditingGeofence(geofence);
    setShowCreateModal(true);
  };

  const handleDelete = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this geofence?')) {
      await deleteGeofence(id);
    }
  };

  const handleCreate = () => {
    setEditingGeofence(null);
    setDrawnCoordinates([]);
    setDrawnCenter(null);
    setDrawnRadius(0);
    setShowCreateModal(true);
  };

  const handleSave = async (geofenceData: Partial<Geofence>) => {
    if (editingGeofence) {
      await updateGeofence(editingGeofence.id, geofenceData);
    } else {
      await createGeofence(geofenceData);
    }
    setShowCreateModal(false);
    setEditingGeofence(null);
    setDrawnCoordinates([]);
    setDrawnCenter(null);
    setDrawnRadius(0);
    setDrawingMode(null);
  };

  const handleDrawComplete = (type: 'polygon' | 'circle', coords: number[][], center?: number[], radius?: number) => {
    if (type === 'polygon') {
      setDrawnCoordinates(coords);
    } else {
      setDrawnCenter(center || null);
      setDrawnRadius(radius || 0);
    }
    setDrawingMode(null);
    setShowCreateModal(true);
  };

  const formatTimestamp = (ts: number) => {
    return new Date(ts).toLocaleString();
  };

  return (
    <div className="p-6 h-full flex flex-col">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Geofences</h1>
        <p className="text-gray-600 dark:text-gray-400">Manage geographic boundaries and breach alerts</p>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <div className="text-sm text-gray-500 dark:text-gray-400">Total Geofences</div>
            <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.total_geofences}</div>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <div className="text-sm text-gray-500 dark:text-gray-400">Enabled</div>
            <div className="text-2xl font-bold text-green-600 dark:text-green-400">{stats.enabled_geofences}</div>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <div className="text-sm text-gray-500 dark:text-gray-400">Total Breaches</div>
            <div className="text-2xl font-bold text-red-600 dark:text-red-400">{stats.total_breaches}</div>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <div className="text-sm text-gray-500 dark:text-gray-400">Tracked Devices</div>
            <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.tracked_devices}</div>
          </div>
        </div>
      )}

      {/* Error display */}
      {error && (
        <div className="mb-4 p-4 bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300 rounded-lg">
          {error}
        </div>
      )}

      {/* Tabs */}
      <div className="mb-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('map')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'map'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'
            }`}
          >
            Map View
          </button>
          <button
            onClick={() => setActiveTab('list')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'list'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'
            }`}
          >
            List View
          </button>
          <button
            onClick={() => setActiveTab('breaches')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'breaches'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'
            }`}
          >
            Breaches
          </button>
        </nav>
        <div className="flex gap-2">
          {activeTab === 'map' && (
            <>
              <button
                onClick={() => setDrawingMode('polygon')}
                className={`px-3 py-1 text-sm rounded ${
                  drawingMode === 'polygon'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
                }`}
              >
                Draw Polygon
              </button>
              <button
                onClick={() => setDrawingMode('circle')}
                className={`px-3 py-1 text-sm rounded ${
                  drawingMode === 'circle'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
                }`}
              >
                Draw Circle
              </button>
            </>
          )}
          <button
            onClick={handleCreate}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Create Geofence
          </button>
        </div>
      </div>

      {/* Map Tab */}
      {activeTab === 'map' && (
        <div className="flex-1 bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden" style={{ minHeight: '500px' }}>
          <GeofenceMap
            geofences={geofences}
            drawingMode={drawingMode}
            onDrawComplete={handleDrawComplete}
            onCancelDraw={() => setDrawingMode(null)}
          />
        </div>
      )}

      {/* List Tab */}
      {activeTab === 'list' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          {isLoading ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">Loading...</div>
          ) : geofences.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">No geofences configured</div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-900">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                    Type
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                    Alerts
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {geofences.map((gf) => (
                  <tr key={gf.id}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      {gf.name}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {gf.type === 'polygon' ? 'Polygon' : `Circle (${gf.radius}m)`}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {gf.alert_on_enter && <span className="mr-2">Enter</span>}
                      {gf.alert_on_exit && <span>Exit</span>}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded ${
                          gf.enabled
                            ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
                            : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
                        }`}
                      >
                        {gf.enabled ? 'Enabled' : 'Disabled'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => handleEdit(gf)}
                        className="text-blue-600 hover:text-blue-900 dark:text-blue-400 mr-3"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(gf.id)}
                        className="text-red-600 hover:text-red-900 dark:text-red-400"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Breaches Tab */}
      {activeTab === 'breaches' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <div className="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-end">
            <button
              onClick={() => clearBreaches()}
              className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
            >
              Clear Breaches
            </button>
          </div>
          {breaches.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">No breaches recorded</div>
          ) : (
            <div className="divide-y divide-gray-200 dark:divide-gray-700">
              {breaches.map((breach) => {
                const gf = geofences.find((g) => g.id === breach.geofence_id);
                return (
                  <div key={breach.id} className="p-4">
                    <div className="flex items-center gap-2 mb-1">
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded ${
                          breach.type === 'enter'
                            ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
                            : 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
                        }`}
                      >
                        {breach.type === 'enter' ? 'ENTERED' : 'EXITED'}
                      </span>
                      <span className="text-sm text-gray-900 dark:text-white font-medium">
                        {gf?.name || breach.geofence_id}
                      </span>
                    </div>
                    <div className="text-sm text-gray-500 dark:text-gray-400">
                      <span>Device: {breach.device_id}</span>
                      <span className="mx-2">|</span>
                      <span>
                        Position: {breach.lat.toFixed(6)}, {breach.lon.toFixed(6)} @ {breach.alt.toFixed(1)}m
                      </span>
                      <span className="mx-2">|</span>
                      <span>{formatTimestamp(breach.timestamp)}</span>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* Create/Edit Modal */}
      {showCreateModal && (
        <GeofenceModal
          geofence={editingGeofence}
          drawnCoordinates={drawnCoordinates}
          drawnCenter={drawnCenter}
          drawnRadius={drawnRadius}
          onSave={handleSave}
          onClose={() => {
            setShowCreateModal(false);
            setEditingGeofence(null);
            setDrawnCoordinates([]);
            setDrawnCenter(null);
            setDrawnRadius(0);
          }}
        />
      )}
    </div>
  );
}

// Geofence Map Component
interface GeofenceMapProps {
  geofences: Geofence[];
  drawingMode: 'polygon' | 'circle' | null;
  onDrawComplete: (type: 'polygon' | 'circle', coords: number[][], center?: number[], radius?: number) => void;
  onCancelDraw: () => void;
}

function GeofenceMap({ geofences, drawingMode, onDrawComplete, onCancelDraw }: GeofenceMapProps) {
  return (
    <MapContainer
      center={[39.9042, 116.4074]}
      zoom={10}
      className="w-full h-full"
      style={{ minHeight: '500px' }}
    >
      <LayerControl />
      <DrawControl
        drawingMode={drawingMode}
        onDrawComplete={onDrawComplete}
        onCancelDraw={onCancelDraw}
      />
      {geofences.map((gf) => (
        <GeofenceLayer key={gf.id} geofence={gf} />
      ))}
    </MapContainer>
  );
}

// Draw Control Component
interface DrawControlProps {
  drawingMode: 'polygon' | 'circle' | null;
  onDrawComplete: (type: 'polygon' | 'circle', coords: number[][], center?: number[], radius?: number) => void;
  onCancelDraw: () => void;
}

function DrawControl({ drawingMode, onDrawComplete }: DrawControlProps) {
  const map = useMap();
  const drawControlRef = useRef<L.Control.Draw | null>(null);
  const drawnItemsRef = useRef<L.FeatureGroup>(new L.FeatureGroup());

  useEffect(() => {
    map.addLayer(drawnItemsRef.current);

    return () => {
      map.removeLayer(drawnItemsRef.current);
    };
  }, [map]);

  useEffect(() => {
    if (drawingMode) {
      // Create draw control
      const drawControl = new L.Control.Draw({
        position: 'topright',
        draw: {
          polygon: drawingMode === 'polygon' ? { shapeOptions: { color: '#3B82F6' } } : false,
          circle: drawingMode === 'circle' ? { shapeOptions: { color: '#3B82F6' } } : false,
          polyline: false,
          rectangle: false,
          marker: false,
          circlemarker: false,
        },
        edit: {
          featureGroup: drawnItemsRef.current,
          remove: false,
          edit: false,
        },
      });

      map.addControl(drawControl);
      drawControlRef.current = drawControl;

      // Handle draw complete
      const handleCreated = (e: L.LeafletEvent) => {
        const event = e as L.DrawEvents.Created;
        const layer = event.layer;

        if (event.layerType === 'polygon') {
          const polygon = layer as L.Polygon;
          const latlngs = polygon.getLatLngs()[0] as L.LatLng[];
          const coords = latlngs.map((ll) => [ll.lat, ll.lng]);
          onDrawComplete('polygon', coords);
        } else if (event.layerType === 'circle') {
          const circle = layer as L.Circle;
          const center = circle.getLatLng();
          const radius = circle.getRadius();
          onDrawComplete('circle', [], [center.lat, center.lng], radius);
        }
      };

      map.on(L.Draw.Event.CREATED, handleCreated);

      return () => {
        map.removeControl(drawControl);
        map.off(L.Draw.Event.CREATED, handleCreated);
      };
    }
  }, [map, drawingMode, onDrawComplete]);

  return null;
}

// Geofence Layer Component
interface GeofenceLayerProps {
  geofence: Geofence;
}

function GeofenceLayer({ geofence }: GeofenceLayerProps) {
  const color = geofence.enabled ? '#3B82F6' : '#9CA3AF';

  if (geofence.type === 'polygon' && geofence.coordinates) {
    const positions = geofence.coordinates.map((coord) => [coord[0], coord[1]] as [number, number]);
    return (
      <Polygon
        positions={positions}
        pathOptions={{ color, fillColor: color, fillOpacity: 0.2 }}
      />
    );
  }

  if (geofence.type === 'circle' && geofence.center && geofence.radius) {
    return (
      <Circle
        center={[geofence.center[0], geofence.center[1]]}
        radius={geofence.radius}
        pathOptions={{ color, fillColor: color, fillOpacity: 0.2 }}
      />
    );
  }

  return null;
}

// Geofence Modal Component
interface GeofenceModalProps {
  geofence: Geofence | null;
  drawnCoordinates: number[][];
  drawnCenter: number[] | null;
  drawnRadius: number;
  onSave: (geofence: Partial<Geofence>) => void;
  onClose: () => void;
}

function GeofenceModal({
  geofence,
  drawnCoordinates,
  drawnCenter,
  drawnRadius,
  onSave,
  onClose,
}: GeofenceModalProps) {
  const [formData, setFormData] = useState<Partial<Geofence>>({
    name: geofence?.name || '',
    type: geofence?.type || (drawnCoordinates.length > 0 ? 'polygon' : drawnCenter ? 'circle' : 'polygon'),
    coordinates: geofence?.coordinates || (drawnCoordinates.length > 0 ? drawnCoordinates : undefined),
    center: geofence?.center || drawnCenter || undefined,
    radius: geofence?.radius || drawnRadius || 500,
    min_altitude: geofence?.min_altitude,
    max_altitude: geofence?.max_altitude,
    alert_on_enter: geofence?.alert_on_enter ?? true,
    alert_on_exit: geofence?.alert_on_exit ?? true,
    enabled: geofence?.enabled ?? true,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  const hasValidGeometry =
    (formData.type === 'polygon' && formData.coordinates && formData.coordinates.length >= 3) ||
    (formData.type === 'circle' && formData.center && formData.radius && formData.radius > 0);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            {geofence ? 'Edit Geofence' : 'Create Geofence'}
          </h2>
        </div>
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Name
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Type
            </label>
            <select
              value={formData.type}
              onChange={(e) => setFormData({ ...formData, type: e.target.value as GeofenceType })}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
              disabled={!!geofence}
            >
              <option value="polygon">Polygon</option>
              <option value="circle">Circle</option>
            </select>
          </div>

          {formData.type === 'circle' && (
            <>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Center Lat
                  </label>
                  <input
                    type="number"
                    step="any"
                    value={formData.center?.[0] || ''}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        center: [parseFloat(e.target.value), formData.center?.[1] || 0],
                      })
                    }
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Center Lon
                  </label>
                  <input
                    type="number"
                    step="any"
                    value={formData.center?.[1] || ''}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        center: [formData.center?.[0] || 0, parseFloat(e.target.value)],
                      })
                    }
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Radius (meters)
                </label>
                <input
                  type="number"
                  value={formData.radius || ''}
                  onChange={(e) => setFormData({ ...formData, radius: parseInt(e.target.value) })}
                  className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  min="1"
                />
              </div>
            </>
          )}

          {formData.type === 'polygon' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Coordinates
              </label>
              <textarea
                value={formData.coordinates?.map((c) => `${c[0]}, ${c[1]}`).join('\n') || ''}
                onChange={(e) => {
                  const lines = e.target.value.split('\n');
                  const coords = lines
                    .map((line) => {
                      const parts = line.split(',').map((p) => parseFloat(p.trim()));
                      return parts.length === 2 && !isNaN(parts[0]) && !isNaN(parts[1]) ? parts : null;
                    })
                    .filter((c): c is number[] => c !== null);
                  setFormData({ ...formData, coordinates: coords });
                }}
                className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white font-mono text-sm"
                rows={4}
                placeholder="lat, lon (one per line)"
              />
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Enter coordinates as "lat, lon" (one per line). Need at least 3 points.
              </p>
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Min Altitude (m)
              </label>
              <input
                type="number"
                value={formData.min_altitude ?? ''}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    min_altitude: e.target.value ? parseFloat(e.target.value) : undefined,
                  })
                }
                className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                placeholder="Optional"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Max Altitude (m)
              </label>
              <input
                type="number"
                value={formData.max_altitude ?? ''}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    max_altitude: e.target.value ? parseFloat(e.target.value) : undefined,
                  })
                }
                className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                placeholder="Optional"
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Alert Triggers
            </label>
            <div className="flex items-center gap-4">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={formData.alert_on_enter}
                  onChange={(e) => setFormData({ ...formData, alert_on_enter: e.target.checked })}
                  className="h-4 w-4 text-blue-600 rounded border-gray-300"
                />
                <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">On Enter</span>
              </label>
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={formData.alert_on_exit}
                  onChange={(e) => setFormData({ ...formData, alert_on_exit: e.target.checked })}
                  className="h-4 w-4 text-blue-600 rounded border-gray-300"
                />
                <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">On Exit</span>
              </label>
            </div>
          </div>

          <div className="flex items-center">
            <input
              type="checkbox"
              id="enabled"
              checked={formData.enabled}
              onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
              className="h-4 w-4 text-blue-600 rounded border-gray-300"
            />
            <label htmlFor="enabled" className="ml-2 text-sm text-gray-700 dark:text-gray-300">
              Enabled
            </label>
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-700 dark:text-gray-300 border rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!formData.name || !hasValidGeometry}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
