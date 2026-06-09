import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import type { JenisSurat } from '../hooks/useSurat';
import type { Warga } from '../hooks/useWarga';
import { SuratForm } from '../components/SuratForm';

interface FormSuratPageProps {
  warga: Warga | null;
  jenisSurat: JenisSurat | null;
  onFormSubmitted: (formData: any) => void;
}

export const FormSuratPage: React.FC<FormSuratPageProps> = ({
  warga,
  jenisSurat,
  onFormSubmitted,
}) => {
  const navigate = useNavigate();

  useEffect(() => {
    if (!warga) {
      navigate('/');
      return;
    }
    if (!jenisSurat) {
      navigate('/select-surat');
      return;
    }
  }, [warga, jenisSurat, navigate]);

  const handleSubmit = (data: any) => {
    onFormSubmitted(data);
    navigate('/preview');
  };

  const handleBack = () => {
    navigate('/select-surat');
  };

  if (!warga || !jenisSurat) return null;

  return (
    <div className="page-container">
      {/* Title block */}
      <div style={{ marginBottom: '24px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: 800 }}>Isi Formulir Surat</h2>
        <p style={{ color: 'var(--text-muted)', fontSize: '15px' }}>
          Melengkapi informasi tambahan untuk pembuatan <strong>{jenisSurat.nama}</strong>.
        </p>
      </div>

      {/* Dynamic Form wrapper */}
      <div className="glass-card" style={{ padding: '28px', flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <SuratForm
          fields={jenisSurat.fields_schema.fields}
          warga={warga}
          onSubmit={handleSubmit}
          onBack={handleBack}
        />
      </div>
    </div>
  );
};
