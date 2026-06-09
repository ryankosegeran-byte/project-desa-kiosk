import React, { useState } from 'react';
import { VirtualKeyboard } from './VirtualKeyboard';
import { Search, AlertCircle } from 'lucide-react';

interface NIKInputProps {
  onSubmit: (nik: string) => void;
  loading: boolean;
}

export const NIKInput: React.FC<NIKInputProps> = ({ onSubmit, loading }) => {
  const [nik, setNik] = useState<string>('');
  const [validationError, setValidationError] = useState<string | null>(null);

  const handleKeyPress = (value: string) => {
    if (nik.length < 16) {
      setNik(prev => prev + value);
      setValidationError(null);
    }
  };

  const handleDelete = () => {
    setNik(prev => prev.slice(0, -1));
    setValidationError(null);
  };

  const handleClear = () => {
    setNik('');
    setValidationError(null);
  };

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (nik.length !== 16) {
      setValidationError('NIK harus terdiri dari tepat 16 digit angka.');
      return;
    }
    onSubmit(nik);
  };

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      gap: '24px',
      maxWidth: '480px',
      width: '100%',
      margin: '0 auto'
    }}>
      <form onSubmit={handleSearchSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        <div className="form-group" style={{ marginBottom: 0 }}>
          <label className="form-label" htmlFor="nik-input">Masukkan NIK Secara Manual</label>
          <div style={{ position: 'relative', display: 'flex', alignItems: 'center' }}>
            <input
              id="nik-input"
              type="text"
              readOnly
              value={nik}
              className="form-control"
              placeholder="Contoh: 3201234567890001"
              style={{
                fontSize: '24px',
                textAlign: 'center',
                letterSpacing: '2px',
                height: '60px',
                paddingRight: '60px'
              }}
            />
            {nik.length > 0 && (
              <span style={{
                position: 'absolute',
                right: '20px',
                color: 'var(--text-muted)',
                fontWeight: 600,
                fontSize: '14px'
              }}>
                {nik.length}/16
              </span>
            )}
          </div>
        </div>

        {validationError && (
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            color: 'var(--danger)',
            fontSize: '14px',
            fontWeight: 500
          }}>
            <AlertCircle size={16} />
            <span>{validationError}</span>
          </div>
        )}

        <button
          type="submit"
          disabled={loading || nik.length !== 16}
          className={`btn btn-primary ${(loading || nik.length !== 16) ? 'btn-disabled' : ''}`}
          style={{ width: '100%', height: '54px', fontSize: '18px' }}
        >
          <Search size={20} />
          {loading ? "Mencari..." : "Cari Data NIK"}
        </button>
      </form>

      <VirtualKeyboard
        onKeyPress={handleKeyPress}
        onDelete={handleDelete}
        onClear={handleClear}
      />
    </div>
  );
};
