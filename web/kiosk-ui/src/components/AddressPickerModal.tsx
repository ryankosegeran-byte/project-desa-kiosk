import React, { useState, useEffect, useRef } from 'react';
import { X, MapPin, Keyboard, Check, Loader } from 'lucide-react';
import { FullKeyboard } from './FullKeyboard';

interface AddressPickerModalProps {
  isOpen: boolean;
  currentValue: string;
  onSelect: (address: string) => void;
  onClose: () => void;
}

export const AddressPickerModal: React.FC<AddressPickerModalProps> = ({
  isOpen,
  currentValue,
  onSelect,
  onClose,
}) => {
  const [mode, setMode] = useState<'map' | 'manual'>('map');
  const [address, setAddress] = useState(currentValue || '');
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [isLoadingGeocode, setIsLoadingGeocode] = useState(false);
  const [, setMarker] = useState<{ lat: number; lng: number } | null>(null);
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<any>(null);
  const markerRef = useRef<any>(null);

  // Check online status
  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => { setIsOnline(false); setMode('manual'); };
    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);
    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  // Auto-switch to manual if offline
  useEffect(() => {
    if (!isOnline && mode === 'map') {
      setMode('manual');
    }
  }, [isOnline, mode]);

  // Initialize Leaflet map
  useEffect(() => {
    if (!isOpen || mode !== 'map' || !mapContainerRef.current) return;

    // Dynamic import of Leaflet to avoid SSR issues
    const initMap = async () => {
      const L = await import('leaflet');
      await import('leaflet/dist/leaflet.css');

      // Fix default marker icon issue with bundlers
      delete (L.Icon.Default.prototype as any)._getIconUrl;
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
        iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
        shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
      });

      if (mapRef.current) {
        mapRef.current.remove();
      }

      // Default center: Kalawat, Minahasa Utara
      const defaultCenter: [number, number] = [1.4748, 124.9870];
      const map = L.map(mapContainerRef.current!, {
        center: defaultCenter,
        zoom: 15,
        zoomControl: true,
      });

      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap',
      }).addTo(map);

      // Click handler for pin placement
      map.on('click', async (e: any) => {
        const { lat, lng } = e.latlng;
        setMarker({ lat, lng });

        // Update or create marker
        if (markerRef.current) {
          markerRef.current.setLatLng([lat, lng]);
        } else {
          markerRef.current = L.marker([lat, lng]).addTo(map);
        }

        // Reverse geocode
        setIsLoadingGeocode(true);
        try {
          const res = await fetch(
            `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}&zoom=18&addressdetails=1`,
            { headers: { 'Accept-Language': 'id' } }
          );
          if (res.ok) {
            const data = await res.json();
            const addr = data.display_name || `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
            setAddress(addr);
          }
        } catch {
          setAddress(`${lat.toFixed(6)}, ${lng.toFixed(6)}`);
        } finally {
          setIsLoadingGeocode(false);
        }
      });

      mapRef.current = map;

      // Force map to recalculate size after render
      setTimeout(() => map.invalidateSize(), 200);
    };

    initMap();

    return () => {
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
        markerRef.current = null;
      }
    };
  }, [isOpen, mode]);

  if (!isOpen) return null;

  const handleConfirm = () => {
    if (address.trim()) {
      onSelect(address.trim());
      onClose();
    }
  };

  const handleKeyPress = (key: string) => {
    setAddress(prev => prev + key);
  };

  const handleDelete = () => {
    setAddress(prev => prev.slice(0, -1));
  };

  const handleClear = () => {
    setAddress('');
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-container modal-large" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="modal-header">
          <h3><MapPin size={20} /> Pilih Alamat Usaha</h3>
          <button className="modal-close-btn" onClick={onClose}>
            <X size={24} />
          </button>
        </div>

        {/* Mode Tabs */}
        <div className="modal-tabs">
          <button
            className={`modal-tab ${mode === 'map' ? 'active' : ''}`}
            onClick={() => isOnline && setMode('map')}
            disabled={!isOnline}
          >
            <MapPin size={16} />
            Peta {!isOnline && '(Offline)'}
          </button>
          <button
            className={`modal-tab ${mode === 'manual' ? 'active' : ''}`}
            onClick={() => setMode('manual')}
          >
            <Keyboard size={16} />
            Ketik Manual
          </button>
        </div>

        <div className="modal-body">
          {mode === 'map' ? (
            <>
              {/* Map Container */}
              <div
                ref={mapContainerRef}
                style={{ width: '100%', height: '350px', borderRadius: 'var(--radius-sm)', overflow: 'hidden', marginBottom: '16px' }}
              />
              <p style={{ fontSize: '13px', color: 'var(--text-muted)', marginBottom: '12px' }}>
                Tap pada peta untuk menaruh pin lokasi usaha. Alamat akan terisi otomatis.
              </p>
              {/* Address Preview */}
              <div className="custom-input-preview">
                <span className="custom-input-label">Alamat:</span>
                <div className="custom-input-display">
                  {isLoadingGeocode ? (
                    <span style={{ display: 'flex', alignItems: 'center', gap: '8px', color: 'var(--text-muted)' }}>
                      <Loader size={16} className="spin-animation" /> Mencari alamat...
                    </span>
                  ) : (
                    address || <span style={{ color: 'var(--text-muted)' }}>Belum ada pin...</span>
                  )}
                </div>
              </div>
            </>
          ) : (
            <>
              {/* Manual Input */}
              <div className="custom-input-preview">
                <span className="custom-input-label">Ketik Alamat Usaha:</span>
                <div className="custom-input-display">
                  {address || <span style={{ color: 'var(--text-muted)' }}>Ketik di sini...</span>}
                  <span className="cursor-blink">|</span>
                </div>
              </div>
              <FullKeyboard
                onKeyPress={handleKeyPress}
                onDelete={handleDelete}
                onClear={handleClear}
                onEnter={handleConfirm}
              />
            </>
          )}

          {/* Confirm Button */}
          <button
            className="btn btn-primary"
            style={{ width: '100%', marginTop: '16px' }}
            onClick={handleConfirm}
            disabled={!address.trim()}
          >
            <Check size={18} />
            Pilih Alamat Ini
          </button>
        </div>
      </div>
    </div>
  );
};
