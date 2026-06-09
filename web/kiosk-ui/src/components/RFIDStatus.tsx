import React from 'react';
import { CreditCard } from 'lucide-react';

export const RFIDStatus: React.FC = () => {
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '40px',
      gap: '24px',
      textAlign: 'center'
    }}>
      {/* Pulsing ring animation */}
      <div style={{
        position: 'relative',
        width: '120px',
        height: '120px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center'
      }}>
        {/* Pulsing rings */}
        <div style={{
          position: 'absolute',
          width: '100%',
          height: '100%',
          borderRadius: '50%',
          border: '2px solid var(--primary)',
          opacity: 0.8,
          animation: 'pulse-ring 2s infinite ease-in-out'
        }} />
        <div style={{
          position: 'absolute',
          width: '80%',
          height: '80%',
          borderRadius: '50%',
          border: '2px solid var(--secondary)',
          opacity: 0.6,
          animation: 'pulse-ring 2s infinite ease-in-out',
          animationDelay: '0.6s'
        }} />
        
        {/* Central Card Icon */}
        <div style={{
          position: 'relative',
          width: '70px',
          height: '70px',
          borderRadius: '50%',
          background: 'linear-gradient(135deg, var(--primary), var(--secondary))',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: '0 0 20px var(--primary-glow)',
          zIndex: 1
        }}>
          <CreditCard size={32} style={{ color: 'var(--text-dark)' }} />
        </div>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: 700 }}>TEMPELKAN KTP ELEKTRONIK ANDA</h2>
        <p style={{ color: 'var(--text-muted)', fontSize: '16px', maxWidth: '340px', margin: '0 auto' }}>
          Letakkan KTP-el Anda pada mesin RFID reader di bawah layar untuk mengisi data secara otomatis.
        </p>
      </div>
    </div>
  );
};
