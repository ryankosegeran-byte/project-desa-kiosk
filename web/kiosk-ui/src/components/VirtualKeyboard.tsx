import React from 'react';

interface VirtualKeyboardProps {
  onKeyPress: (value: string) => void;
  onDelete: () => void;
  onClear: () => void;
}

export const VirtualKeyboard: React.FC<VirtualKeyboardProps> = ({
  onKeyPress,
  onDelete,
  onClear,
}) => {
  const keys = ['1', '2', '3', '4', '5', '6', '7', '8', '9', 'C', '0', '⌫'];

  const handleKeyClick = (key: string) => {
    if (key === '⌫') {
      onDelete();
    } else if (key === 'C') {
      onClear();
    } else {
      onKeyPress(key);
    }
  };

  return (
    <div className="keyboard-container glass-card">
      <div className="keyboard-grid">
        {keys.map((key, index) => {
          let className = "keyboard-key";
          if (key === '⌫' || key === 'C') {
            className += " wide";
          }
          return (
            <button
              key={index}
              type="button"
              className={className}
              onClick={() => handleKeyClick(key)}
            >
              {key}
            </button>
          );
        })}
      </div>
    </div>
  );
};
