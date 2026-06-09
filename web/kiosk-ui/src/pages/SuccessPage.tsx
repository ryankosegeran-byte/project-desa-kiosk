import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { CheckCircle2, Printer, Home } from 'lucide-react';

interface SuccessPageProps {
  onReset: () => void;
}

export const SuccessPage: React.FC<SuccessPageProps> = ({ onReset }) => {
  const navigate = useNavigate();
  const [countdown, setCountdown] = useState<number>(10);

  useEffect(() => {
    // 10-second countdown to reset and redirect to home screen
    const timer = setInterval(() => {
      setCountdown(prev => {
        if (prev <= 1) {
          clearInterval(timer);
          onReset();
          navigate('/');
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [onReset, navigate]);

  const handleGoHome = () => {
    onReset();
    navigate('/');
  };

  return (
    <div className="page-container" style={{ justifyContent: 'center', alignItems: 'center' }}>
      <div className="glass-card" style={{
        maxWidth: '540px',
        width: '100%',
        padding: '50px 40px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '32px',
        textAlign: 'center',
        boxShadow: '0 16px 48px rgba(0, 220, 140, 0.15)',
        borderColor: 'var(--success)'
      }}>
        {/* Success Icon Animation container */}
        <div style={{
          position: 'relative',
          width: '100px',
          height: '100px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center'
        }}>
          {/* Pulsing Success Ring */}
          <div style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            borderRadius: '50%',
            border: '3px solid var(--success)',
            animation: 'pulse-ring 2s infinite ease-in-out'
          }} />
          
          <div style={{
            width: '80px',
            height: '80px',
            borderRadius: '50%',
            background: 'linear-gradient(135deg, var(--success), hsl(145, 90%, 55%))',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 0 20px var(--success-glow)',
            color: 'var(--text-dark)'
          }}>
            <CheckCircle2 size={46} />
          </div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          <h2 style={{ fontSize: '28px', fontWeight: 800 }}>Pencetakan Berhasil!</h2>
          <p style={{ color: 'var(--text-muted)', fontSize: '16px', lineHeight: 1.6 }}>
            Surat Anda sedang diproses oleh printer. Silakan ambil kertas hasil cetak di tray keluar printer.
          </p>
          <div style={{
            background: 'var(--bg-surface-glow)',
            border: '1px dashed var(--border-color)',
            borderRadius: 'var(--radius-sm)',
            padding: '12px',
            fontSize: '14px',
            color: 'var(--text-muted)',
            marginTop: '8px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '8px'
          }}>
            <Printer size={16} />
            <span>Format Cetakan: Kertas HVS A4 Standar</span>
          </div>
        </div>

        <div style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <button onClick={handleGoHome} className="btn btn-primary" style={{ width: '100%', height: '52px', fontSize: '16px' }}>
            <Home size={18} />
            Kembali ke Beranda Utama
          </button>
          
          <p style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
            Layar akan otomatis kembali ke beranda dalam <strong>{countdown} detik</strong>.
          </p>
        </div>
      </div>
    </div>
  );
};
