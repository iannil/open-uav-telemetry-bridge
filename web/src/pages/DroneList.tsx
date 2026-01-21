// Drone list page

import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDroneStore } from '../store/droneStore';
import { DroneCard } from '../components/Drone/DroneCard';
import { DroneDetail } from '../components/Drone/DroneDetail';

type SortField = 'device_id' | 'protocol' | 'battery' | 'altitude' | 'timestamp';
type SortOrder = 'asc' | 'desc';

export function DroneList() {
  const navigate = useNavigate();
  const drones = useDroneStore((state) => Array.from(state.drones.values()));
  const selectedDroneId = useDroneStore((state) => state.selectedDroneId);
  const selectDrone = useDroneStore((state) => state.selectDrone);

  const [searchTerm, setSearchTerm] = useState('');
  const [sortField, setSortField] = useState<SortField>('device_id');
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc');
  const [viewMode, setViewMode] = useState<'grid' | 'table'>('grid');

  const selectedDrone = selectedDroneId
    ? drones.find((d) => d.device_id === selectedDroneId)
    : null;

  // Filter and sort drones
  const filteredDrones = useMemo(() => {
    let result = drones.filter((drone) =>
      drone.device_id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      drone.protocol_source.toLowerCase().includes(searchTerm.toLowerCase())
    );

    result.sort((a, b) => {
      let comparison = 0;
      switch (sortField) {
        case 'device_id':
          comparison = a.device_id.localeCompare(b.device_id);
          break;
        case 'protocol':
          comparison = a.protocol_source.localeCompare(b.protocol_source);
          break;
        case 'battery':
          comparison = a.status.battery_percent - b.status.battery_percent;
          break;
        case 'altitude':
          comparison = a.location.alt_gnss - b.location.alt_gnss;
          break;
        case 'timestamp':
          comparison = a.timestamp - b.timestamp;
          break;
      }
      return sortOrder === 'asc' ? comparison : -comparison;
    });

    return result;
  }, [drones, searchTerm, sortField, sortOrder]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('asc');
    }
  };

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortField !== field) return null;
    return sortOrder === 'asc' ? (
      <span className="ml-1"></span>
    ) : (
      <span className="ml-1"></span>
    );
  };

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-white">Drones</h1>
        <span className="text-sm text-gray-500">
          {filteredDrones.length} of {drones.length} drones
        </span>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4 mb-4">
        <div className="flex-1">
          <input
            type="text"
            placeholder="Search by ID or protocol..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
        </div>
        <select
          value={sortField}
          onChange={(e) => setSortField(e.target.value as SortField)}
          className="px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white focus:outline-none focus:border-blue-500"
        >
          <option value="device_id">Sort by ID</option>
          <option value="protocol">Sort by Protocol</option>
          <option value="battery">Sort by Battery</option>
          <option value="altitude">Sort by Altitude</option>
          <option value="timestamp">Sort by Last Update</option>
        </select>
        <button
          onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
          className="px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white hover:bg-gray-700"
        >
          {sortOrder === 'asc' ? 'Asc' : 'Desc'}
        </button>
        <div className="flex border border-gray-700 rounded-lg overflow-hidden">
          <button
            onClick={() => setViewMode('grid')}
            className={`px-3 py-2 ${
              viewMode === 'grid' ? 'bg-blue-600' : 'bg-gray-800 hover:bg-gray-700'
            }`}
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
            </svg>
          </button>
          <button
            onClick={() => setViewMode('table')}
            className={`px-3 py-2 ${
              viewMode === 'table' ? 'bg-blue-600' : 'bg-gray-800 hover:bg-gray-700'
            }`}
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 10h16M4 14h16M4 18h16" />
            </svg>
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 flex gap-4 min-h-0 overflow-hidden">
        <div className={`${selectedDrone ? 'flex-1' : 'w-full'} overflow-auto`}>
          {filteredDrones.length > 0 ? (
            viewMode === 'grid' ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {filteredDrones.map((drone) => (
                  <DroneCard
                    key={drone.device_id}
                    drone={drone}
                    isSelected={drone.device_id === selectedDroneId}
                    onClick={() => selectDrone(drone.device_id === selectedDroneId ? null : drone.device_id)}
                  />
                ))}
              </div>
            ) : (
              <table className="w-full text-left">
                <thead className="bg-gray-800 sticky top-0">
                  <tr>
                    <th
                      className="px-4 py-3 cursor-pointer hover:bg-gray-700"
                      onClick={() => handleSort('device_id')}
                    >
                      Device ID <SortIcon field="device_id" />
                    </th>
                    <th
                      className="px-4 py-3 cursor-pointer hover:bg-gray-700"
                      onClick={() => handleSort('protocol')}
                    >
                      Protocol <SortIcon field="protocol" />
                    </th>
                    <th className="px-4 py-3">Position</th>
                    <th
                      className="px-4 py-3 cursor-pointer hover:bg-gray-700"
                      onClick={() => handleSort('altitude')}
                    >
                      Altitude <SortIcon field="altitude" />
                    </th>
                    <th
                      className="px-4 py-3 cursor-pointer hover:bg-gray-700"
                      onClick={() => handleSort('battery')}
                    >
                      Battery <SortIcon field="battery" />
                    </th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-700">
                  {filteredDrones.map((drone) => (
                    <tr
                      key={drone.device_id}
                      className={`hover:bg-gray-800 cursor-pointer ${
                        drone.device_id === selectedDroneId ? 'bg-blue-900/30' : ''
                      }`}
                      onClick={() => selectDrone(drone.device_id === selectedDroneId ? null : drone.device_id)}
                    >
                      <td className="px-4 py-3 font-medium">{drone.device_id}</td>
                      <td className="px-4 py-3 text-gray-400">{drone.protocol_source}</td>
                      <td className="px-4 py-3 text-gray-400 font-mono text-sm">
                        {drone.location.lat.toFixed(4)}, {drone.location.lon.toFixed(4)}
                      </td>
                      <td className="px-4 py-3">{drone.location.alt_gnss.toFixed(1)}m</td>
                      <td className="px-4 py-3">
                        <span
                          className={
                            drone.status.battery_percent > 50
                              ? 'text-green-400'
                              : drone.status.battery_percent > 20
                              ? 'text-yellow-400'
                              : 'text-red-400'
                          }
                        >
                          {drone.status.battery_percent}%
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`px-2 py-0.5 rounded text-xs ${
                            drone.status.armed
                              ? 'bg-red-500/20 text-red-400'
                              : 'bg-green-500/20 text-green-400'
                          }`}
                        >
                          {drone.status.armed ? 'Armed' : 'Disarmed'}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            navigate(`/tracks?drone=${drone.device_id}`);
                          }}
                          className="text-blue-400 hover:text-blue-300 text-sm"
                        >
                          View Track
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )
          ) : (
            <div className="flex items-center justify-center h-full">
              <div className="text-center text-gray-500">
                <p>No drones found</p>
                {searchTerm && (
                  <p className="text-sm mt-2">Try adjusting your search term</p>
                )}
              </div>
            </div>
          )}
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

export default DroneList;
