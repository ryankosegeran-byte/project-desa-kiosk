import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useRFID } from '../hooks/useRFID';
import { useWarga } from '../hooks/useWarga';
import { useRegistrationSession } from '../hooks/useRegistrationSession';
import type { Warga } from '../hooks/useWarga';
import { RFIDStatus } from '../components/RFIDStatus';
import { NIKInput } from '../components/NIKInput';
import { Keyboard, CreditCard, AlertTriangle, ArrowLeft, UserPlus, CheckCircle2 } from 'lucide-react';

interface HomePageProps {
  onWargaIdentified: (w: Warga) => void;
}

export const HomePage: React.FC<HomePageProps> = ({ onWargaIdentified }) => {
  const navigate = useNavigate();
  const [mode, setMode] = useState<'rfid' | 'nik'>('rfid');
  const { lookupByRFID, lookupByNIK, loading, error, clearWarga } = useWarga();

  // Registration session driven by the admin panel (via the online server).
  const registrationSession = useRegistrationSession();
  const registrationActive = registrationSession.active;

  // Reset the "scan sent" confirmation when the session ends.
  useEffect(() => {
    if (!registrationActive) setScanSent(false);
  }, [registrationActive]);
  const [scanSent, setScanSent] = useState(false);

  // Listen to RFID scan events (both wedge and SSE)
  useRFID(async (uid) => {
    // Registration mode: the scanned UID is forwarded to the server by the
    // kiosk backend automatically. Here we only show feedback, no local lookup.
    if (registrationActive) {
      console.log('Registration mode scan, UID forwarded to server (hidden from kiosk):', uid);
      setScanSent(true);
      return;
    }

    // Normal mode: look up the resident locally.
    if (mode === 'rfid' && !loading && !error) {
      console.log('Processing RFID scan for UID:', uid);
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

  // ============================================================
  // REGISTRATION MODE (operator is linking a card in admin panel)
  // ============================================================
  if (registrationActive) {
    return (
      <div className="page-container" style={{ justifyContent: 'center', alignItems: 'center', overflowY: 'auto', padding: '16px 0' }}>
        <div className="glass-card" style={{
          width: '100%',
          maxWidth: '560px',
          padding: '28px 32px',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: '16px',
          textAlign: 'center',
          borderColor: 'var(--primary)',
          boxShadow: '0 8px 40px rgba(56, 132, 255, 0.18)'
        }}>
          <div style={{
            width: '60px',
            height: '60px',
            borderRadius: '50%',
            background: 'rgba(56, 132, 255, 0.12)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--primary)',
            flexShrink: 0
          }}>
            <UserPlus size={30} />
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
            <span className="badge" style={{
              alignSelf: 'center',
              background: 'rgba(56, 132, 255, 0.15)',
              color: 'var(--primary)',
              fontWeight: 700,
              padding: '3px 12px',
              borderRadius: '999px',
              fontSize: '11px',
              letterSpacing: '0.5px'
            }}>
              MODE PENDAFTARAN
            </span>
            <h2 style={{ fontSize: '22px', fontWeight: 800 }}>Pendaftaran Warga Baru</h2>
            <p style={{ color: 'var(--text-muted)', fontSize: '14px', lineHeight: 1.5, margin: 0 }}>
              Operator sedang mendaftarkan warga. Tempelkan e-KTP pada alat pembaca.
            </p>
          </div>

          {registrationSession.name && (
            <div style={{
              width: '100%',
              padding: '14px 16px',
              background: 'rgba(56, 132, 255, 0.08)',
              border: '1px solid rgba(56, 132, 255, 0.35)',
              borderRadius: 'var(--radius-md)',
              textAlign: 'center'
            }}>
              <span style={{ fontSize: '12px', color: 'var(--text-muted)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px' }}>Warga yang Didaftarkan</span>
              <h3 style={{ fontSize: '22px', fontWeight: 800, marginTop: '4px', color: 'var(--text-main)' }}>{registrationSession.name}</h3>
            </div>
          )}

          {scanSent ? (
            <div style={{
              width: '100%',
              padding: '18px',
              background: 'rgba(34, 197, 94, 0.08)',
              border: '1px solid rgba(34, 197, 94, 0.4)',
              borderRadius: 'var(--radius-md)',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              gap: '8px'
            }}>
              <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', color: '#22c55e', fontWeight: 700, fontSize: '16px' }}>
                <CheckCircle2 size={20} />
                Kartu Berhasil Dibaca
              </div>
              <p style={{ color: 'var(--text-muted)', fontSize: '13px', margin: 0 }}>
                Data terkirim ke operator. Tunggu operator menyimpan, atau tempelkan kartu lain.
              </p>
            </div>
          ) : (
            <div style={{
              width: '100%',
              padding: '22px 20px',
              background: 'var(--bg-inset)',
              border: '2px dashed var(--border-color)',
              borderRadius: 'var(--radius-md)',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              gap: '14px'
            }}>
              <div style={{
                width: '40px',
                height: '40px',
                border: '4px solid var(--border-color)',
                borderTopColor: 'var(--primary)',
                borderRadius: '50%',
                animation: 'pulse-ring 1s infinite linear'
              }} />
              <p style={{ color: 'var(--text-muted)', fontWeight: 500, margin: 0 }}>Menunggu kartu KTP ditempelkan...</p>
            </div>
          )}
        </div>
      </div>
    );
  }

  // ============================================================
  // NORMAL MODE (resident self-service lookup)
  // ============================================================
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