import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSurat, placeholdersToFields } from '../hooks/useSurat';
import type { JenisSurat, FieldDef } from '../hooks/useSurat';
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
  const { fetchTemplate } = useSurat();
  const [fields, setFields] = useState<FieldDef[]>(jenisSurat?.fields_schema?.fields || []);
  const [loadingTpl, setLoadingTpl] = useState<boolean>(true);

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

  // Build the form from the template: DOCX templates (Strategi B) drive the form
  // from their manual placeholders; otherwise fall back to the jenis surat schema.
  useEffect(() => {
    if (!jenisSurat) return;
    let active = true;
    (async () => {
      setLoadingTpl(true);
      const tpl = await fetchTemplate(jenisSurat.id);
      if (!active) return;
      if (tpl && tpl.placeholders && tpl.placeholders.length > 0) {
        setFields(placeholdersToFields(tpl.placeholders));
      } else {
        setFields(jenisSurat.fields_schema?.fields || []);
      }
      setLoadingTpl(false);
    })();
    return () => {
      active = false;
    };
  }, [jenisSurat, fetchTemplate]);

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
      <div style={{ marginBottom: '24px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: 800 }}>Isi Formulir Surat</h2>
        <p style={{ color: 'var(--text-muted)', fontSize: '15px' }}>
          Melengkapi informasi tambahan untuk pembuatan <strong>{jenisSurat.nama}</strong>.
        </p>
      </div>

      <div className="glass-card" style={{ padding: '28px', flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {loadingTpl ? (
          <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-muted)' }}>
            Memuat formulir...
          </div>
        ) : (
          <SuratForm
            fields={fields}
            warga={warga}
            onSubmit={handleSubmit}
            onBack={handleBack}
          />
        )}
      </div>
    </div>
  );
};
