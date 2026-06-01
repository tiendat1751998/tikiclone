"use client";

import { useState, useRef, useCallback, useEffect } from "react";
import { deliveryApi } from "@/lib/api/client";
import { MapPin, Search, X, Navigation, Clock, Truck, ChevronRight } from "lucide-react";

interface LocationResult {
  address: string;
  name?: string;
  lat: number;
  lng: number;
}

interface DeliveryLocationPickerProps {
  onLocationSelect: (location: LocationResult) => void;
  onRouteCalculated?: (distance: number, duration: number) => void;
  pickupLocation?: LocationResult | null;
  label?: string;
  placeholder?: string;
  defaultValue?: string;
}

export function DeliveryLocationPicker({
  onLocationSelect,
  onRouteCalculated,
  pickupLocation,
  label = "Giao đến",
  placeholder = "Tìm địa chỉ giao hàng...",
  defaultValue = "",
}: DeliveryLocationPickerProps) {
  const [query, setQuery] = useState(defaultValue);
  const [results, setResults] = useState<LocationResult[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isSearching, setIsSearching] = useState(false);
  const [selectedLocation, setSelectedLocation] = useState<LocationResult | null>(null);
  const [routeInfo, setRouteInfo] = useState<{ distance: number; duration: number } | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<NodeJS.Timeout>(undefined);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const searchAddress = useCallback(async (q: string) => {
    if (!q || q.length < 3) {
      setResults([]);
      return;
    }
    setIsSearching(true);
    try {
      const data = await deliveryApi.searchAddress(q);
      setResults(data || []);
      setIsOpen(true);
    } catch {
      setResults([]);
    } finally {
      setIsSearching(false);
    }
  }, []);

  const handleInputChange = (value: string) => {
    setQuery(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => searchAddress(value), 400);
  };

  const calculateRoute = useCallback(
    async (dropoff: LocationResult) => {
      if (!pickupLocation) return;
      try {
        const route = await deliveryApi.calculateRoute(
          pickupLocation.lat,
          pickupLocation.lng,
          dropoff.lat,
          dropoff.lng
        );
        setRouteInfo({ distance: route.distance_meters, duration: route.duration_seconds });
        onRouteCalculated?.(route.distance_meters, route.duration_seconds);
      } catch {
        setRouteInfo(null);
      }
    },
    [pickupLocation, onRouteCalculated]
  );

  const handleSelectLocation = (loc: LocationResult) => {
    setSelectedLocation(loc);
    setQuery(loc.address);
    setIsOpen(false);
    onLocationSelect(loc);
    if (pickupLocation) {
      calculateRoute(loc);
    }
  };

  const clearSelection = () => {
    setSelectedLocation(null);
    setQuery("");
    setRouteInfo(null);
    setResults([]);
    inputRef.current?.focus();
  };

  const formatDistance = (meters: number) => {
    if (meters >= 1000) return `${(meters / 1000).toFixed(1)} km`;
    return `${meters} m`;
  };

  const formatDuration = (seconds: number) => {
    if (seconds >= 3600) {
      const h = Math.floor(seconds / 3600);
      const m = Math.floor((seconds % 3600) / 60);
      return `${h}h ${m}p`;
    }
    return `${Math.ceil(seconds / 60)} phút`;
  };

  return (
    <div className="relative" ref={dropdownRef}>
      {label && (
        <label className="flex items-center gap-1.5 text-sm font-medium text-tiki-text mb-1.5">
          <MapPin className="w-4 h-4 text-tiki-red" />
          {label}
        </label>
      )}

      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => handleInputChange(e.target.value)}
          onFocus={() => results.length > 0 && setIsOpen(true)}
          placeholder={placeholder}
          className="w-full pl-10 pr-10 py-2.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue"
        />
        {selectedLocation && (
          <button
            onClick={clearSelection}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
          >
            <X className="w-4 h-4" />
          </button>
        )}
        {isSearching && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2">
            <div className="w-4 h-4 border-2 border-tiki-blue border-t-transparent rounded-full animate-spin" />
          </div>
        )}
      </div>

      {/* Selected location + route info */}
      {selectedLocation && routeInfo && (
        <div className="mt-2 p-3 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="flex items-center gap-2 text-sm">
            <Navigation className="w-4 h-4 text-tiki-blue" />
            <span className="text-tiki-text font-medium">{formatDistance(routeInfo.distance)}</span>
            <span className="text-gray-300">•</span>
            <Clock className="w-4 h-4 text-tiki-blue" />
            <span className="text-tiki-text">~{formatDuration(routeInfo.duration)}</span>
            <span className="text-gray-300">•</span>
            <Truck className="w-4 h-4 text-tiki-green" />
            <span className="text-tiki-green font-medium">
              {routeInfo.distance < 2000 ? "Giao nhanh" : "Giao tiêu chuẩn"}
            </span>
          </div>
          <p className="text-xs text-tiki-text-secondary mt-1 truncate">{selectedLocation.address}</p>
        </div>
      )}

      {selectedLocation && !routeInfo && pickupLocation && (
        <div className="mt-2 p-3 bg-gray-50 border border-gray-200 rounded-lg">
          <p className="text-xs text-tiki-text-secondary truncate">{selectedLocation.address}</p>
          <p className="text-xs text-gray-400 mt-0.5">Đang tính tuyến đường...</p>
        </div>
      )}

      {selectedLocation && !pickupLocation && (
        <div className="mt-2 p-3 bg-gray-50 border border-gray-200 rounded-lg">
          <p className="text-xs text-tiki-text-secondary truncate">{selectedLocation.address}</p>
        </div>
      )}

      {/* Dropdown results */}
      {isOpen && results.length > 0 && (
        <div className="absolute z-50 top-full left-0 right-0 mt-1 bg-white border border-gray-200 rounded-lg shadow-lg max-h-80 overflow-y-auto">
          {results.map((loc, i) => (
            <button
              key={`${loc.lat}-${loc.lng}-${i}`}
              onClick={() => handleSelectLocation(loc)}
              className="w-full text-left px-4 py-3 hover:bg-blue-50 transition flex items-start gap-3 border-b border-gray-100 last:border-0"
            >
              <MapPin className="w-4 h-4 text-gray-400 mt-0.5 shrink-0" />
              <div className="flex-1 min-w-0">
                <p className="text-sm text-tiki-text font-medium truncate">{loc.name || loc.address}</p>
                <p className="text-xs text-tiki-text-secondary truncate">{loc.address}</p>
              </div>
              <ChevronRight className="w-4 h-4 text-gray-300 mt-0.5 shrink-0" />
            </button>
          ))}
        </div>
      )}

      {isOpen && query.length >= 3 && results.length === 0 && !isSearching && (
        <div className="absolute z-50 top-full left-0 right-0 mt-1 bg-white border border-gray-200 rounded-lg shadow-lg p-4">
          <p className="text-sm text-tiki-text-secondary text-center">Không tìm thấy địa chỉ</p>
          <p className="text-xs text-gray-400 text-center mt-1">Thử tìm kiếm với từ khóa khác</p>
        </div>
      )}
    </div>
  );
}
