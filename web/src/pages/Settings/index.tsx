import { useState, useEffect } from 'react';
import { useConfigStore, useConfig, useConfigLoading, useConfigError } from '../../store/configStore';
import { api } from '../../api/client';

type TabType = 'adapters' | 'publishers' | 'general';

export function Settings() {
  const [activeTab, setActiveTab] = useState<TabType>('adapters');
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const fetchConfig = useConfigStore((state) => state.fetchConfig);
  const config = useConfig();
  const isLoading = useConfigLoading();
  const error = useConfigError();

  useEffect(() => {
    fetchConfig();
  }, [fetchConfig]);

  const handleExport = async () => {
    try {
      const yaml = await api.exportConfig();
      const blob = new Blob([yaml], { type: 'application/x-yaml' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'config.yaml';
      a.click();
      URL.revokeObjectURL(url);
      setMessage({ type: 'success', text: 'Configuration exported successfully' });
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Export failed' });
    }
  };

  if (isLoading) {
    return (
      <div className="p-6 flex items-center justify-center">
        <div className="text-slate-400">Loading configuration...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="bg-red-500/10 border border-red-500/50 rounded-lg p-4 text-red-400">
          {error}
        </div>
      </div>
    );
  }

  if (!config) {
    return (
      <div className="p-6">
        <div className="text-slate-400">No configuration available</div>
      </div>
    );
  }

  const tabs: { id: TabType; label: string }[] = [
    { id: 'adapters', label: 'Adapters' },
    { id: 'publishers', label: 'Publishers' },
    { id: 'general', label: 'General' },
  ];

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Settings</h1>
        <button
          onClick={handleExport}
          className="px-4 py-2 bg-slate-700 text-white rounded-lg hover:bg-slate-600 transition-colors text-sm"
        >
          Export Config
        </button>
      </div>

      {message && (
        <div
          className={`mb-4 p-4 rounded-lg ${
            message.type === 'success'
              ? 'bg-green-500/10 border border-green-500/50 text-green-400'
              : 'bg-red-500/10 border border-red-500/50 text-red-400'
          }`}
        >
          {message.text}
          <button
            onClick={() => setMessage(null)}
            className="ml-2 text-current opacity-70 hover:opacity-100"
          >
            &times;
          </button>
        </div>
      )}

      {/* Tabs */}
      <div className="border-b border-slate-700 mb-6">
        <nav className="flex space-x-4">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-400'
                  : 'border-transparent text-slate-400 hover:text-white'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="space-y-6">
        {activeTab === 'adapters' && <AdaptersTab config={config} setMessage={setMessage} />}
        {activeTab === 'publishers' && <PublishersTab config={config} setMessage={setMessage} />}
        {activeTab === 'general' && <GeneralTab config={config} setMessage={setMessage} />}
      </div>
    </div>
  );
}

// Adapters Tab Component
function AdaptersTab({
  config,
  setMessage,
}: {
  config: NonNullable<ReturnType<typeof useConfig>>;
  setMessage: (msg: { type: 'success' | 'error'; text: string } | null) => void;
}) {
  const fetchConfig = useConfigStore((state) => state.fetchConfig);

  const handleSaveMAVLink = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);

    try {
      await api.updateMAVLinkConfig({
        enabled: data.get('enabled') === 'on',
        connection_type: data.get('connection_type') as string,
        address: data.get('address') as string,
        serial_port: data.get('serial_port') as string,
        serial_baud: parseInt(data.get('serial_baud') as string) || 57600,
      });
      await fetchConfig();
      setMessage({ type: 'success', text: 'MAVLink configuration saved' });
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' });
    }
  };

  const handleSaveDJI = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);

    try {
      await api.updateDJIConfig({
        enabled: data.get('enabled') === 'on',
        listen_address: data.get('listen_address') as string,
        max_clients: parseInt(data.get('max_clients') as string) || 10,
      });
      await fetchConfig();
      setMessage({ type: 'success', text: 'DJI configuration saved' });
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' });
    }
  };

  return (
    <div className="space-y-6">
      {/* MAVLink */}
      <div className="bg-slate-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">MAVLink Adapter</h3>
        <form onSubmit={handleSaveMAVLink} className="space-y-4">
          <div className="flex items-center">
            <input
              type="checkbox"
              id="mavlink-enabled"
              name="enabled"
              defaultChecked={config.mavlink.enabled}
              className="mr-2"
            />
            <label htmlFor="mavlink-enabled" className="text-slate-300">Enabled</label>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Connection Type</label>
              <select
                name="connection_type"
                defaultValue={config.mavlink.connection_type}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              >
                <option value="udp">UDP</option>
                <option value="tcp">TCP</option>
                <option value="serial">Serial</option>
              </select>
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Address</label>
              <input
                type="text"
                name="address"
                defaultValue={config.mavlink.address}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Serial Port</label>
              <input
                type="text"
                name="serial_port"
                defaultValue={config.mavlink.serial_port}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Serial Baud Rate</label>
              <input
                type="number"
                name="serial_baud"
                defaultValue={config.mavlink.serial_baud}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
          </div>

          <button
            type="submit"
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            Save MAVLink Settings
          </button>
        </form>
      </div>

      {/* DJI */}
      <div className="bg-slate-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">DJI Adapter</h3>
        <form onSubmit={handleSaveDJI} className="space-y-4">
          <div className="flex items-center">
            <input
              type="checkbox"
              id="dji-enabled"
              name="enabled"
              defaultChecked={config.dji.enabled}
              className="mr-2"
            />
            <label htmlFor="dji-enabled" className="text-slate-300">Enabled</label>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Listen Address</label>
              <input
                type="text"
                name="listen_address"
                defaultValue={config.dji.listen_address}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Max Clients</label>
              <input
                type="number"
                name="max_clients"
                defaultValue={config.dji.max_clients}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
          </div>

          <button
            type="submit"
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            Save DJI Settings
          </button>
        </form>
      </div>
    </div>
  );
}

// Publishers Tab Component
function PublishersTab({
  config,
  setMessage,
}: {
  config: NonNullable<ReturnType<typeof useConfig>>;
  setMessage: (msg: { type: 'success' | 'error'; text: string } | null) => void;
}) {
  const fetchConfig = useConfigStore((state) => state.fetchConfig);

  const handleSaveMQTT = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);

    const update: Record<string, unknown> = {
      enabled: data.get('enabled') === 'on',
      broker: data.get('broker'),
      client_id: data.get('client_id'),
      topic_prefix: data.get('topic_prefix'),
      qos: parseInt(data.get('qos') as string),
      username: data.get('username'),
    };

    const password = data.get('password') as string;
    if (password) {
      update.password = password;
    }

    try {
      await api.updateMQTTConfig(update);
      await fetchConfig();
      setMessage({ type: 'success', text: 'MQTT configuration saved' });
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' });
    }
  };

  const handleSaveGB28181 = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);

    const update: Record<string, unknown> = {
      enabled: data.get('enabled') === 'on',
      device_id: data.get('device_id'),
      device_name: data.get('device_name'),
      local_ip: data.get('local_ip'),
      local_port: parseInt(data.get('local_port') as string),
      server_id: data.get('server_id'),
      server_ip: data.get('server_ip'),
      server_port: parseInt(data.get('server_port') as string),
      server_domain: data.get('server_domain'),
      username: data.get('username'),
      transport: data.get('transport'),
      register_expires: parseInt(data.get('register_expires') as string),
      heartbeat_interval: parseInt(data.get('heartbeat_interval') as string),
      position_interval: parseInt(data.get('position_interval') as string),
    };

    const password = data.get('password') as string;
    if (password) {
      update.password = password;
    }

    try {
      await api.updateGB28181Config(update);
      await fetchConfig();
      setMessage({ type: 'success', text: 'GB28181 configuration saved' });
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' });
    }
  };

  return (
    <div className="space-y-6">
      {/* MQTT */}
      <div className="bg-slate-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">MQTT Publisher</h3>
        <form onSubmit={handleSaveMQTT} className="space-y-4">
          <div className="flex items-center">
            <input
              type="checkbox"
              id="mqtt-enabled"
              name="enabled"
              defaultChecked={config.mqtt.enabled}
              className="mr-2"
            />
            <label htmlFor="mqtt-enabled" className="text-slate-300">Enabled</label>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Broker URL</label>
              <input
                type="text"
                name="broker"
                defaultValue={config.mqtt.broker}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Client ID</label>
              <input
                type="text"
                name="client_id"
                defaultValue={config.mqtt.client_id}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Topic Prefix</label>
              <input
                type="text"
                name="topic_prefix"
                defaultValue={config.mqtt.topic_prefix}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">QoS</label>
              <select
                name="qos"
                defaultValue={config.mqtt.qos}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              >
                <option value="0">0 - At most once</option>
                <option value="1">1 - At least once</option>
                <option value="2">2 - Exactly once</option>
              </select>
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Username</label>
              <input
                type="text"
                name="username"
                defaultValue={config.mqtt.username}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">
                Password {config.mqtt.has_password && <span className="text-green-400">(set)</span>}
              </label>
              <input
                type="password"
                name="password"
                placeholder="Leave empty to keep current"
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
          </div>

          <button
            type="submit"
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            Save MQTT Settings
          </button>
        </form>
      </div>

      {/* GB28181 */}
      <div className="bg-slate-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">GB/T 28181 Publisher</h3>
        <form onSubmit={handleSaveGB28181} className="space-y-4">
          <div className="flex items-center">
            <input
              type="checkbox"
              id="gb28181-enabled"
              name="enabled"
              defaultChecked={config.gb28181.enabled}
              className="mr-2"
            />
            <label htmlFor="gb28181-enabled" className="text-slate-300">Enabled</label>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Device ID</label>
              <input
                type="text"
                name="device_id"
                defaultValue={config.gb28181.device_id}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Device Name</label>
              <input
                type="text"
                name="device_name"
                defaultValue={config.gb28181.device_name}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Local IP</label>
              <input
                type="text"
                name="local_ip"
                defaultValue={config.gb28181.local_ip}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Local Port</label>
              <input
                type="number"
                name="local_port"
                defaultValue={config.gb28181.local_port}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Server ID</label>
              <input
                type="text"
                name="server_id"
                defaultValue={config.gb28181.server_id}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Server IP</label>
              <input
                type="text"
                name="server_ip"
                defaultValue={config.gb28181.server_ip}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Server Port</label>
              <input
                type="number"
                name="server_port"
                defaultValue={config.gb28181.server_port}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Server Domain</label>
              <input
                type="text"
                name="server_domain"
                defaultValue={config.gb28181.server_domain}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Username</label>
              <input
                type="text"
                name="username"
                defaultValue={config.gb28181.username}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">
                Password {config.gb28181.has_password && <span className="text-green-400">(set)</span>}
              </label>
              <input
                type="password"
                name="password"
                placeholder="Leave empty to keep current"
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Transport</label>
              <select
                name="transport"
                defaultValue={config.gb28181.transport}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              >
                <option value="udp">UDP</option>
                <option value="tcp">TCP</option>
              </select>
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Register Expires (s)</label>
              <input
                type="number"
                name="register_expires"
                defaultValue={config.gb28181.register_expires}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Heartbeat Interval (s)</label>
              <input
                type="number"
                name="heartbeat_interval"
                defaultValue={config.gb28181.heartbeat_interval}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Position Interval (s)</label>
              <input
                type="number"
                name="position_interval"
                defaultValue={config.gb28181.position_interval}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
          </div>

          <button
            type="submit"
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            Save GB28181 Settings
          </button>
        </form>
      </div>
    </div>
  );
}

// General Tab Component
function GeneralTab({
  config,
  setMessage,
}: {
  config: NonNullable<ReturnType<typeof useConfig>>;
  setMessage: (msg: { type: 'success' | 'error'; text: string } | null) => void;
}) {
  const fetchConfig = useConfigStore((state) => state.fetchConfig);

  const handleSaveThrottle = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);

    try {
      await api.updateThrottleConfig({
        default_rate_hz: parseFloat(data.get('default_rate_hz') as string),
        min_rate_hz: parseFloat(data.get('min_rate_hz') as string),
        max_rate_hz: parseFloat(data.get('max_rate_hz') as string),
      });
      await fetchConfig();
      setMessage({ type: 'success', text: 'Throttle configuration saved' });
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' });
    }
  };

  const handleSaveCoordinate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);

    try {
      await api.updateCoordinateConfig({
        convert_gcj02: data.get('convert_gcj02') === 'on',
        convert_bd09: data.get('convert_bd09') === 'on',
      });
      await fetchConfig();
      setMessage({ type: 'success', text: 'Coordinate configuration saved' });
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' });
    }
  };

  const handleSaveTrack = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);

    try {
      await api.updateTrackConfig({
        enabled: data.get('enabled') === 'on',
        max_points_per_drone: parseInt(data.get('max_points_per_drone') as string),
        sample_interval_ms: parseInt(data.get('sample_interval_ms') as string),
      });
      await fetchConfig();
      setMessage({ type: 'success', text: 'Track configuration saved' });
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Save failed' });
    }
  };

  return (
    <div className="space-y-6">
      {/* Throttle */}
      <div className="bg-slate-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Frequency Throttling</h3>
        <form onSubmit={handleSaveThrottle} className="space-y-4">
          <div className="grid grid-cols-3 gap-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Default Rate (Hz)</label>
              <input
                type="number"
                name="default_rate_hz"
                step="0.1"
                defaultValue={config.throttle.default_rate_hz}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Min Rate (Hz)</label>
              <input
                type="number"
                name="min_rate_hz"
                step="0.1"
                defaultValue={config.throttle.min_rate_hz}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Max Rate (Hz)</label>
              <input
                type="number"
                name="max_rate_hz"
                step="0.1"
                defaultValue={config.throttle.max_rate_hz}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
          </div>

          <button
            type="submit"
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            Save Throttle Settings
          </button>
        </form>
      </div>

      {/* Coordinate Conversion */}
      <div className="bg-slate-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Coordinate Conversion</h3>
        <p className="text-sm text-slate-400 mb-4">
          Convert WGS84 coordinates to China map coordinate systems
        </p>
        <form onSubmit={handleSaveCoordinate} className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center">
              <input
                type="checkbox"
                id="convert_gcj02"
                name="convert_gcj02"
                defaultChecked={config.coordinate.convert_gcj02}
                className="mr-2"
              />
              <label htmlFor="convert_gcj02" className="text-slate-300">
                Convert to GCJ02 (Amap, Tencent, Google China)
              </label>
            </div>
            <div className="flex items-center">
              <input
                type="checkbox"
                id="convert_bd09"
                name="convert_bd09"
                defaultChecked={config.coordinate.convert_bd09}
                className="mr-2"
              />
              <label htmlFor="convert_bd09" className="text-slate-300">
                Convert to BD09 (Baidu Maps)
              </label>
            </div>
          </div>

          <button
            type="submit"
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            Save Coordinate Settings
          </button>
        </form>
      </div>

      {/* Track Storage */}
      <div className="bg-slate-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Track Storage</h3>
        <form onSubmit={handleSaveTrack} className="space-y-4">
          <div className="flex items-center mb-4">
            <input
              type="checkbox"
              id="track-enabled"
              name="enabled"
              defaultChecked={config.track.enabled}
              className="mr-2"
            />
            <label htmlFor="track-enabled" className="text-slate-300">Enable Track Storage</label>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Max Points Per Drone</label>
              <input
                type="number"
                name="max_points_per_drone"
                defaultValue={config.track.max_points_per_drone}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Sample Interval (ms)</label>
              <input
                type="number"
                name="sample_interval_ms"
                defaultValue={config.track.sample_interval_ms}
                className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white"
              />
            </div>
          </div>

          <button
            type="submit"
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
          >
            Save Track Settings
          </button>
        </form>
      </div>
    </div>
  );
}

export default Settings;
