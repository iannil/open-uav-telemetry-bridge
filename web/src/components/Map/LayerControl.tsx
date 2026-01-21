// Layer control component for switching between tile providers

import { useState, useEffect, useRef } from 'react';
import { TileLayer, useMap } from 'react-leaflet';
import {
  tileProviders,
  getTileProvider,
  getTileProviderKeys,
  getSavedTileProvider,
  saveTileProvider,
} from './TileProviders';

interface LayerControlProps {
  defaultProvider?: string;
  onProviderChange?: (providerKey: string) => void;
}

// Layer switcher UI component
function LayerSwitcher({
  currentProvider,
  onChange,
}: {
  currentProvider: string;
  onChange: (key: string) => void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const providerKeys = getTileProviderKeys();

  // Close menu when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const currentProviderName = tileProviders[currentProvider]?.name || 'Map';

  return (
    <div
      ref={menuRef}
      className="leaflet-control leaflet-bar"
      style={{ position: 'absolute', top: '10px', right: '10px', zIndex: 1000 }}
    >
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="bg-white dark:bg-gray-800 px-3 py-2 rounded shadow-lg hover:bg-gray-50 dark:hover:bg-gray-700 border border-gray-300 dark:border-gray-600 flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-200"
        title="Change map layer"
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"
          />
        </svg>
        <span className="hidden sm:inline">{currentProviderName}</span>
        <svg
          className={`w-4 h-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-1 w-48 bg-white dark:bg-gray-800 rounded-lg shadow-xl border border-gray-200 dark:border-gray-600 py-1 max-h-64 overflow-y-auto">
          {providerKeys.map((key) => {
            const provider = tileProviders[key];
            const isActive = key === currentProvider;
            return (
              <button
                key={key}
                onClick={() => {
                  onChange(key);
                  setIsOpen(false);
                }}
                className={`w-full text-left px-4 py-2 text-sm transition-colors ${
                  isActive
                    ? 'bg-blue-50 dark:bg-blue-900 text-blue-600 dark:text-blue-400 font-medium'
                    : 'text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
                }`}
              >
                {provider.name}
                {isActive && (
                  <svg
                    className="w-4 h-4 inline-block ml-2"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                      clipRule="evenodd"
                    />
                  </svg>
                )}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

// The main LayerControl component that manages both the tile layer and switcher UI
export function LayerControl({ defaultProvider, onProviderChange }: LayerControlProps) {
  const [providerKey, setProviderKey] = useState<string>(defaultProvider || getSavedTileProvider());
  const map = useMap();

  // Handle provider change
  const handleProviderChange = (key: string) => {
    setProviderKey(key);
    saveTileProvider(key);
    onProviderChange?.(key);
  };

  // Stop map events from propagating to the control
  useEffect(() => {
    const container = document.querySelector('.leaflet-control-container');
    if (container) {
      container.addEventListener('mousedown', (e) => e.stopPropagation());
      container.addEventListener('dblclick', (e) => e.stopPropagation());
    }
  }, [map]);

  const provider = getTileProvider(providerKey);

  return (
    <>
      <TileLayer
        key={providerKey} // Force re-render when provider changes
        attribution={provider.attribution}
        url={provider.url}
        subdomains={provider.subdomains}
        maxZoom={provider.maxZoom}
      />
      <LayerSwitcher currentProvider={providerKey} onChange={handleProviderChange} />
    </>
  );
}

export default LayerControl;
