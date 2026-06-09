import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useRFID } from '../hooks/useRFID';
import { useWarga } from '../hooks/useWarga';
import type { Warga } from '../hooks/useWarga';
import { RFIDStatus } from '../components/RFIDStatus';
import { NIKInput } from '../components/NIKInput';
import { Keyboard, CreditCard, AlertTriangle, ArrowLeft } from 'lucide-react';

interface HomePageProps {
  onWargaIdentified: (w: Warga) => void;
}

export const HomePage: React.FC<HomePageProps> = ({ onWargaIdentified }) => {
  const navigate = useNavigate();
  const [mode, setMode] = useState<'rfid' | 'nik'>('rfid');
  const { lookupByRFID, lookupByNIK, loading, error, clearWarga } = useWarga();

  // Listen to RFID scan events (both wedge and SSE)
  useRFID(async (uid) => {
    // Only trigger RFID lookup when in RFID mode and not currently loading or showing an error
    if (mode === 'rfid' && !loading && !error) {
      console.log("Processing RFID scan for UID:", uid);
      const wargaData = await lookupByRFID(uid);
      if (wargaData) {
        onWargaIdentified(wargaData);
        navigate('/select-surat');
      }
    }
  });

  const handleNIKSubmit = async (nik: string) => {
    const wargaData = await lookupByNIK(nik);
    if (wargaData) {
      onWargaIdentified(wargaData);
      navigate('/select-surat');
    }
  };

  const handleRetry = () => {
    clearWarga();
  };

  return (
    <div className="page-container" style={{ justifyContent: 'center', alignItems: 'center' }}>
      
      {/* 1. Error State Screen */}
      {error ? (
        <div className="glass-card" style={{
          maxWidth: '500px',
          width: '100%',
          padding: '40px',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: '24px',
          textAlign: 'center',
          borderColor: 'var(--danger)',
          boxShadow: '0 8px 32px rgba(255, 50, 50, 0.15)'
        }}>
          <div style={{
            width: '80px',
            height: '80px',
            borderRadius: '50%',
            background: 'rgba(255, 50, 50, 0.1)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--danger)'
          }}>
            <AlertTriangle size={40} />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <h2 style={{ fontSize: '24px', fontWeight: 700 }}>Data Tidak Ditemukan</h2>
            <p style={{ color: 'var(--text-muted)', fontSize: '16px', lineHeight: 1.5 }}>
              {error}
            </p>
          </div>
          <button onClick={handleRetry} className="btn btn-secondary" style={{ width: '100%' }}>
            <ArrowLeft size={18} />
            Kembali dan Coba Lagi
          </button>
        </div>
      ) : (
        /* 2. Standard Search Screen */
        <div className="glass-card" style={{
          width: '100%',
          maxWidth: '640px',
          padding: '40px',
          display: 'flex',
          flexDirection: 'column',
          gap: '32px'
        }}>
          {loading ? (
            /* Loading State spinner */
            <div style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              padding: '60px',
              gap: '20px'
            }}>
              <div style={{
                width: '50px',
                height: '50px',
                border: '4px solid var(--border-color)',
                borderTopColor: 'var(--primary)',
                borderRadius: '50%',
                animation: 'pulse-ring 1s infinite linear'
              }} />
              <p style={{ color: 'var(--text-muted)', fontWeight: 500 }}>Mencari data warga...</p>
            </div>
          ) : (
            <>
              {/* Toggle Mode Button */}
              <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                {mode === 'rfid' ? (
                  <button onClick={() => setMode('nik')} className="btn btn-secondary" style={{ padding: '8px 16px', minHeight: '40px' }}>
                    <Keyboard size={16} />
                    Masukkan NIK Manual
                  </button>
                ) : (
                  <button onClick={() => setMode('rfid')} className="btn btn-secondary" style={{ padding: '8px 16px', minHeight: '40px' }}>
                    <CreditCard size={16} />
                    Gunakan Scan KTP
                  </button>
                )}
              </div>

              {/* Render Mode Screen */}
              {mode === 'rfid' ? (
                <RFIDStatus />
              ) : (
                <NIKInput onSubmit={handleNIKSubmit} loading={loading} />
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
};
