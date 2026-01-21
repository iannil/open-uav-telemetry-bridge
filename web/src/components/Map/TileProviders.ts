// Map tile provider configurations for different map services

export interface TileProvider {
  name: string;
  url: string;
  attribution: string;
  subdomains?: string;
  maxZoom?: number;
}

export const tileProviders: Record<string, TileProvider> = {
  // OpenStreetMap - Default
  openstreetmap: {
    name: 'OpenStreetMap',
    url: 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    maxZoom: 19,
  },

  // AMap (Gaode) - China
  amap: {
    name: '高德地图',
    url: 'https://webrd0{s}.is.autonavi.com/appmaptile?lang=zh_cn&size=1&scale=1&style=8&x={x}&y={y}&z={z}',
    attribution: '&copy; <a href="https://www.amap.com/">高德地图</a>',
    subdomains: '1234',
    maxZoom: 18,
  },

  // AMap Satellite
  amapSatellite: {
    name: '高德卫星',
    url: 'https://webst0{s}.is.autonavi.com/appmaptile?style=6&x={x}&y={y}&z={z}',
    attribution: '&copy; <a href="https://www.amap.com/">高德地图</a>',
    subdomains: '1234',
    maxZoom: 18,
  },

  // Google Maps
  google: {
    name: 'Google Maps',
    url: 'https://mt{s}.google.com/vt/lyrs=m&x={x}&y={y}&z={z}',
    attribution: '&copy; <a href="https://www.google.com/maps">Google Maps</a>',
    subdomains: '0123',
    maxZoom: 20,
  },

  // Google Satellite
  googleSatellite: {
    name: 'Google Satellite',
    url: 'https://mt{s}.google.com/vt/lyrs=s&x={x}&y={y}&z={z}',
    attribution: '&copy; <a href="https://www.google.com/maps">Google Maps</a>',
    subdomains: '0123',
    maxZoom: 20,
  },

  // Google Hybrid (Satellite + Labels)
  googleHybrid: {
    name: 'Google Hybrid',
    url: 'https://mt{s}.google.com/vt/lyrs=y&x={x}&y={y}&z={z}',
    attribution: '&copy; <a href="https://www.google.com/maps">Google Maps</a>',
    subdomains: '0123',
    maxZoom: 20,
  },

  // CartoDB Light - Clean minimal style
  cartoLight: {
    name: 'CartoDB Light',
    url: 'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png',
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>',
    maxZoom: 20,
  },

  // CartoDB Dark - Dark theme
  cartoDark: {
    name: 'CartoDB Dark',
    url: 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png',
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>',
    maxZoom: 20,
  },
};

// Default provider key
export const defaultProvider = 'openstreetmap';

// Get tile provider by key
export function getTileProvider(key: string): TileProvider {
  return tileProviders[key] || tileProviders[defaultProvider];
}

// Get all provider keys
export function getTileProviderKeys(): string[] {
  return Object.keys(tileProviders);
}

// Local storage key for persisting user preference
export const TILE_PROVIDER_STORAGE_KEY = 'outb-map-tile-provider';

// Get saved tile provider from localStorage
export function getSavedTileProvider(): string {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(TILE_PROVIDER_STORAGE_KEY) || defaultProvider;
  }
  return defaultProvider;
}

// Save tile provider to localStorage
export function saveTileProvider(key: string): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem(TILE_PROVIDER_STORAGE_KEY, key);
  }
}
