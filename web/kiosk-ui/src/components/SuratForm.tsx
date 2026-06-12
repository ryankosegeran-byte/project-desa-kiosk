import React, { useState, useEffect } from 'react';
import type { FieldDef } from '../hooks/useSurat';
import { Plus, Trash2, Calendar, FileText } from 'lucide-react';
import { SelectOrInputModal } from './SelectOrInputModal';
import { AddressPickerModal } from './AddressPickerModal';

interface SuratFormProps {
  fields: FieldDef[];
  onSubmit: (data: any) => void;
  onBack: () => void;
  warga: {
    nama: string;
    nik: string;
    alamat?: string;
    tempat_lahir?: string;
    tanggal_lahir?: string;
  };
}

export const SuratForm: React.FC<SuratFormProps> = ({
  fields,
  onSubmit,
  onBack,
  warga,
}) => {
  const [formData, setFormData] = useState<any>({});
  const [selectOrInputModal, setSelectOrInputModal] = useState<{ isOpen: boolean; fieldKey: string; title: string; options: string[] }>({
    isOpen: false, fieldKey: '', title: '', options: []
  });
  const [addressModal, setAddressModal] = useState<{ isOpen: boolean; fieldKey: string }>({
    isOpen: false, fieldKey: ''
  });

  // Initialize form state
  useEffect(() => {
    const initialData: any = {};
    fields.forEach(f => {
      if (f.type === 'repeater') {
        initialData[f.key] = [{}]; // start with one empty item in the list
      } else if (f.type === 'checkbox') {
        initialData[f.key] = false;
      } else {
        initialData[f.key] = '';
        // Auto-fill tahun kewajiban dengan tahun berjalan
        if (f.key === 'tahun_kewajiban') {
          initialData[f.key] = new Date().getFullYear().toString();
        }
      }
    });
    setFormData(initialData);
  }, [fields]);

  const handleFieldChange = (key: string, value: any) => {
    setFormData((prev: any) => ({
      ...prev,
      [key]: value
    }));
  };

  const handleRepeaterChange = (repeaterKey: string, index: number, fieldKey: string, value: any) => {
    setFormData((prev: any) => {
      const list = [...(prev[repeaterKey] || [])];
      list[index] = { ...list[index], [fieldKey]: value };
      return { ...prev, [repeaterKey]: list };
    });
  };

  const addRepeaterItem = (repeaterKey: string) => {
    setFormData((prev: any) => ({
      ...prev,
      [repeaterKey]: [...(prev[repeaterKey] || []), {}]
    }));
  };

  const removeRepeaterItem = (repeaterKey: string, index: number) => {
    setFormData((prev: any) => {
      const list = [...(prev[repeaterKey] || [])];
      if (list.length > 1) {
        list.splice(index, 1);
      }
      return { ...prev, [repeaterKey]: list };
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  // Helper to render individual form field
  const renderField = (field: FieldDef, value: any, onChange: (val: any) => void) => {
    switch (field.type) {
      case 'text':
        return (
          <input
            type="text"
            className="form-control"
            placeholder={field.placeholder || `Masukkan ${field.label.toLowerCase()}`}
            required={field.required}
            value={value || ''}
            onChange={(e) => onChange(e.target.value)}
          />
        );
      case 'textarea':
        return (
          <textarea
            className="form-control"
            placeholder={field.placeholder || `Masukkan ${field.label.toLowerCase()}`}
            required={field.required}
            rows={4}
            value={value || ''}
            onChange={(e) => onChange(e.target.value)}
            style={{ resize: 'none' }}
          />
        );
      case 'number':
        return (
          <input
            type="number"
            className="form-control"
            placeholder={field.placeholder || `Masukkan angka`}
            required={field.required}
            value={value || ''}
            onChange={(e) => onChange(e.target.value)}
          />
        );
      case 'date':
        return (
          <div style={{ position: 'relative', display: 'flex', alignItems: 'center' }}>
            <input
              type="date"
              className="form-control"
              required={field.required}
              value={value || ''}
              onChange={(e) => onChange(e.target.value)}
              style={{ paddingRight: '46px' }}
            />
            <Calendar size={20} style={{ position: 'absolute', right: '14px', color: 'var(--text-muted)', pointerEvents: 'none' }} />
          </div>
        );
      case 'select':
        return (
          <select
            className="form-control"
            required={field.required}
            value={value || ''}
            onChange={(e) => onChange(e.target.value)}
          >
            <option value="">-- Pilih {field.label} --</option>
            {field.options?.map((opt, i) => (
              <option key={i} value={opt}>{opt}</option>
            ))}
          </select>
        );
      case 'radio':
        return (
          <div style={{ display: 'flex', gap: '24px', padding: '10px 0' }}>
            {field.options?.map((opt, i) => (
              <label key={i} style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', fontSize: '16px' }}>
                <input
                  type="radio"
                  name={field.key}
                  required={field.required}
                  value={opt}
                  checked={value === opt}
                  onChange={() => onChange(opt)}
                  style={{ width: '18px', height: '18px', accentColor: 'var(--primary)' }}
                />
                <span>{opt}</span>
              </label>
            ))}
          </div>
        );
      case 'checkbox':
        return (
          <label style={{ display: 'flex', alignItems: 'center', gap: '12px', cursor: 'pointer', padding: '10px 0', fontSize: '16px' }}>
            <input
              type="checkbox"
              checked={!!value}
              onChange={(e) => onChange(e.target.checked)}
              style={{ width: '22px', height: '22px', accentColor: 'var(--primary)' }}
            />
            <span style={{ color: 'var(--text-muted)' }}>{field.placeholder || "Ya, setuju."}</span>
          </label>
        );
      case 'select_or_input':
        return (
          <button
            type="button"
            className="form-control"
            style={{
              textAlign: 'left',
              cursor: 'pointer',
              color: value ? 'var(--text-main)' : 'var(--text-muted)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
            onClick={() => setSelectOrInputModal({
              isOpen: true,
              fieldKey: field.key,
              title: field.label,
              options: field.options || [],
            })}
          >
            <span>{value || field.placeholder || `Pilih ${field.label}`}</span>
            <span style={{ fontSize: '12px', opacity: 0.6 }}>▼</span>
          </button>
        );
      case 'address':
        return (
          <button
            type="button"
            className="form-control"
            style={{
              textAlign: 'left',
              cursor: 'pointer',
              color: value ? 'var(--text-main)' : 'var(--text-muted)',
              minHeight: '52px',
            }}
            onClick={() => setAddressModal({ isOpen: true, fieldKey: field.key })}
          >
            {value || field.placeholder || 'Klik untuk pilih alamat'}
          </button>
        );
      default:
        return null;
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ flex: 1, overflowY: 'auto', paddingRight: '8px', marginBottom: '24px' }}>
        
        {/* Section: Data Pemohon (Auto-filled) */}
        <div className="glass-card" style={{ padding: '20px', marginBottom: '24px', borderLeft: '4px solid var(--primary)' }}>
          <h3 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--primary)', textTransform: 'uppercase', marginBottom: '12px' }}>
            Data Pemohon (Otomatis Terisi)
          </h3>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <div>
              <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>NAMA LENGKAP</span>
              <p style={{ fontSize: '18px', fontWeight: 600 }}>{warga.nama}</p>
            </div>
            <div>
              <span style={{ fontSize: '12px', color: 'var(--text-muted)' }}>NIK</span>
              <p style={{ fontSize: '18px', fontWeight: 600, letterSpacing: '1px' }}>{warga.nik}</p>
            </div>
          </div>
        </div>

        {/* Section: Dynamic Fields */}
        <h3 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--secondary)', textTransform: 'uppercase', marginBottom: '16px' }}>
          Informasi Tambahan Surat
        </h3>
        
        {fields.map((field) => {
          if (field.type === 'repeater') {
            const list = formData[field.key] || [];
            return (
              <div key={field.key} style={{ marginBottom: '24px', display: 'flex', flexDirection: 'column', gap: '12px' }}>
                <span className="form-label">{field.label}</span>
                
                {list.map((item: any, index: number) => (
                  <div key={index} className="glass-card" style={{ padding: '20px', position: 'relative' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                      <span style={{ fontSize: '14px', fontWeight: 600, color: 'var(--text-muted)' }}>
                        Data #{index + 1}
                      </span>
                      {list.length > 1 && (
                        <button
                          type="button"
                          className="btn btn-secondary"
                          onClick={() => removeRepeaterItem(field.key, index)}
                          style={{
                            padding: '6px 12px',
                            minHeight: 'auto',
                            background: 'hsla(355, 80%, 40%, 0.1)',
                            borderColor: 'hsla(355, 80%, 40%, 0.3)',
                            color: 'var(--danger)'
                          }}
                        >
                          <Trash2 size={16} />
                          Hapus
                        </button>
                      )}
                    </div>

                    <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                      {field.sub_fields?.map(sub => (
                        <div key={sub.key} className="form-group" style={{ marginBottom: 0 }}>
                          <label className="form-label" style={{ fontSize: '12px' }}>{sub.label}</label>
                          {renderField(sub, item[sub.key], (val) => handleRepeaterChange(field.key, index, sub.key, val))}
                        </div>
                      ))}
                    </div>
                  </div>
                ))}

                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => addRepeaterItem(field.key)}
                  style={{ alignSelf: 'flex-start', borderStyle: 'dashed' }}
                >
                  <Plus size={16} />
                  Tambah Baris
                </button>
              </div>
            );
          }

          return (
            <div key={field.key} className="form-group">
              <label className="form-label">
                {field.label} {field.required && <span style={{ color: 'var(--danger)' }}>*</span>}
              </label>
              {renderField(field, formData[field.key], (val) => handleFieldChange(field.key, val))}
            </div>
          );
        })}
      </div>

      {/* Form Actions Footer */}
      <div style={{
        display: 'flex',
        gap: '16px',
        paddingTop: '16px',
        borderTop: '1px solid var(--border-color)',
        marginTop: 'auto'
      }}>
        <button type="button" className="btn btn-secondary" onClick={onBack} style={{ flex: 1 }}>
          Kembali
        </button>
        <button type="submit" className="btn btn-primary" style={{ flex: 2 }}>
          <FileText size={20} />
          Lanjutkan ke Pratinjau
        </button>
      </div>

      {/* Modal: Select or Input */}
      <SelectOrInputModal
        isOpen={selectOrInputModal.isOpen}
        title={selectOrInputModal.title}
        options={selectOrInputModal.options}
        currentValue={formData[selectOrInputModal.fieldKey] || ''}
        onSelect={(val) => handleFieldChange(selectOrInputModal.fieldKey, val)}
        onClose={() => setSelectOrInputModal(prev => ({ ...prev, isOpen: false }))}
      />

      {/* Modal: Address Picker */}
      <AddressPickerModal
        isOpen={addressModal.isOpen}
        currentValue={formData[addressModal.fieldKey] || ''}
        onSelect={(val) => handleFieldChange(addressModal.fieldKey, val)}
        onClose={() => setAddressModal(prev => ({ ...prev, isOpen: false }))}
      />
    </form>
  );
};
