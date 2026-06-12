import React, { useState } from 'react';
import { Delete, CornerDownLeft, Space } from 'lucide-react';

interface FullKeyboardProps {
  onKeyPress: (key: string) => void;
  onDelete: () => void;
  onClear: () => void;
  onEnter?: () => void;
}

export const FullKeyboard: React.FC<FullKeyboardProps> = ({
  onKeyPress,
  onDelete,
  onClear,
  onEnter,
}) => {
  const [isShift, setIsShift] = useState(false);

  const rows = [
    ['1', '2', '3', '4', '5', '6', '7', '8', '9', '0'],
    ['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'],
    ['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L'],
    ['Z', 'X', 'C', 'V', 'B', 'N', 'M'],
  ];

  const symbolRow = ['.', ',', '/', '-', '(', ')', '"'];

  const handleKey = (key: string) => {
    const char = isShift ? key.toUpperCase() : key.toLowerCase();
    onKeyPress(char);
    if (isShift) setIsShift(false);
  };

  return (
    <div className="full-keyboard">
      {/* Number row */}
      <div className="full-kb-row">
        {rows[0].map((key) => (
          <button key={key} type="button" className="full-kb-key" onClick={() => onKeyPress(key)}>
            {key}
          </button>
        ))}
      </div>

      {/* QWERTY rows */}
      {rows.slice(1).map((row, rowIndex) => (
        <div key={rowIndex} className="full-kb-row">
          {rowIndex === 2 && (
            <button
              type="button"
              className={`full-kb-key full-kb-key-fn ${isShift ? 'active' : ''}`}
              onClick={() => setIsShift(!isShift)}
            >
              ⇧
            </button>
          )}
          {row.map((key) => (
            <button key={key} type="button" className="full-kb-key" onClick={() => handleKey(key)}>
              {isShift ? key.toUpperCase() : key.toLowerCase()}
            </button>
          ))}
          {rowIndex === 2 && (
            <button type="button" className="full-kb-key full-kb-key-fn" onClick={onDelete}>
              <Delete size={18} />
            </button>
          )}
        </div>
      ))}

      {/* Symbols row */}
      <div className="full-kb-row">
        {symbolRow.map((key) => (
          <button key={key} type="button" className="full-kb-key full-kb-key-sym" onClick={() => onKeyPress(key)}>
            {key}
          </button>
        ))}
      </div>

      {/* Bottom row: Clear + Space + Enter */}
      <div className="full-kb-row">
        <button type="button" className="full-kb-key full-kb-key-fn" onClick={onClear} style={{ flex: 1.5 }}>
          Hapus
        </button>
        <button type="button" className="full-kb-key" onClick={() => onKeyPress(' ')} style={{ flex: 5 }}>
          <Space size={18} />
        </button>
        {onEnter && (
          <button type="button" className="full-kb-key full-kb-key-enter" onClick={onEnter} style={{ flex: 2 }}>
            <CornerDownLeft size={18} /> OK
          </button>
        )}
      </div>
    </div>
  );
};
