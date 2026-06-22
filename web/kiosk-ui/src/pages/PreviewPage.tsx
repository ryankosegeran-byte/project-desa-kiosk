import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSurat } from '../hooks/useSurat';
import type { JenisSurat } from '../hooks/useSurat';
import type { Warga } from '../hooks/useWarga';
import { Printer, ArrowLeft, RefreshCw, ZoomIn, ZoomOut, Maximize2, X } from 'lucide-react';

interface PreviewPageProps {
  warga: Warga | null;
  jenisSurat: JenisSurat | null;
  formData: any;
  onSuratPrinted: () => void;
}

const ZOOM_MIN = 0.5;
const ZOOM_MAX = 3.0;
const ZOOM_STEP = 0.25;
const ZOOM_DEFAULT = 1.0;

export const PreviewPage: React.FC<PreviewPageProps> = ({
  warga,
  jenisSurat,
  formData,
  onSuratPrinted,
}) => {
  const navigate = useNavigate();
  const { createSurat, printSurat, previewSurat, loading, error, setError } = useSurat();

  const [mode, setMode] = useState<'loading' | 'pdf' | 'html' | 'error'>('loading');
  const [pdfUrl, setPdfUrl] = useState<string>('');
  const [previewHTML, setPreviewHTML] = useState<string>('');
  const [selectedFormat, setSelectedFormat] = useState<string>('A4');
  const [zoom, setZoom] = useState<number>(ZOOM_DEFAULT);
  const [expanded, setExpanded] = useState<boolean>(false);

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

  // Ask the kiosk to render a live preview. The kiosk uses the SAME engine as
  // printing (html/template or Word), so the preview equals the printed output.
  useEffect(() => {
    if (!warga || !jenisSurat) return;
    let active = true;
    let createdUrl = '';

    setMode('loading');
    setZoom(ZOOM_DEFAULT);
    setError(null);

    (async () => {
      const result = await previewSurat({
        jenis_surat_id: jenisSurat.id,
        nik: warga.nik,
        data_surat: formData,
      });
      if (!active) {
        if (result && result.mode === 'pdf') URL.revokeObjectURL(result.url);
        return;
      }
      if (!result) {
        setMode('error');
        return;
      }
      if (result.mode === 'pdf') {
        createdUrl = result.url;
        setPdfUrl(result.url);
        setMode('pdf');
      } else {
        setPreviewHTML(result.html);
        setSelectedFormat(result.format_kertas || 'A4');
        setMode('html');
      }
    })();

    return () => {
      active = false;
      if (createdUrl) URL.revokeObjectURL(createdUrl);
    };
  }, [warga, jenisSurat, formData, previewSurat, setError]);

  // Close the expanded modal on Escape, and lock background scroll while open.
  useEffect(() => {
    if (!expanded) return;
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setExpanded(false); };
    document.addEventListener('keydown', onKey);
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', onKey);
      document.body.style.overflow = prevOverflow;
    };
  }, [expanded]);

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

  const zoomIn = () => setZoom((z) => Math.min(ZOOM_MAX, Math.round((z + ZOOM_STEP) * 100) / 100));
  const zoomOut = () => setZoom((z) => Math.max(ZOOM_MIN, Math.round((z - ZOOM_STEP) * 100) / 100));
  const zoomReset = () => setZoom(ZOOM_DEFAULT);

  // For the embedded PDF, drive Chrome's PDF viewer zoom via the URL fragment.
  // Inline preview keeps Chrome's toolbar; the expanded modal hides it
  // (#toolbar=0) so only our themed controls remain.
  const pdfSrc = useMemo(() => {
    if (!pdfUrl) return '';
    const pct = Math.round(zoom * 100);
    const toolbar = expanded ? 0 : 1;
    return `${pdfUrl}#toolbar=${toolbar}&navpanes=0&zoom=${pct}`;
  }, [pdfUrl, zoom, expanded]);

  if (!warga || !jenisSurat) return null;

  const canPrint = (mode === 'pdf' && !!pdfUrl) || (mode === 'html' && !!previewHTML);
  const showZoom = mode === 'pdf' || mode === 'html';

  const zoomControls = (dark: boolean) => (
    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
      <button type="button" className="btn btn-secondary" onClick={zoomOut} disabled={zoom <= ZOOM_MIN} title="Perkecil" style={{ padding: '10px 14px' }}>
        <ZoomOut size={20} />
      </button>
      <button type="button" className="btn btn-secondary" onClick={zoomReset} title="Reset ukuran" style={{ padding: '10px 14px', minWidth: '78px', fontWeight: 700 }}>
        {Math.round(zoom * 100)}%
      </button>
      <button type="button" className="btn btn-secondary" onClick={zoomIn} disabled={zoom >= ZOOM_MAX} title="Perbesar" style={{ padding: '10px 14px' }}>
        <ZoomIn size={20} />
      </button>
      {dark ? (
        <button type="button" className="btn btn-primary" onClick={() => setExpanded(false)} title="Tutup" style={{ padding: '10px 16px' }}>
          <X size={20} /> Tutup
        </button>
      ) : (
        <button type="button" className="btn btn-primary" onClick={() => setExpanded(true)} title="Buka layar penuh" style={{ padding: '10px 14px' }}>
          <Maximize2 size={20} />
        </button>
      )}
    </div>
  );

  const renderPreviewSurface = (fullHeight: boolean) => {
    if (mode === 'pdf' && pdfUrl) {
      return <iframe key={pdfSrc} title="Pratinjau Surat" src={pdfSrc} style={{ flex: 1, width: '100%', height: fullHeight ? '100%' : undefined, border: 'none' }} />;
    }
    if (mode === 'html' && previewHTML) {
      return (
        <div style={{ flex: 1, overflow: 'auto', padding: '40px', display: 'flex', justifyContent: 'center' }}>
          <div
            style={{
              width: '100%',
              maxWidth: selectedFormat === 'F4' ? '720px' : '700px',
              color: '#000',
              background: '#fff',
              transform: `scale(${zoom})`,
              transformOrigin: 'top center',
              transition: 'transform 0.12s ease-out',
            }}
            dangerouslySetInnerHTML={{ __html: previewHTML }}
          />
        </div>
      );
    }
    if (mode === 'error') {
      return (
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--danger)' }}>
          Gagal memuat pratinjau. Periksa koneksi ke backend kiosk lalu coba lagi.
        </div>
      );
    }
    return (
      <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-dark)', gap: '10px' }}>
        <RefreshCw size={20} className="spinner" /> Merender dokumen...
      </div>
    );
  };

  return (
    <div className="page-container" style={{ height: '100%' }}>
      <div style={{ marginBottom: '20px', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', gap: '16px', flexWrap: 'wrap' }}>
        <div>
          <h2 style={{ fontSize: '24px', fontWeight: 800 }}>Pratinjau Surat</h2>
          <p style={{ color: 'var(--text-muted)', fontSize: '15px' }}>
            Periksa kembali isi surat sebelum dicetak. Tampilan ini sama persis dengan hasil cetak.
          </p>
        </div>
        {showZoom && zoomControls(false)}
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
        {renderPreviewSurface(false)}
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

      {/* Fullscreen themed modal */}
      {expanded && (
        <div
          role="dialog"
          aria-modal="true"
          onClick={(e) => { if (e.target === e.currentTarget) setExpanded(false); }}
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 1000,
            background: 'rgba(8, 12, 24, 0.82)',
            backdropFilter: 'blur(6px)',
            display: 'flex',
            flexDirection: 'column',
            padding: '24px',
            gap: '16px',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '16px', flexWrap: 'wrap' }}>
            <h2 style={{ fontSize: '22px', fontWeight: 800, color: 'var(--text-light, #fff)' }}>Pratinjau Surat — Layar Penuh</h2>
            {zoomControls(true)}
          </div>

          <div
            className="glass-card"
            style={{
              flex: 1,
              minHeight: 0,
              background: '#ffffff',
              borderRadius: 'var(--radius-md)',
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
              boxShadow: '0 20px 60px rgba(0,0,0,0.5)',
            }}
          >
            {renderPreviewSurface(true)}
          </div>
        </div>
      )}
    </div>
  );
};