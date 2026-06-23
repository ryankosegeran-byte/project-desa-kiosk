import React, { useState, useRef, useCallback } from 'react';
import { Upload, FileText, Check, X, ArrowRight, AlertCircle } from 'lucide-react';

// Variable mapping patterns - common Indonesian letter placeholders
const PLACEHOLDER_PATTERNS = [
  // Nama & Identitas
  { pattern: /Nama\s*(Lengkap)?/gi, variable: '{{.Warga.Nama}}', label: 'Nama Lengkap' },
  { pattern: /NIK|Nomor\s*KTP/gi, variable: '{{.Warga.NIK}}', label: 'NIK' },
  { pattern: /Tempat\s*(dan\s*)?(Tgl|Tanggal)\s*Lahir/gi, variable: '{{.Warga.TempatLahir}}, {{.Warga.TanggalLahir}}', label: 'Tempat/Tanggal Lahir' },
  { pattern: /Tempat\s*Lahir/gi, variable: '{{.Warga.TempatLahir}}', label: 'Tempat Lahir' },
  { pattern: /Tanggal\s*Lahir|Tgl\s*Lahir/gi, variable: '{{.Warga.TanggalLahir}}', label: 'Tanggal Lahir' },
  { pattern: /Jenis\s*Kelamin|JK/gi, variable: '{{.Warga.JenisKelamin}}', label: 'Jenis Kelamin' },
  { pattern: /Agama/gi, variable: '{{.Warga.Agama}}', label: 'Agama' },
  { pattern: /Status\s*(Kawin)?/gi, variable: '{{.Warga.StatusKawin}}', label: 'Status Kawin' },
  { pattern: /Pekerjaan/gi, variable: '{{.Warga.Pekerjaan}}', label: 'Pekerjaan' },

  // Alamat
  { pattern: /Alamat\s*(Lengkap)?/gi, variable: '{{.Warga.Alamat}}', label: 'Alamat' },
  { pattern: /RT/gi, variable: '{{.Warga.RT}}', label: 'RT' },
  { pattern: /RW/gi, variable: '{{.Warga.RW}}', label: 'RW' },
  { pattern: /Desa|Kelurahan/gi, variable: '{{.Warga.Kelurahan}}', label: 'Desa/Kelurahan' },
  { pattern: /Kecamatan/gi, variable: '{{.Warga.Kecamatan}}', label: 'Kecamatan' },
  { pattern: /Kabupaten|Kota/gi, variable: '{{.Warga.Kabupaten}}', label: 'Kabupaten' },

  // Desa & Surat
  { pattern: /Nama\s*Desa/gi, variable: '{{.DesaNama}}', label: 'Nama Desa' },
  { pattern: /Alamat\s*Desa|Kantor\s*Desa/gi, variable: '{{.DesaAlamat}}', label: 'Alamat Desa' },
  { pattern: /Kepala\s*Desa|H老祖?/gi, variable: '{{.DesaKepalaDesa}}', label: 'Kepala Desa' },
  { pattern: /NIP\s*(Kades)?/gi, variable: '{{.DesaNIP}}', label: 'NIP Kepala Desa' },
  { pattern: /Nomor\s*Surat/gi, variable: '{{.NomorSurat}}', label: 'Nomor Surat' },
  { pattern: /Tanggal\s*(Surat)?/gi, variable: '{{.DateToday}}', label: 'Tanggal Surat' },
];

interface DetectedPlaceholder {
  original: string;
  variable: string;
  label: string;
  selected: boolean;
}

interface DOCXImportWizardProps {
  onImport: (html: string, detectedPlaceholders: string[]) => void;
  onClose: () => void;
}

export const DOCXImportWizard: React.FC<DOCXImportWizardProps> = ({
  onImport,
  onClose,
}) => {
  const [step, setStep] = useState<'upload' | 'review' | 'mapping'>('upload');
  const [fileName, setFileName] = useState<string>('');
  const [rawHTML, setRawHTML] = useState<string>('');
  const [convertedHTML, setConvertedHTML] = useState<string>('');
  const [detectedPlaceholders, setDetectedPlaceholders] = useState<DetectedPlaceholder[]>([]);
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Handle file selection
  const handleFileSelect = useCallback(async (file: File) => {
    if (!file.name.endsWith('.docx')) {
      setError('Hanya file DOCX yang didukung');
      return;
    }

    setFileName(file.name);
    setLoading(true);
    setError('');

    try {
      // Dynamic import mammoth.js
      const mammoth = await import('mammoth');

      const arrayBuffer = await file.arrayBuffer();

      // Convert DOCX to HTML
      const result = await mammoth.convertToHtml(
        { arrayBuffer },
        {
          styleMap: [
            "p[style-name='Heading 1'] => h1:fresh",
            "p[style-name='Heading 2'] => h2:fresh",
            "p[style-name='Heading 3'] => h3:fresh",
            "p[style-name='Title'] => h1.document-title:fresh",
            "b => strong",
            "i => em",
            "u => u",
          ],
        }
      );

      // Handle warnings
      if (result.messages && result.messages.length > 0) {
        console.log('Mammoth warnings:', result.messages);
      }

      const html = result.value;
      setRawHTML(html);
      setConvertedHTML(html);

      // Detect placeholders
      const found: DetectedPlaceholder[] = [];
      const seenVariables = new Set<string>();

      for (const pattern of PLACEHOLDER_PATTERNS) {
        if (pattern.pattern.test(html)) {
          if (!seenVariables.has(pattern.variable)) {
            seenVariables.add(pattern.variable);
            found.push({
              original: html.match(pattern.pattern)?.[0] || pattern.label,
              variable: pattern.variable,
              label: pattern.label,
              selected: true,
            });
          }
        }
      }

      setDetectedPlaceholders(found);
      setStep('review');
    } catch (err: any) {
      console.error('DOCX conversion error:', err);
      setError('Gagal mengkonversi file DOCX: ' + (err.message || 'Unknown error'));
    } finally {
      setLoading(false);
    }
  }, []);

  // Handle file input change
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  // Toggle placeholder selection
  const togglePlaceholder = (index: number) => {
    setDetectedPlaceholders((prev) =>
      prev.map((p, i) =>
        i === index ? { ...p, selected: !p.selected } : p
      )
    );
  };

  // Apply placeholder replacements
  const applyMappings = () => {
    let html = rawHTML;

    // Replace detected text with variables
    for (const placeholder of detectedPlaceholders) {
      if (placeholder.selected) {
        // Create regex that matches the original text loosely
        const escapedOriginal = placeholder.original.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        html = html.replace(new RegExp(escapedOriginal, 'gi'), placeholder.variable);
      }
    }

    setConvertedHTML(html);
    setStep('mapping');
  };

  // Final import
  const handleImport = () => {
    const selectedVariables = detectedPlaceholders
      .filter((p) => p.selected)
      .map((p) => p.variable);

    onImport(convertedHTML, selectedVariables);
  };

  // Drag and drop handler
  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    const file = e.dataTransfer.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      background: 'var(--overlay)',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      zIndex: 1200,
      padding: '20px'
    }}>
      <div className="glass-card" style={{
        maxWidth: '900px',
        width: '100%',
        maxHeight: '90vh',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden'
      }}>
        {/* Header */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          padding: '20px',
          borderBottom: '1px solid var(--border-color)'
        }}>
          <div>
            <h2 style={{ fontSize: '20px', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '8px' }}>
              <FileText size={24} />
              Import Template dari DOCX
            </h2>
            <p style={{ color: 'var(--text-muted)', fontSize: '13px', marginTop: '4px' }}>
              Langkah {step === 'upload' ? '1' : step === 'review' ? '2' : '3'} dari 3
            </p>
          </div>
          <button className="btn btn-secondary" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        {/* Step Indicator */}
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          gap: '8px',
          padding: '16px',
          background: 'rgba(0,0,0,0.2)'
        }}>
          {['upload', 'review', 'mapping'].map((s, i) => (
            <div key={s} style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '50%',
                background: step === s ? 'var(--primary)' : i < ['upload', 'review', 'mapping'].indexOf(step) ? '#22c55e' : 'rgba(255,255,255,0.1)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontWeight: 600,
                fontSize: '14px'
              }}>
                {i + 1}
              </div>
              <span style={{ fontSize: '13px', color: step === s ? 'var(--text-main)' : 'var(--text-muted)' }}>
                {s === 'upload' ? 'Upload' : s === 'review' ? 'Review' : 'Mapping'}
              </span>
              {i < 2 && <ArrowRight size={16} style={{ color: 'var(--text-muted)' }} />}
            </div>
          ))}
        </div>

        {/* Content */}
        <div style={{ flex: 1, overflow: 'auto', padding: '24px' }}>
          {/* STEP 1: Upload */}
          {step === 'upload' && (
            <div>
              {error && (
                <div style={{
                  padding: '12px 16px',
                  background: 'rgba(239,68,68,0.1)',
                  border: '1px solid var(--danger)',
                  borderRadius: '8px',
                  color: 'var(--danger)',
                  marginBottom: '16px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px'
                }}>
                  <AlertCircle size={18} />
                  {error}
                </div>
              )}

              <div
                style={{
                  border: '2px dashed var(--border-color)',
                  borderRadius: '12px',
                  padding: '60px 40px',
                  textAlign: 'center',
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                  background: 'rgba(0,0,0,0.2)'
                }}
                onDrop={handleDrop}
                onDragOver={(e) => e.preventDefault()}
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload size={48} style={{ margin: '0 auto 16px', color: 'var(--primary)' }} />
                <h3 style={{ marginBottom: '8px' }}>
                  {loading ? 'Mengonversi...' : 'Drag & Drop file DOCX di sini'}
                </h3>
                <p style={{ color: 'var(--text-muted)', fontSize: '14px' }}>
                  atau klik untuk pilih file
                </p>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".docx"
                  style={{ display: 'none' }}
                  onChange={handleFileChange}
                />
              </div>

              <div style={{
                marginTop: '24px',
                padding: '16px',
                background: 'rgba(255,255,255,0.05)',
                borderRadius: '8px',
                fontSize: '13px',
                color: 'var(--text-muted)'
              }}>
                <strong style={{ color: 'var(--text-main)' }}>Tips:</strong>
                <ul style={{ marginTop: '8px', paddingLeft: '20px' }}>
                  <li>Pastikan file DOCX menggunakan font standar (Times New Roman, Arial)</li>
                  <li>Hapus header/footer dari dokumen Word sebelum import</li>
                  <li>Sistem akan mendeteksi placeholder seperti "Nama", "NIK", "Alamat", dll</li>
                </ul>
              </div>
            </div>
          )}

          {/* STEP 2: Review */}
          {step === 'review' && (
            <div>
              <div style={{
                padding: '16px',
                background: 'rgba(34,197,94,0.1)',
                border: '1px solid #22c55e',
                borderRadius: '8px',
                marginBottom: '20px'
              }}>
                <p style={{ display: 'flex', alignItems: 'center', gap: '8px', margin: 0 }}>
                  <Check size={18} color="#22c55e" />
                  File <strong>{fileName}</strong> berhasil dikonversi!
                </p>
              </div>

              {/* Detected Placeholders */}
              <h4 style={{ marginBottom: '12px' }}>Placeholder Terdeteksi ({detectedPlaceholders.length})</h4>
              <p style={{ color: 'var(--text-muted)', fontSize: '13px', marginBottom: '16px' }}>
                Pilih placeholder yang ingin diganti dengan variabel template:
              </p>

              <div style={{ display: 'grid', gap: '8px', maxHeight: '200px', overflow: 'auto' }}>
                {detectedPlaceholders.map((p, i) => (
                  <label
                    key={i}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '12px',
                      padding: '12px',
                      background: p.selected ? 'rgba(59,130,246,0.1)' : 'rgba(0,0,0,0.2)',
                      borderRadius: '8px',
                      cursor: 'pointer',
                      transition: 'all 0.2s'
                    }}
                  >
                    <input
                      type="checkbox"
                      checked={p.selected}
                      onChange={() => togglePlaceholder(i)}
                      style={{ width: '18px', height: '18px', accentColor: 'var(--primary)' }}
                    />
                    <div style={{ flex: 1 }}>
                      <div style={{ fontWeight: 600, fontSize: '14px' }}>{p.label}</div>
                      <code style={{ fontSize: '12px', color: 'var(--primary)' }}>{p.variable}</code>
                    </div>
                  </label>
                ))}
              </div>

              {detectedPlaceholders.length === 0 && (
                <div style={{
                  padding: '24px',
                  textAlign: 'center',
                  color: 'var(--text-muted)',
                  background: 'rgba(0,0,0,0.2)',
                  borderRadius: '8px'
                }}>
                  Tidak ada placeholder terdeteksi. Klik "Lanjut" untuk melihat hasil konversi.
                </div>
              )}

              {/* Preview HTML */}
              <details style={{ marginTop: '20px' }}>
                <summary style={{ cursor: 'pointer', color: 'var(--text-muted)', marginBottom: '8px' }}>
                  Preview HTML hasil konversi
                </summary>
                <div style={{
                  maxHeight: '200px',
                  overflow: 'auto',
                  padding: '16px',
                  background: 'rgba(0,0,0,0.3)',
                  borderRadius: '8px',
                  fontSize: '12px',
                  fontFamily: 'monospace'
                }}>
                  <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>
                    {convertedHTML.substring(0, 2000)}
                    {convertedHTML.length > 2000 && '...'}
                  </pre>
                </div>
              </details>
            </div>
          )}

          {/* STEP 3: Mapping */}
          {step === 'mapping' && (
            <div>
              <div style={{
                padding: '16px',
                background: 'rgba(59,130,246,0.1)',
                border: '1px solid var(--primary)',
                borderRadius: '8px',
                marginBottom: '20px'
              }}>
                <p style={{ display: 'flex', alignItems: 'center', gap: '8px', margin: 0 }}>
                  <Check size={18} color="var(--primary)" />
                  {detectedPlaceholders.filter(p => p.selected).length} placeholder berhasil dipetakan
                </p>
              </div>

              {/* Final Preview */}
              <h4 style={{ marginBottom: '12px' }}>Preview Template Final</h4>
              <div style={{
                maxHeight: '300px',
                overflow: 'auto',
                padding: '24px',
                background: '#fff',
                color: '#000',
                borderRadius: '8px',
                fontSize: '12px',
                fontFamily: 'Times New Roman, serif',
                lineHeight: 1.6
              }}>
                <div dangerouslySetInnerHTML={{ __html: convertedHTML }} />
              </div>

              <div style={{
                marginTop: '16px',
                padding: '12px',
                background: 'rgba(255,255,255,0.05)',
                borderRadius: '8px',
                fontSize: '13px'
              }}>
                <strong>Variabel yang akan digunakan:</strong>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', marginTop: '8px' }}>
                  {detectedPlaceholders
                    .filter(p => p.selected)
                    .map((p, i) => (
                      <span key={i} style={{
                        padding: '4px 8px',
                        background: 'var(--primary)',
                        borderRadius: '4px',
                        fontSize: '11px'
                      }}>
                        {p.variable}
                      </span>
                    ))}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          padding: '16px 24px',
          borderTop: '1px solid var(--border-color)'
        }}>
          <button
            className="btn btn-secondary"
            onClick={() => {
              if (step === 'review') setStep('upload');
              else if (step === 'mapping') setStep('review');
              else onClose();
            }}
          >
            {step === 'upload' ? 'Batal' : 'Kembali'}
          </button>

          {step === 'upload' && (
            <div />
          )}

          {step === 'review' && (
            <button className="btn btn-primary" onClick={applyMappings}>
              Lanjut ke Mapping →
            </button>
          )}

          {step === 'mapping' && (
            <button className="btn btn-primary" onClick={handleImport}>
              <Check size={18} />
              Import Template
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default DOCXImportWizard;
