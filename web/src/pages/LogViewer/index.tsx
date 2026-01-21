import { useEffect, useRef } from 'react';
import {
  useLogStore,
  useLogs,
  useLogsLoading,
  useLogsError,
  useLogsStreaming,
  useLogFilter,
} from '../../store/logStore';
import type { LogLevel, LogEntry } from '../../api/types';

const levelColors: Record<LogLevel, string> = {
  debug: 'text-gray-400',
  info: 'text-blue-400',
  warn: 'text-yellow-400',
  error: 'text-red-400',
};

const levelBgColors: Record<LogLevel, string> = {
  debug: 'bg-gray-500/20',
  info: 'bg-blue-500/20',
  warn: 'bg-yellow-500/20',
  error: 'bg-red-500/20',
};

function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }) + '.' + String(date.getMilliseconds()).padStart(3, '0');
}

function LogEntryRow({ log }: { log: LogEntry }) {
  return (
    <div className={`flex items-start gap-2 px-3 py-1.5 hover:bg-slate-700/50 font-mono text-sm ${levelBgColors[log.level]}`}>
      <span className="text-slate-500 shrink-0">{formatTimestamp(log.timestamp)}</span>
      <span className={`shrink-0 uppercase text-xs font-bold w-12 ${levelColors[log.level]}`}>
        {log.level}
      </span>
      <span className="text-cyan-400 shrink-0">[{log.source}]</span>
      <span className="text-slate-200 break-all">{log.message}</span>
    </div>
  );
}

export function LogViewer() {
  const logs = useLogs();
  const isLoading = useLogsLoading();
  const error = useLogsError();
  const isStreaming = useLogsStreaming();
  const filter = useLogFilter();

  const fetchLogs = useLogStore((state) => state.fetchLogs);
  const clearLogs = useLogStore((state) => state.clearLogs);
  const setFilter = useLogStore((state) => state.setFilter);
  const startStreaming = useLogStore((state) => state.startStreaming);
  const stopStreaming = useLogStore((state) => state.stopStreaming);

  const logContainerRef = useRef<HTMLDivElement>(null);
  const autoScrollRef = useRef(true);

  // Fetch initial logs on mount
  useEffect(() => {
    fetchLogs();
    return () => {
      stopStreaming();
    };
  }, [fetchLogs, stopStreaming]);

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (autoScrollRef.current && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs]);

  // Handle scroll to detect if user scrolled up
  const handleScroll = () => {
    if (logContainerRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = logContainerRef.current;
      autoScrollRef.current = scrollHeight - scrollTop - clientHeight < 50;
    }
  };

  const handleLevelChange = (level: LogLevel) => {
    setFilter({ level });
    if (!isStreaming) {
      fetchLogs();
    } else {
      // Restart streaming with new filter
      stopStreaming();
      setTimeout(() => startStreaming(), 100);
    }
  };

  const handleToggleStreaming = () => {
    if (isStreaming) {
      stopStreaming();
    } else {
      startStreaming();
    }
  };

  const handleClearLogs = async () => {
    if (confirm('Are you sure you want to clear all logs?')) {
      await clearLogs();
    }
  };

  return (
    <div className="p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-white">Log Viewer</h1>
        <div className="flex items-center gap-4">
          {/* Level Filter */}
          <div className="flex items-center gap-2">
            <label className="text-sm text-slate-400">Level:</label>
            <select
              value={filter.level}
              onChange={(e) => handleLevelChange(e.target.value as LogLevel)}
              className="px-3 py-1.5 bg-slate-700 border border-slate-600 rounded text-white text-sm"
            >
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warn</option>
              <option value="error">Error</option>
            </select>
          </div>

          {/* Source Filter */}
          <div className="flex items-center gap-2">
            <label className="text-sm text-slate-400">Source:</label>
            <input
              type="text"
              value={filter.source}
              onChange={(e) => setFilter({ source: e.target.value })}
              placeholder="All sources"
              className="px-3 py-1.5 bg-slate-700 border border-slate-600 rounded text-white text-sm w-32"
            />
          </div>

          {/* Streaming Toggle */}
          <button
            onClick={handleToggleStreaming}
            className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
              isStreaming
                ? 'bg-green-600 text-white hover:bg-green-700'
                : 'bg-slate-700 text-white hover:bg-slate-600'
            }`}
          >
            {isStreaming ? (
              <span className="flex items-center gap-2">
                <span className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
                Streaming
              </span>
            ) : (
              'Start Stream'
            )}
          </button>

          {/* Refresh Button */}
          <button
            onClick={() => fetchLogs()}
            disabled={isLoading}
            className="px-4 py-1.5 bg-slate-700 text-white rounded text-sm hover:bg-slate-600 transition-colors disabled:opacity-50"
          >
            {isLoading ? 'Loading...' : 'Refresh'}
          </button>

          {/* Clear Button */}
          <button
            onClick={handleClearLogs}
            className="px-4 py-1.5 bg-red-600/20 text-red-400 border border-red-500/50 rounded text-sm hover:bg-red-600/30 transition-colors"
          >
            Clear
          </button>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="mb-4 p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
          {error}
        </div>
      )}

      {/* Log Container */}
      <div
        ref={logContainerRef}
        onScroll={handleScroll}
        className="flex-1 bg-slate-900 rounded-lg border border-slate-700 overflow-auto"
      >
        {logs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-slate-500">
            {isLoading ? 'Loading logs...' : 'No logs available'}
          </div>
        ) : (
          <div className="py-2">
            {logs.map((log) => (
              <LogEntryRow key={log.id} log={log} />
            ))}
          </div>
        )}
      </div>

      {/* Status Bar */}
      <div className="mt-2 flex items-center justify-between text-xs text-slate-500">
        <span>{logs.length} log entries</span>
        <span>
          {autoScrollRef.current ? 'Auto-scrolling enabled' : 'Scroll to bottom to enable auto-scroll'}
        </span>
      </div>
    </div>
  );
}

export default LogViewer;
