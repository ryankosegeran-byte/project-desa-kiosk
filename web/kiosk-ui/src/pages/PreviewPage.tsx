import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSurat } from '../hooks/useSurat';
import type { JenisSurat } from '../hooks/useSurat';
import type { Warga } from '../hooks/useWarga';
import { Printer, ArrowLeft, RefreshCw } from 'lucide-react';

interface PreviewPageProps {
  warga: Warga | null;
  jenisSurat: JenisSurat | null;
  formData: any;
  onSuratPrinted: () => void;
}

export const PreviewPage: React.FC<PreviewPageProps> = ({
  warga,
  jenisSurat,
  formData,
  onSuratPrinted,
}) => {
  const navigate = useNavigate();
  const { createSurat, printSurat, fetchTemplateHTML, loading, error, setError } = useSurat();
  const [templateHTML, setTemplateHTML] = useState<string>('');
  const [parsedHTML, setParsedHTML] = useState<string>('');

  useEffect(() => {
    if (!warga) {
      navigate('/');
      return;
    }
    if (!jenisSurat) {
      navigate('/select-surat');
      return;
    }

    // Load template HTML
    const loadTemplate = async () => {
      const html = await fetchTemplateHTML(jenisSurat.id);
      setTemplateHTML(html);
    };

    loadTemplate();
  }, [warga, jenisSurat, fetchTemplateHTML, navigate]);

  // Parse HTML templates client-side (replicate Go HTML templates)
  useEffect(() => {
    if (!templateHTML || !warga) return;

    let parsed = templateHTML;

    // 1. Format today's date in Indonesian format
    const months = [
      'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
      'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
    ];
    const today = new Date();
    const formattedDate = `${today.getDate()} ${months[today.getMonth()]} ${today.getFullYear()}`;

    // 2. Replacements
    parsed = parsed.replace(/\{\{\.DateToday\}\}/g, formattedDate);
    parsed = parsed.replace(/\{\{\.Warga\.Nama\}\}/g, warga.nama);
    parsed = parsed.replace(/\{\{\.Warga\.NIK\}\}/g, warga.nik);
    parsed = parsed.replace(/\{\{\.Warga\.TempatLahir\}\}/g, warga.tempat_lahir || '');
    parsed = parsed.replace(/\{\{\.Warga\.TanggalLahir\}\}/g, warga.tanggal_lahir || '');
    parsed = parsed.replace(/\{\{\.Warga\.JenisKelamin\}\}/g, warga.jenis_kelamin === 'L' ? 'Laki-laki' : 'Perempuan');
    parsed = parsed.replace(/\{\{\.Warga\.Pekerjaan\}\}/g, warga.pekerjaan || '');
    parsed = parsed.replace(/\{\{\.Warga\.Alamat\}\}/g, warga.alamat || '');
    parsed = parsed.replace(/\{\{\.Warga\.RT\}\}/g, warga.rt || '');
    parsed = parsed.replace(/\{\{\.Warga\.RW\}\}/g, warga.rw || '');
    parsed = parsed.replace(/\{\{\.Warga\.Kelurahan\}\}/g, warga.kelurahan || '');
    parsed = parsed.replace(/\{\{\.Warga\.Kecamatan\}\}/g, warga.kecamatan || '');

    // 3. Handle data_surat details table loop: {{range $key, $value := .DataSurat}}...{{end}}
    // Since our seeder has a standard loop structure:
    // <table ...> {{range $key, $value := .DataSurat}} <tr> ... </tr> {{end}} </table>
    // We can replace the entire loop block with generated HTML table rows.
    const loopRegex = /\{\{range\s+\$key,\s+\$value\s+:=\s+\.DataSurat\}\}([\s\S]*?)\{\{end\}\}/g;
    
    if (loopRegex.test(parsed)) {
      parsed = parsed.replace(loopRegex, (_, rowTemplate) => {
        let rowsHtml = '';
        
        // Find human readable label for each key from schema
        const getLabel = (key: string) => {
          const field = jenisSurat?.fields_schema.fields.find(f => f.key === key);
          return field ? field.label : key;
        };

        // Render each field value
        Object.entries(formData).forEach(([key, val]) => {
          if (val === undefined || val === null || val === '') return;
          
          let row = rowTemplate;

          // Handle repeater field custom display
          if (Array.isArray(val)) {
            let listHtml = '<ol style="padding-left: 16px;">';
            val.forEach((item: any) => {
              const itemDetails = Object.entries(item)
                .map(([, subVal]) => `${subVal}`)
                .join(' - ');
              listHtml += `<li>${itemDetails}</li>`;
            });
            listHtml += '</ol>';
            
            row = row.replace(/\{\{\$key\}\}/g, getLabel(key));
            row = row.replace(/\{\{\$value\}\}/g, listHtml);
          } else {
            row = row.replace(/\{\{\$key\}\}/g, getLabel(key));
            row = row.replace(/\{\{\$value\}\}/g, String(val));
          }

          rowsHtml += row;
        });

        return rowsHtml;
      });
    }

    setParsedHTML(parsed);
  }, [templateHTML, warga, formData, jenisSurat]);

  const handlePrint = async () => {
    if (!warga || !jenisSurat) return;

    setError(null);
    try {
      // 1. Create Surat Draft in backend
      const surat = await createSurat({
        jenis_surat_id: jenisSurat.id,
        warga_id: warga.id,
        nik_pemohon: warga.nik,
        nama_pemohon: warga.nama,
        data_surat: formData
      });

      if (!surat) {
        throw new Error("Gagal menyimpan draf surat ke kiosk lokal.");
      }

      // 2. Call print endpoint
      const printed = await printSurat(surat.id);
      if (!printed) {
        throw new Error("Gagal mengirim perintah cetak ke printer.");
      }

      // 3. Trigger print success callback
      onSuratPrinted();
      navigate('/success');
    } catch (err: any) {
      setError(err.message || "Terjadi kesalahan saat memproses cetak.");
    }
  };

  const handleBack = () => {
    navigate('/form-surat');
  };

  if (!warga || !jenisSurat) return null;

  return (
    <div className="page-container" style={{ height: '100%' }}>
      <div style={{ marginBottom: '20px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: 800 }}>Pratinjau Surat</h2>
        <p style={{ color: 'var(--text-muted)', fontSize: '15px' }}>
          Periksa kembali format dan keselarasan isi surat sebelum dicetak.
        </p>
      </div>

      {error && (
        <div className="glass-card" style={{ padding: '16px', color: 'var(--danger)', marginBottom: '16px', fontWeight: 500 }}>
          Terjadi Kesalahan: {error}
        </div>
      )}

      {/* Embedded Document Preview container */}
      <div className="glass-card" style={{
        flex: 1,
        background: '#ffffff',
        borderRadius: 'var(--radius-md)',
        overflow: 'hidden',
        display: 'flex',
        flexDirection: 'column',
        marginBottom: '20px',
        boxShadow: 'inset 0 0 10px rgba(0,0,0,0.1)'
      }}>
        {/* Document Content viewport */}
        <div style={{
          flex: 1,
          overflowY: 'auto',
          padding: '40px',
          display: 'flex',
          justifyContent: 'center'
        }}>
          {parsedHTML ? (
            <div 
              style={{
                width: '100%',
                maxWidth: '700px',
                color: '#000000',
                background: '#ffffff'
              }}
              dangerouslySetInnerHTML={{ __html: parsedHTML }}
            />
          ) : (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-dark)' }}>
              Memuat pratinjau...
            </div>
          )}
        </div>
      </div>

      {/* Action footer */}
      <div style={{
        display: 'flex',
        gap: '16px',
        paddingTop: '16px',
        borderTop: '1px solid var(--border-color)'
      }}>
        <button type="button" disabled={loading} className="btn btn-secondary" onClick={handleBack} style={{ flex: 1 }}>
          <ArrowLeft size={18} />
          Kembali Edit
        </button>
        <button
          type="button"
          disabled={loading || !parsedHTML}
          className={`btn btn-primary ${(loading || !parsedHTML) ? 'btn-disabled' : ''}`}
          onClick={handlePrint}
          style={{ flex: 2 }}
        >
          {loading ? (
            <>
              <RefreshCw size={20} className="spinner" style={{ animation: 'pulse-ring 1s infinite linear' }} />
              Sedang Mencetak...
            </>
          ) : (
            <>
              <Printer size={20} />
              Cetak Surat Sekarang (A4)
            </>
          )}
        </button>
      </div>
    </div>
  );
};
