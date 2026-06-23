import React, { useState, useEffect } from 'react';
import { History, RotateCcw, Eye, EyeOff, Clock, User } from 'lucide-react';
import { request } from '../lib/api';

interface TemplateVersion {
  id: string;
  template_id: string;
  version: number;
  template_html: string;
  format_kertas: string;
  change_note?: string;
  created_by?: string;
  created_at: string;
}

interface VersionHistoryProps {
  templateId: string;
  currentHTML: string;
  currentFormatKertas: string;
  onRollback: (html: string, formatKertas: string, version: number) => void;
  onClose: () => void;
}

export const VersionHistory: React.FC<VersionHistoryProps> = ({
  templateId,
  currentHTML,
  currentFormatKertas,
  onRollback,
  onClose,
}) => {
  const [versions, setVersions] = useState<TemplateVersion[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedVersion, setSelectedVersion] = useState<TemplateVersion | null>(null);
  const [showDiff, setShowDiff] = useState(false);

  useEffect(() => {
    loadVersions();
  }, [templateId]);

  const loadVersions = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await request(`/api/templates/history?template_id=${templateId}`);
      setVersions(data || []);
    } catch (err: any) {
      setError(err.message || 'Gagal memuat history versi');
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const handleRollback = () => {
    if (!selectedVersion) return;

    if (confirm(`Yakin ingin mengembalikan template ke versi ${selectedVersion.version}?`)) {
      onRollback(
        selectedVersion.template_html,
        selectedVersion.format_kertas,
        selectedVersion.version
      );
    }
  };

  // Simple diff view (text-based)
  const renderDiff = () => {
    if (!selectedVersion) return null;

    const oldLines = selectedVersion.template_html.split('\n');
    const newLines = currentHTML.split('\n');

    // Simple line-by-line comparison
    const diff: { type: 'same' | 'old' | 'new'; content: string }[] = [];

    const maxLines = Math.max(oldLines.length, newLines.length);
    for (let i = 0; i < maxLines; i++) {
      const oldLine = oldLines[i] || '';
      const newLine = newLines[i] || '';

      if (oldLine === newLine) {
        diff.push({ type: 'same', content: oldLine });
      } else {
        if (oldLine) diff.push({ type: 'old', content: oldLine });
        if (newLine) diff.push({ type: 'new', content: newLine });
      }
    }

    return (
      <div style={{
        maxHeight: '400px',
        overflow: 'auto',
        fontFamily: 'monospace',
        fontSize: '12px',
        background: 'rgba(0,0,0,0.3)',
        borderRadius: '8px',
        padding: '12px'
      }}>
        <div style={{ marginBottom: '12px', display: 'flex', justifyContent: 'space-between' }}>
          <span style={{ color: '#ef4444' }}>■ Versi {selectedVersion.version}</span>
          <span style={{ color: '#22c55e' }}>■ Versi Saat Ini</span>
        </div>
        <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
          {diff.map((line, i) => (
            <div
              key={i}
              style={{
                background: line.type === 'old' ? 'rgba(239,68,68,0.2)' :
                           line.type === 'new' ? 'rgba(34,197,94,0.2)' : 'transparent',
                padding: '2px 8px',
                borderLeft: line.type === 'old' ? '3px solid #ef4444' :
                           line.type === 'new' ? '3px solid #22c55e' : 'none',
              }}
            >
              <span style={{ color: 'var(--text-muted)', marginRight: '12px' }}>
                {i + 1}
              </span>
              {line.content || <span style={{ color: 'var(--text-muted)' }}>(empty line)</span>}
            </div>
          ))}
        </pre>
      </div>
    );
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
        maxWidth: '1000px',
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
              <History size={24} />
              Riwayat Versi Template
            </h2>
            <p style={{ color: 'var(--text-muted)', fontSize: '13px', marginTop: '4px' }}>
              {versions.length} versi tersimpan
            </p>
          </div>
          <button className="btn btn-secondary" onClick={onClose}>
            ✕ Tutup
          </button>
        </div>

        {/* Content */}
        <div style={{ flex: 1, overflow: 'hidden', display: 'flex' }}>
          {/* Version List */}
          <div style={{
            width: '300px',
            borderRight: '1px solid var(--border-color)',
            overflow: 'auto',
            padding: '16px'
          }}>
            {loading && (
              <div style={{ textAlign: 'center', padding: '40px', color: 'var(--text-muted)' }}>
                Memuat history...
              </div>
            )}

            {error && (
              <div style={{
                padding: '12px',
                background: 'rgba(239,68,68,0.1)',
                borderRadius: '8px',
                color: 'var(--danger)',
                fontSize: '13px'
              }}>
                {error}
              </div>
            )}

            {!loading && !error && versions.length === 0 && (
              <div style={{
                textAlign: 'center',
                padding: '40px',
                color: 'var(--text-muted)'
              }}>
                Belum ada history versi
              </div>
            )}

            {versions.map((version) => (
              <div
                key={version.id}
                onClick={() => setSelectedVersion(version)}
                style={{
                  padding: '12px',
                  marginBottom: '8px',
                  borderRadius: '8px',
                  cursor: 'pointer',
                  background: selectedVersion?.id === version.id ? 'rgba(59,130,246,0.2)' : 'rgba(0,0,0,0.2)',
                  border: selectedVersion?.id === version.id ? '1px solid var(--primary)' : '1px solid transparent',
                  transition: 'all 0.2s'
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '8px' }}>
                  <span style={{ fontWeight: 600, fontSize: '14px' }}>
                    Versi {version.version}
                  </span>
                  <span style={{
                    padding: '2px 8px',
                    background: version.format_kertas === 'F4' ? 'var(--secondary)' : 'var(--primary)',
                    borderRadius: '4px',
                    fontSize: '11px'
                  }}>
                    {version.format_kertas}
                  </span>
                </div>
                <div style={{ fontSize: '12px', color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: '4px' }}>
                  <Clock size={12} />
                  {formatDate(version.created_at)}
                </div>
                {version.change_note && (
                  <div style={{
                    marginTop: '8px',
                    fontSize: '11px',
                    color: 'var(--text-muted)',
                    fontStyle: 'italic'
                  }}>
                    "{version.change_note}"
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Preview Panel */}
          <div style={{ flex: 1, overflow: 'auto', padding: '20px' }}>
            {!selectedVersion && (
              <div style={{
                height: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--text-muted)'
              }}>
                Pilih versi untuk melihat preview
              </div>
            )}

            {selectedVersion && (
              <div>
                {/* Actions */}
                <div style={{ display: 'flex', gap: '12px', marginBottom: '16px' }}>
                  <button
                    className="btn btn-primary"
                    onClick={handleRollback}
                    style={{ display: 'flex', alignItems: 'center', gap: '8px' }}
                  >
                    <RotateCcw size={16} />
                    Kembalikan ke Versi {selectedVersion.version}
                  </button>
                  <button
                    className="btn btn-secondary"
                    onClick={() => setShowDiff(!showDiff)}
                    style={{ display: 'flex', alignItems: 'center', gap: '8px' }}
                  >
                    {showDiff ? <EyeOff size={16} /> : <Eye size={16} />}
                    {showDiff ? 'Sembunyikan' : 'Lihat'} Perbandingan
                  </button>
                </div>

                {/* Version Info */}
                <div style={{
                  padding: '12px',
                  background: 'rgba(0,0,0,0.2)',
                  borderRadius: '8px',
                  marginBottom: '16px',
                  fontSize: '13px'
                }}>
                  <div style={{ display: 'flex', gap: '24px' }}>
                    <div>
                      <span style={{ color: 'var(--text-muted)' }}>Versi:</span>
                      <strong style={{ marginLeft: '8px' }}>{selectedVersion.version}</strong>
                    </div>
                    <div>
                      <span style={{ color: 'var(--text-muted)' }}>Format:</span>
                      <strong style={{ marginLeft: '8px' }}>{selectedVersion.format_kertas}</strong>
                    </div>
                    <div>
                      <span style={{ color: 'var(--text-muted)' }}>Tanggal:</span>
                      <strong style={{ marginLeft: '8px' }}>{formatDate(selectedVersion.created_at)}</strong>
                    </div>
                  </div>
                  {selectedVersion.change_note && (
                    <div style={{ marginTop: '8px', color: 'var(--text-muted)' }}>
                      Catatan: {selectedVersion.change_note}
                    </div>
                  )}
                </div>

                {/* Diff View */}
                {showDiff && (
                  <div style={{ marginBottom: '20px' }}>
                    <h4 style={{ marginBottom: '12px' }}>Perbandingan dengan Versi Saat Ini</h4>
                    {renderDiff()}
                  </div>
                )}

                {/* HTML Preview */}
                <h4 style={{ marginBottom: '12px' }}>Preview Template</h4>
                <div style={{
                  maxHeight: '400px',
                  overflow: 'auto',
                  padding: '24px',
                  background: '#fff',
                  color: '#000',
                  borderRadius: '8px',
                  fontSize: '12px',
                  fontFamily: 'Times New Roman, serif',
                  lineHeight: 1.6
                }}>
                  <div dangerouslySetInnerHTML={{ __html: selectedVersion.template_html }} />
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default VersionHistory;
