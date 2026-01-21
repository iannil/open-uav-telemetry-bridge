// Alerts page - Alert management and rule configuration

import { useEffect, useState } from 'react';
import {
  useAlertStore,
  useAlerts,
  useAlertRules,
  useAlertStats,
  useAlertsLoading,
  useAlertsError,
  useAlertFilter,
} from '../../store/alertStore';
import type { AlertRule, AlertSeverity, AlertType } from '../../api/types';

// Severity badge colors
const severityColors: Record<AlertSeverity, string> = {
  info: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  critical: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
};

// Alert type display names
const alertTypeNames: Record<AlertType, string> = {
  battery_low: 'Battery Low',
  connection_lost: 'Connection Lost',
  signal_weak: 'Signal Weak',
  geofence_breach: 'Geofence Breach',
  custom: 'Custom',
};

// Operator display names
const operatorNames: Record<string, string> = {
  '<': 'Less than',
  '>': 'Greater than',
  '<=': 'Less than or equal',
  '>=': 'Greater than or equal',
  '==': 'Equal to',
  '!=': 'Not equal to',
};

type TabType = 'alerts' | 'rules';

export default function AlertsPage() {
  const [activeTab, setActiveTab] = useState<TabType>('alerts');
  const [showRuleModal, setShowRuleModal] = useState(false);
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null);

  const alerts = useAlerts();
  const rules = useAlertRules();
  const stats = useAlertStats();
  const isLoading = useAlertsLoading();
  const error = useAlertsError();
  const filter = useAlertFilter();

  const {
    fetchAlerts,
    fetchRules,
    acknowledgeAlert,
    clearAlerts,
    createRule,
    updateRule,
    deleteRule,
    setFilter,
  } = useAlertStore();

  // Initial data fetch
  useEffect(() => {
    fetchAlerts();
    fetchRules();
  }, [fetchAlerts, fetchRules]);

  // Refetch when filter changes
  useEffect(() => {
    fetchAlerts();
  }, [filter, fetchAlerts]);

  const handleAcknowledge = async (alertId: string) => {
    await acknowledgeAlert(alertId);
  };

  const handleClearAll = async () => {
    if (window.confirm('Are you sure you want to clear all alerts?')) {
      await clearAlerts();
      await fetchAlerts();
    }
  };

  const handleFilterChange = (key: 'acknowledged' | 'deviceId', value: boolean | null | string) => {
    setFilter({ [key]: value });
  };

  const handleEditRule = (rule: AlertRule) => {
    setEditingRule(rule);
    setShowRuleModal(true);
  };

  const handleCreateRule = () => {
    setEditingRule(null);
    setShowRuleModal(true);
  };

  const handleDeleteRule = async (ruleId: string) => {
    if (window.confirm('Are you sure you want to delete this rule?')) {
      await deleteRule(ruleId);
    }
  };

  const handleSaveRule = async (rule: Partial<AlertRule>) => {
    if (editingRule) {
      await updateRule(editingRule.id, rule);
    } else {
      await createRule(rule);
    }
    setShowRuleModal(false);
    setEditingRule(null);
  };

  const formatTimestamp = (ts: number) => {
    return new Date(ts).toLocaleString();
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Alerts</h1>
        <p className="text-gray-600 dark:text-gray-400">Monitor and manage system alerts</p>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <div className="text-sm text-gray-500 dark:text-gray-400">Total Alerts</div>
            <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.total_alerts}</div>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <div className="text-sm text-gray-500 dark:text-gray-400">Unacknowledged</div>
            <div className="text-2xl font-bold text-red-600 dark:text-red-400">{stats.unacknowledged}</div>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <div className="text-sm text-gray-500 dark:text-gray-400">Active Rules</div>
            <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.rules_count}</div>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <div className="text-sm text-gray-500 dark:text-gray-400">Devices with Alerts</div>
            <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.devices_with_alerts}</div>
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
      <div className="mb-4 border-b border-gray-200 dark:border-gray-700">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('alerts')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'alerts'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'
            }`}
          >
            Alerts
          </button>
          <button
            onClick={() => setActiveTab('rules')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'rules'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'
            }`}
          >
            Rules
          </button>
        </nav>
      </div>

      {/* Alerts Tab */}
      {activeTab === 'alerts' && (
        <div>
          {/* Filters and Actions */}
          <div className="mb-4 flex flex-wrap gap-4 items-center justify-between">
            <div className="flex gap-4 items-center">
              <select
                value={filter.acknowledged === null ? 'all' : filter.acknowledged ? 'acked' : 'unacked'}
                onChange={(e) => {
                  const val = e.target.value;
                  handleFilterChange('acknowledged', val === 'all' ? null : val === 'acked');
                }}
                className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
              >
                <option value="all">All Alerts</option>
                <option value="unacked">Unacknowledged</option>
                <option value="acked">Acknowledged</option>
              </select>
              <input
                type="text"
                placeholder="Filter by Device ID"
                value={filter.deviceId}
                onChange={(e) => handleFilterChange('deviceId', e.target.value)}
                className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
              />
            </div>
            <button
              onClick={handleClearAll}
              className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
            >
              Clear All
            </button>
          </div>

          {/* Alerts List */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            {isLoading ? (
              <div className="p-8 text-center text-gray-500 dark:text-gray-400">Loading...</div>
            ) : alerts.length === 0 ? (
              <div className="p-8 text-center text-gray-500 dark:text-gray-400">No alerts found</div>
            ) : (
              <div className="divide-y divide-gray-200 dark:divide-gray-700">
                {alerts.map((alert) => (
                  <div
                    key={alert.id}
                    className={`p-4 ${alert.acknowledged ? 'opacity-60' : ''}`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <span className={`px-2 py-1 text-xs font-medium rounded ${severityColors[alert.severity]}`}>
                            {alert.severity.toUpperCase()}
                          </span>
                          <span className="text-xs text-gray-500 dark:text-gray-400">
                            {alertTypeNames[alert.type]}
                          </span>
                          {alert.acknowledged && (
                            <span className="text-xs text-green-600 dark:text-green-400">
                              Acknowledged
                            </span>
                          )}
                        </div>
                        <p className="text-gray-900 dark:text-white font-medium">{alert.message}</p>
                        <div className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                          <span>Device: {alert.device_id}</span>
                          <span className="mx-2">|</span>
                          <span>{formatTimestamp(alert.timestamp)}</span>
                          {alert.acked_at && (
                            <>
                              <span className="mx-2">|</span>
                              <span>Acked: {formatTimestamp(alert.acked_at)}</span>
                            </>
                          )}
                        </div>
                      </div>
                      {!alert.acknowledged && (
                        <button
                          onClick={() => handleAcknowledge(alert.id)}
                          className="ml-4 px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700"
                        >
                          Acknowledge
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Rules Tab */}
      {activeTab === 'rules' && (
        <div>
          <div className="mb-4 flex justify-end">
            <button
              onClick={handleCreateRule}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Create Rule
            </button>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            {rules.length === 0 ? (
              <div className="p-8 text-center text-gray-500 dark:text-gray-400">No rules configured</div>
            ) : (
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-900">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Name
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Type
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Condition
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Severity
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                  {rules.map((rule) => (
                    <tr key={rule.id}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                        {rule.name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        {alertTypeNames[rule.type]}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        {rule.condition.field} {rule.condition.operator} {rule.condition.threshold}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`px-2 py-1 text-xs font-medium rounded ${severityColors[rule.severity]}`}>
                          {rule.severity}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`px-2 py-1 text-xs font-medium rounded ${
                            rule.enabled
                              ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
                              : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
                          }`}
                        >
                          {rule.enabled ? 'Enabled' : 'Disabled'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <button
                          onClick={() => handleEditRule(rule)}
                          className="text-blue-600 hover:text-blue-900 dark:text-blue-400 mr-3"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDeleteRule(rule.id)}
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
        </div>
      )}

      {/* Rule Modal */}
      {showRuleModal && (
        <RuleModal
          rule={editingRule}
          onSave={handleSaveRule}
          onClose={() => {
            setShowRuleModal(false);
            setEditingRule(null);
          }}
        />
      )}
    </div>
  );
}

// Rule Modal Component
interface RuleModalProps {
  rule: AlertRule | null;
  onSave: (rule: Partial<AlertRule>) => void;
  onClose: () => void;
}

function RuleModal({ rule, onSave, onClose }: RuleModalProps) {
  const [formData, setFormData] = useState<Partial<AlertRule>>({
    name: rule?.name || '',
    type: rule?.type || 'battery_low',
    severity: rule?.severity || 'warning',
    enabled: rule?.enabled ?? true,
    condition: rule?.condition || { field: 'battery_percent', operator: '<', threshold: 20 },
    cooldown_ms: rule?.cooldown_ms || 60000,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  const handleConditionChange = (key: string, value: string | number) => {
    setFormData({
      ...formData,
      condition: { ...formData.condition!, [key]: value },
    });
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            {rule ? 'Edit Rule' : 'Create Rule'}
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
              onChange={(e) => setFormData({ ...formData, type: e.target.value as AlertType })}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            >
              {Object.entries(alertTypeNames).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Severity
            </label>
            <select
              value={formData.severity}
              onChange={(e) => setFormData({ ...formData, severity: e.target.value as AlertSeverity })}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            >
              <option value="info">Info</option>
              <option value="warning">Warning</option>
              <option value="critical">Critical</option>
            </select>
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Condition
            </label>
            <div className="grid grid-cols-3 gap-2">
              <select
                value={formData.condition?.field}
                onChange={(e) => handleConditionChange('field', e.target.value)}
                className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
              >
                <option value="battery_percent">Battery %</option>
                <option value="signal_quality">Signal Quality</option>
                <option value="altitude">Altitude</option>
                <option value="speed">Speed</option>
              </select>
              <select
                value={formData.condition?.operator}
                onChange={(e) => handleConditionChange('operator', e.target.value)}
                className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
              >
                {Object.entries(operatorNames).map(([value, _]) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </select>
              <input
                type="number"
                value={formData.condition?.threshold}
                onChange={(e) => handleConditionChange('threshold', parseFloat(e.target.value))}
                className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Cooldown (seconds)
            </label>
            <input
              type="number"
              value={(formData.cooldown_ms || 0) / 1000}
              onChange={(e) => setFormData({ ...formData, cooldown_ms: parseInt(e.target.value) * 1000 })}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
              min="1"
            />
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
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
