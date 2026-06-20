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
  const { createSurat, printSurat, fetchTemplate, previewSuratPDF, loading, error, setError } = useSurat();

  const [mode, setMode] = useState<'loading' | 'docx' | 'html'>('loading');
  const [pdfUrl, setPdfUrl] = useState<string>('');
  const [rendering, setRendering] = useState<boolean>(false);
  const [templateHTML, setTemplateHTML] = useState<string>('');
  const [parsedHTML, setParsedHTML] = useState<string>('');
  const [selectedFormat, setSelectedFormat] = useState<string>('A4');

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

  // Load template and decide DOCX (server-rendered PDF) vs HTML (client preview).
  useEffect(() => {
    if (!warga || !jenisSurat) return;
    let active = true;
    let createdUrl = '';
    (async () => {
      const tpl = await fetchTemplate(jenisSurat.id);
      if (!active) return;
      if (tpl) setSelectedFormat(tpl.format_kertas || 'A4');

      if (tpl && tpl.placeholders && tpl.placeholders.length > 0) {
        // DOCX (Strategi B): rendered to PDF by the kiosk (Word) — 100% sesuai cetak.
        setMode('docx');
        setRendering(true);
        const url = await previewSuratPDF({
          jenis_surat_id: jenisSurat.id,
          nik: warga.nik,
          data_surat: formData,
        });
        if (!active) return;
        if (url) {
          createdUrl = url;
          setPdfUrl(url);
        }
        setRendering(false);
      } else {
        setTemplateHTML(tpl?.template_html || '');
        setMode('html');
      }
    })();
    return () => {
      active = false;
      if (createdUrl) URL.revokeObjectURL(createdUrl);
    };
  }, [warga, jenisSurat, formData, fetchTemplate, previewSuratPDF]);

  // Client-side HTML preview (only for legacy HTML templates).
  useEffect(() => {
    if (mode !== 'html' || !templateHTML || !warga) return;

    let parsed = templateHTML;
    const months = [
      'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
      'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
    ];
    const today = new Date();
    const formattedDate = `${today.getDate()} ${months[today.getMonth()]} ${today.getFullYear()}`;

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

    setParsedHTML(parsed);
  }, [mode, templateHTML, warga, formData]);

  const handlePrint = async () => {
    if (!warga || !jenisSurat) return;
    setError(null);
    try {
      const surat = await createSurat({
        jenis_surat_id: jenisSurat.id,
        warga_id: warga.id,
        nik_pemohon: warga.nik,
        nama_pemohon: warga.nama,
        data_surat: formData,
      });
      if (!surat) throw new Error("Gagal menyimpan draf surat ke kiosk lokal.");

      const printed = await printSurat(surat.id);
      if (!printed) throw new Error("Gagal mengirim perintah cetak ke printer.");

      onSuratPrinted();
      navigate('/success');
    } catch (err: any) {
      setError(err.message || "Terjadi kesalahan saat memproses cetak.");
    }
  };

  const handleBack = () => navigate('/form-surat');

  if (!warga || !jenisSurat) return null;

  const canPrint = mode === 'docx' ? (!rendering && !!pdfUrl) : !!parsedHTML;

  return (
    <div className="page-container" style={{ height: '100%' }}>
      <div style={{ marginBottom: '20px' }}>
        <h2 style={{ fontSize: '24px', fontWeight: 800 }}>Pratinjau Surat</h2>
        <p style={{ color: 'var(--text-muted)', fontSize: '15px' }}>
          Periksa kembali isi surat sebelum dicetak. Tampilan ini sama persis dengan hasil cetak.
        </p>
      </div>

      {error && (
        <div className="glass-card" style={{ padding: '16px', color: 'var(--danger)', marginBottom: '16px', fontWeight: 500 }}>
          Terjadi Kesalahan: {error}
        </div>
      )}

      <div className="glass-card" style={{
        flex: 1,
        background: '#ffffff',
        borderRadius: 'var(--radius-md)',
        overflow: 'hidden',
        display: 'flex',
        flexDirection: 'column',
        marginBottom: '20px',
        boxShadow: 'inset 0 0 10px rgba(0,0,0,0.1)',
      }}>
        {mode === 'docx' ? (
          rendering ? (
            <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-dark)', gap: '10px' }}>
              <RefreshCw size={20} className="spinner" /> Merender dokumen...
            </div>
          ) : pdfUrl ? (
            <iframe title="Pratinjau Surat" src={pdfUrl} style={{ flex: 1, width: '100%', border: 'none' }} />
          ) : (
            <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-dark)' }}>
              Gagal memuat pratinjau.
            </div>
          )
        ) : mode === 'html' ? (
          <div style={{ flex: 1, overflowY: 'auto', padding: '40px', display: 'flex', justifyContent: 'center' }}>
            {parsedHTML ? (
              <div
                style={{ width: '100%', maxWidth: selectedFormat === 'F4' ? '720px' : '700px', color: '#000', background: '#fff' }}
                dangerouslySetInnerHTML={{ __html: parsedHTML }}
              />
            ) : (
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-dark)' }}>Memuat pratinjau...</div>
            )}
          </div>
        ) : (
          <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-dark)' }}>Memuat pratinjau...</div>
        )}
      </div>

      <div style={{ display: 'flex', gap: '16px', paddingTop: '16px', borderTop: '1px solid var(--border-color)' }}>
        <button type="button" disabled={loading} className="btn btn-secondary" onClick={handleBack} style={{ flex: 1 }}>
          <ArrowLeft size={18} />
          Kembali Edit
        </button>
        <button
          type="button"
          disabled={loading || !canPrint}
          className={`btn btn-primary ${(loading || !canPrint) ? 'btn-disabled' : ''}`}
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
              Cetak Surat
            </>
          )}
        </button>
      </div>
    </div>
  );
};
