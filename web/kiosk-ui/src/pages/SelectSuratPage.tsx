import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSurat } from '../hooks/useSurat';
import type { JenisSurat } from '../hooks/useSurat';
import type { Warga } from '../hooks/useWarga';
import { FileText, LogOut, ArrowRight } from 'lucide-react';

interface SelectSuratPageProps {
  warga: Warga | null;
  onSuratSelected: (js: JenisSurat) => void;
  onCancel: () => void;
}

export const SelectSuratPage: React.FC<SelectSuratPageProps> = ({
  warga,
  onSuratSelected,
  onCancel,
}) => {
  const navigate = useNavigate();
  const { jenisSuratList, fetchJenisSurat, loading, error } = useSurat();

  useEffect(() => {
    // If no citizen is identified yet, send back to home
    if (!warga) {
      navigate('/');
      return;
    }
    fetchJenisSurat();
  }, [warga, fetchJenisSurat, navigate]);

  const handleSelect = (js: JenisSurat) => {
    onSuratSelected(js);
    navigate('/form-surat');
  };

  const handleCancelClick = () => {
    onCancel();
    navigate('/');
  };

  if (!warga) return null;

  return (
    <div className="page-container">
      {/* Greet Warga */}
      <div style={{ marginBottom: '32px' }}>
        <h2 style={{ fontSize: '28px', fontWeight: 800, marginBottom: '6px' }}>
          Halo, <span style={{ color: 'var(--primary)' }}>{warga.nama}</span>!
        </h2>
        <p style={{ color: 'var(--text-muted)', fontSize: '16px' }}>
          Silakan pilih jenis surat keterangan yang ingin Anda buat hari ini.
        </p>
      </div>

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', flex: 1 }}>
          <div style={{
            width: '40px',
            height: '40px',
            border: '3px solid var(--border-color)',
            borderTopColor: 'var(--primary)',
            borderRadius: '50%',
            animation: 'pulse-ring 1s infinite linear'
          }} />
        </div>
      ) : error ? (
        <div className="glass-card" style={{ padding: '24px', color: 'var(--danger)', textAlign: 'center' }}>
          {error}
        </div>
      ) : (
        /* Letters Grid */
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
          gap: '20px',
          overflowY: 'auto',
          paddingBottom: '20px',
          flex: 1
        }}>
          {jenisSuratList.map((js) => (
            <div
              key={js.id}
              onClick={() => handleSelect(js)}
              className="glass-card interactive-card"
              style={{
                padding: '24px',
                display: 'flex',
                flexDirection: 'column',
                gap: '16px',
                height: '180px',
                justifyContent: 'space-between',
                position: 'relative',
                overflow: 'hidden'
              }}
            >
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <div style={{
                  width: '40px',
                  height: '40px',
                  borderRadius: '10px',
                  background: 'var(--bg-surface-glow)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: 'var(--primary)',
                  border: '1px solid var(--border-color)'
                }}>
                  <FileText size={20} />
                </div>
                <h3 style={{ fontSize: '18px', fontWeight: 700, lineHeight: 1.3 }}>{js.nama}</h3>
                <p style={{ fontSize: '13px', color: 'var(--text-muted)', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
                  {js.deskripsi || "Buat surat keterangan ini secara mandiri."}
                </p>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '14px', fontWeight: 600, color: 'var(--secondary)' }}>
                <span>Pilih Surat</span>
                <ArrowRight size={14} />
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Cancel Button */}
      <div style={{
        paddingTop: '20px',
        borderTop: '1px solid var(--border-color)',
        marginTop: '20px'
      }}>
        <button onClick={handleCancelClick} className="btn btn-secondary" style={{ width: '200px' }}>
          <LogOut size={16} />
          Selesai / Keluar
        </button>
      </div>
    </div>
  );
};
