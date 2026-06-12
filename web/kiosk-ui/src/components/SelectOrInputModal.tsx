import React, { useState } from 'react';
import { X, Check, Keyboard } from 'lucide-react';
import { FullKeyboard } from './FullKeyboard';

interface SelectOrInputModalProps {
  isOpen: boolean;
  title: string;
  options: string[];
  currentValue: string;
  onSelect: (value: string) => void;
  onClose: () => void;
}

export const SelectOrInputModal: React.FC<SelectOrInputModalProps> = ({
  isOpen,
  title,
  options,
  currentValue,
  onSelect,
  onClose,
}) => {
  const [mode, setMode] = useState<'select' | 'custom'>('select');
  const [customValue, setCustomValue] = useState('');

  if (!isOpen) return null;

  const handleOptionClick = (option: string) => {
    onSelect(option);
    onClose();
  };

  const handleCustomConfirm = () => {
    if (customValue.trim()) {
      onSelect(customValue.trim());
      onClose();
    }
  };

  const handleKeyPress = (key: string) => {
    setCustomValue(prev => prev + key);
  };

  const handleDelete = () => {
    setCustomValue(prev => prev.slice(0, -1));
  };

  const handleClear = () => {
    setCustomValue('');
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-container" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="modal-header">
          <h3>{title}</h3>
          <button className="modal-close-btn" onClick={onClose}>
            <X size={24} />
          </button>
        </div>

        {mode === 'select' ? (
          <div className="modal-body">
            {/* Option Cards Grid */}
            <div className="select-options-grid">
              {options.map((option, index) => (
                <button
                  key={index}
                  className={`select-option-card ${currentValue === option ? 'selected' : ''}`}
                  onClick={() => handleOptionClick(option)}
                >
                  {currentValue === option && <Check size={18} />}
                  <span>{option}</span>
                </button>
              ))}
            </div>

            {/* Lainnya Button */}
            <button
              className="btn btn-secondary"
              style={{ width: '100%', marginTop: '16px', borderStyle: 'dashed' }}
              onClick={() => setMode('custom')}
            >
              <Keyboard size={18} />
              Lainnya (Ketik Manual)
            </button>
          </div>
        ) : (
          <div className="modal-body">
            {/* Custom Input Preview */}
            <div className="custom-input-preview">
              <span className="custom-input-label">Ketik Jenis Usaha:</span>
              <div className="custom-input-display">
                {customValue || <span style={{ color: 'var(--text-muted)' }}>Ketik di sini...</span>}
                <span className="cursor-blink">|</span>
              </div>
            </div>

            {/* Full Keyboard */}
            <FullKeyboard
              onKeyPress={handleKeyPress}
              onDelete={handleDelete}
              onClear={handleClear}
              onEnter={handleCustomConfirm}
            />

            {/* Action Buttons */}
            <div style={{ display: 'flex', gap: '12px', marginTop: '16px' }}>
              <button
                className="btn btn-secondary"
                style={{ flex: 1 }}
                onClick={() => { setMode('select'); setCustomValue(''); }}
              >
                Kembali ke Pilihan
              </button>
              <button
                className="btn btn-primary"
                style={{ flex: 1 }}
                onClick={handleCustomConfirm}
                disabled={!customValue.trim()}
              >
                <Check size={18} />
                Simpan
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
