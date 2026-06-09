import { useState, useEffect } from 'react';

interface KioskStatus {
  isBackendConnected: boolean;
  lastSyncTime: string;
  desaID: string;
  kioskName: string;
}

export function useOnlineStatus() {
  const [isOnline, setIsOnline] = useState<boolean>(navigator.onLine);
  const [kioskStatus, setKioskStatus] = useState<KioskStatus>({
    isBackendConnected: false,
    lastSyncTime: 'Never',
    desaID: '',
    kioskName: 'Kiosk Desa'
  });

  const apiBase = import.meta.env.DEV ? 'http://localhost:8080' : '';

  useEffect(() => {
    // 1. Browser online/offline event listeners
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // 2. Ping local backend server status periodically
    const checkKioskBackendStatus = async () => {
      try {
        const res = await fetch(`${apiBase}/api/status`, {
          // Short timeout to avoid hanging
          signal: AbortSignal.timeout(3000)
        });
        if (res.ok) {
          const data = await res.json();
          setKioskStatus({
            isBackendConnected: true,
            lastSyncTime: data.last_sync || 'Never',
            desaID: data.desa_id || '',
            kioskName: data.kiosk_name || 'Kiosk Desa'
          });
        } else {
          throw new Error("Backend returned error status");
        }
      } catch (err) {
        setKioskStatus(prev => ({
          ...prev,
          isBackendConnected: false
        }));
      }
    };

    // Initial check
    checkKioskBackendStatus();

    // Check every 10 seconds
    const interval = setInterval(checkKioskBackendStatus, 10000);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
      clearInterval(interval);
    };
  }, [apiBase]);

  return {
    isOnline, // browser network connection (for cloud status indicator)
    ...kioskStatus // local backend connection state
  };
}
