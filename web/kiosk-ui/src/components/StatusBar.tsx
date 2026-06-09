import React, { useState, useEffect } from 'react';
import { useOnlineStatus } from '../hooks/useOnlineStatus';
import { Wifi, WifiOff, RefreshCw, Cpu } from 'lucide-react';

export const StatusBar: React.FC = () => {
  const { isOnline, isBackendConnected, lastSyncTime, kioskName } = useOnlineStatus();
  const [timeStr, setTimeStr] = useState<string>('');

  useEffect(() => {
    const updateTime = () => {
      const now = new Date();
      setTimeStr(now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) + ' WIB');
    };
    updateTime();
    const interval = setInterval(updateTime, 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="status-bar glass-card">
      <div className="status-item">
        <Cpu size={16} className={isBackendConnected ? "text-success" : "text-danger"} style={{ color: isBackendConnected ? 'var(--success)' : 'var(--danger)' }} />
        <span>Backend Kiosk: <strong>{isBackendConnected ? "TERHUBUNG" : "TERPUTUS"}</strong></span>
      </div>

      <div className="status-item">
        {isOnline ? (
          <Wifi size={16} style={{ color: 'var(--success)' }} />
        ) : (
          <WifiOff size={16} style={{ color: 'var(--warning)' }} />
        )}
        <span>Server Pusat: <strong>{isOnline ? "ONLINE" : "OFFLINE"}</strong></span>
      </div>

      <div className="status-item">
        <RefreshCw size={14} style={{ color: 'var(--text-muted)' }} />
        <span>Sinkronisasi: <strong>{lastSyncTime}</strong></span>
      </div>

      <div style={{ marginLeft: 'auto', fontWeight: 600 }}>
        {kioskName} | {timeStr}
      </div>
    </div>
  );
};
