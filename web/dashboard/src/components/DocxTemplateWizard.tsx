import React, { useRef, useState } from "react";
import { Upload, FileText, X, Check, AlertCircle, Info } from "lucide-react";
import { request } from "../lib/api";

// Mirrors models.PlaceholderDef (Go).
interface PlaceholderDef {
  key: string;
  label: string;
  source: "warga" | "manual" | "sistem";
  warga_field?: string;
  sistem_field?: string;
  type?: string;
  options?: string[];
  required?: boolean;
  urutan?: number;
}

interface JenisSuratLite {
  id: string;
  kode: string;
  nama: string;
}

interface Props {
  jenisSurat: JenisSuratLite;
  desaId: string;
  onClose: () => void;
  onSaved?: () => void;
}

// Harus cocok dengan resolver di kiosk/print (wargaFieldValue / sistemFieldValue).
const WARGA_FIELDS = [
  { value: "Nama", label: "Nama Lengkap" },
  { value: "NIK", label: "NIK" },
  { value: "TempatLahir", label: "Tempat Lahir" },
  { value: "TanggalLahir", label: "Tanggal Lahir" },
  { value: "JenisKelamin", label: "Jenis Kelamin" },
  { value: "Alamat", label: "Alamat" },
  { value: "RT", label: "RT" },
  { value: "RW", label: "RW" },
  { value: "Kelurahan", label: "Desa/Kelurahan" },
  { value: "Kecamatan", label: "Kecamatan" },
  { value: "Kabupaten", label: "Kabupaten" },
  { value: "Provinsi", label: "Provinsi" },
  { value: "Agama", label: "Agama" },
  { value: "StatusKawin", label: "Status Kawin" },
  { value: "Pekerjaan", label: "Pekerjaan" },
  { value: "Kewarganegaraan", label: "Kewarganegaraan" },
];

const SISTEM_FIELDS = [
  { value: "NomorSurat", label: "Nomor Surat" },
  { value: "DateToday", label: "Tanggal Surat" },
  { value: "DesaKepalaDesa", label: "Kepala Desa" },
  { value: "DesaNIP", label: "NIP Kepala Desa" },
];

const MANUAL_TYPES = [
  { value: "text", label: "Teks singkat" },
  { value: "textarea", label: "Teks panjang" },
  { value: "number", label: "Angka" },
  { value: "date", label: "Tanggal" },
  { value: "select", label: "Pilihan (dropdown)" },
  { value: "address", label: "Lokasi (peta)" },
];

const SOURCES: { value: PlaceholderDef["source"]; label: string; desc: string }[] = [
  { value: "warga", label: "Dari KTP", desc: "Terisi otomatis dari data warga" },
  { value: "manual", label: "Diisi warga di kiosk", desc: "Jadi field di form kiosk" },
  { value: "sistem", label: "Otomatis sistem", desc: "Nomor surat / tanggal / kepala desa" },
];

export const DocxTemplateWizard: React.FC<Props> = ({ jenisSurat, desaId, onClose, onSaved }) => {
  const [step, setStep] = useState<"upload" | "mapping">("upload");
  const [fileName, setFileName] = useState("");
  const [uploading, setUploading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [warn, setWarn] = useState("");
  const [templateId, setTemplateId] = useState("");
  const [placeholders, setPlaceholders] = useState<PlaceholderDef[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Dua file dipilih di langkah 1, diunggah bersamaan.
  const [docxFile, setDocxFile] = useState<File | null>(null);
  const [pdfFile, setPdfFile] = useState<File | null>(null);
  const pdfInputRef = useRef<HTMLInputElement>(null);

  const pickDocx = (file: File) => {
    if (!file.name.toLowerCase().endsWith(".docx")) {
      setError("File master harus .docx");
      return;
    }
    setError("");
    setDocxFile(file);
    setFileName(file.name);
  };

  const pickPdf = (file: File) => {
    if (!file.name.toLowerCase().endsWith(".pdf")) {
      setError("Tampilan harus berupa file .pdf");
      return;
    }
    setError("");
    setPdfFile(file);
  };

  // Unggah docx (+ pdf opsional) sekaligus, lalu lanjut ke pemetaan.
  const handleUpload = async () => {
    if (!docxFile) {
      setError("Pilih file .docx dulu");
      return;
    }
    setError("");
    setWarn("");
    setUploading(true);
    try {
      const fd = new FormData();
      fd.append("docx", docxFile);
      if (pdfFile) fd.append("pdf", pdfFile);
      fd.append("jenis_surat_id", jenisSurat.id);
      if (desaId) fd.append("desa_id", desaId);

      const res = await request("/api/templates/upload-docx", { method: "POST", body: fd });
      setTemplateId(res.template.id);
      setPlaceholders(res.template.placeholders || []);
      if (!res.tokens || res.tokens.length === 0) {
        setWarn("Tidak ada penanda {{...}} terdeteksi. Tandai dulu di Word, mis. {{nama}}, {{jenis_usaha}}.");
      }
      setStep("mapping");
    } catch (e: any) {
      setError(e.message || "Gagal mengunggah file");
    } finally {
      setUploading(false);
    }
  };

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    const f = e.dataTransfer.files?.[0];
    if (f) pickDocx(f);
  };

  const patch = (i: number, p: Partial<PlaceholderDef>) =>
    setPlaceholders((prev) => prev.map((ph, idx) => (idx === i ? { ...ph, ...p } : ph)));

  const handleSave = async () => {
    setSaving(true);
    setError("");
    try {
      await request(`/api/templates/${templateId}/placeholders`, {
        method: "PUT",
        body: JSON.stringify(placeholders),
      });
      onSaved?.();
      onClose();
    } catch (e: any) {
      setError(e.message || "Gagal menyimpan pemetaan");
    } finally {
      setSaving(false);
    }
  };

  const manualFields = placeholders.filter((p) => p.source === "manual");

  return (
    <div style={overlay}>
      <div className="glass-card" style={{ maxWidth: "960px", width: "100%", maxHeight: "92vh", display: "flex", flexDirection: "column", overflow: "hidden", padding: 0 }}>
        {/* Header */}
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "20px 24px", borderBottom: "1px solid var(--border-color)" }}>
          <div>
            <h2 style={{ fontSize: "20px", fontWeight: 700, display: "flex", alignItems: "center", gap: "8px" }}>
              <FileText size={22} /> Atur Template DOCX — {jenisSurat.nama}
            </h2>
            <p style={{ color: "var(--text-muted)", fontSize: "13px", marginTop: "4px" }}>
              {step === "upload"
                ? "Langkah 1 dari 2 — Unggah file Word (.docx) yang sudah kamu beri penanda {{...}}."
                : "Langkah 2 dari 2 — Tentukan sumber data untuk tiap penanda."}
            </p>
          </div>
          <button className="btn btn-secondary" onClick={onClose}><X size={18} /></button>
        </div>

        {/* Body */}
        <div style={{ flex: 1, overflow: "auto", padding: "24px" }}>
          {error && <div style={banner("var(--danger)")}><AlertCircle size={18} /> {error}</div>}
          {warn && step === "mapping" && <div style={banner("#f59e0b")}><AlertCircle size={18} /> {warn}</div>}

          {step === "upload" ? (
            <>
              {/* Penjelasan singkat */}
              <div style={infoBox}>
                <Info size={18} style={{ flexShrink: 0, marginTop: 2 }} />
                <div style={{ lineHeight: 1.6 }}>
                  <strong>Cara kerja:</strong> di Word, ketik penanda <code style={code}>{"{{...}}"}</code> di tempat yang akan diisi
                  (mis. <code style={code}>{"Nama: {{nama}}"}</code>). Kop & layout dokumen <strong>dijaga 100%</strong> — kita hanya mengisi penanda, bukan mengubah tata letak.

                  <div style={{ marginTop: 10, display: "flex", flexDirection: "column", gap: 6 }}>
                    <div>
                      <span style={{ color: "var(--text-muted)" }}>Token data warga (otomatis dari KTP): </span>
                      <code style={code}>{"{{nama}}"}</code> <code style={code}>{"{{nik}}"}</code> <code style={code}>{"{{alamat}}"}</code> <code style={code}>{"{{tempat_lahir}}"}</code> <code style={code}>{"{{tanggal_lahir}}"}</code>
                    </div>
                    <div>
                      <span style={{ color: "var(--text-muted)" }}>Token sistem (otomatis): </span>
                      <code style={code}>{"{{tanggal}}"}</code> <code style={code}>{"{{kepala_desa}}"}</code>
                    </div>
                    <div>
                      <span style={{ color: "var(--text-muted)" }}>Token isian warga di kiosk: </span>
                      <code style={code}>{"{{jenis_usaha}}"}</code> <code style={code}>{"{{keperluan}}"}</code> <span style={{ color: "var(--text-muted)" }}>(atau token lain buatanmu)</span>
                    </div>
                  </div>

                  {/* Sorotan khusus nomor surat */}
                  <div style={{ marginTop: 12, padding: "10px 12px", background: "rgba(124,58,237,0.10)", border: "1px solid rgba(124,58,237,0.35)", borderRadius: 8 }}>
                    <strong style={{ color: "#c4b5fd" }}>Penomoran surat:</strong> tulis <code style={code}>{"{{nomor_surat}}"}</code> di lokasi nomor surat.
                    Token ini <strong>otomatis terdeteksi</strong> dan nilainya diisi dari pengaturan penomoran —
                    yang bisa kamu atur lewat tombol <strong>⚙️ Form &amp; Variabel → tab “Penomoran Surat”</strong> setelah template tersimpan.
                    <div style={{ marginTop: 4, color: "var(--text-muted)", fontSize: 12 }}>
                      Contoh penulisan di Word: <code style={code}>{"Nomor : {{nomor_surat}}"}</code> → tercetak <code style={code}>12/SK_USAHA/08.10/VI/2026</code>.
                    </div>
                  </div>
                </div>
              </div>

              {/* 1. DOCX master (wajib) */}
              <div style={dropzone} onDrop={onDrop} onDragOver={(e) => e.preventDefault()} onClick={() => fileInputRef.current?.click()}>
                <Upload size={40} style={{ margin: "0 auto 12px", color: docxFile ? "#22c55e" : "var(--primary)" }} />
                <h3 style={{ marginBottom: "6px" }}>
                  {docxFile ? `✓ ${fileName}` : "1. File Word (.docx) — master surat"}
                </h3>
                <p style={{ color: "var(--text-muted)", fontSize: "13px" }}>
                  {docxFile ? "Klik untuk ganti file" : "Wajib · tarik ke sini atau klik untuk pilih · pastikan sudah ada penanda {{...}}"}
                </p>
                <input ref={fileInputRef} type="file" accept=".docx" style={{ display: "none" }}
                  onChange={(e) => { const f = e.target.files?.[0]; if (f) pickDocx(f); }} />
              </div>

              {/* 2. PDF tampilan (opsional) */}
              <div style={pdfDrop} onClick={() => pdfInputRef.current?.click()}>
                <FileText size={22} style={{ color: pdfFile ? "#22c55e" : "#a78bfa", flexShrink: 0 }} />
                <div style={{ flex: 1 }}>
                  <strong style={{ fontSize: 13 }}>2. PDF tampilan — opsional</strong>
                  <p style={{ color: "var(--text-muted)", fontSize: 12, marginTop: 3, lineHeight: 1.5 }}>
                    Export dari Word (<strong>File → Save As → PDF</strong>). Hanya untuk pratinjau tampilan di dashboard — file inti yang dipakai mencetak tetap dokumen Word di atas.
                  </p>
                </div>
                <span className="btn btn-secondary" style={{ fontSize: 13, whiteSpace: "nowrap" }}>
                  {pdfFile ? `✓ ${pdfFile.name}` : "Pilih PDF"}
                </span>
                <input ref={pdfInputRef} type="file" accept=".pdf" style={{ display: "none" }}
                  onChange={(e) => { const f = e.target.files?.[0]; if (f) pickPdf(f); }} />
              </div>
            </>
          ) : (
            <>
              {/* Penjelasan mapping */}
              <div style={infoBox}>
                <Info size={18} style={{ flexShrink: 0, marginTop: 2 }} />
                <div>
                  Untuk tiap penanda, tentukan <strong>sumber datanya</strong>:
                  <div style={{ marginTop: 6, display: "flex", flexDirection: "column", gap: 3, color: "var(--text-muted)" }}>
                    <span>• <strong style={{ color: "var(--text-main)" }}>Dari KTP</strong> — terisi otomatis dari data warga (nama, NIK, alamat…).</span>
                    <span>• <strong style={{ color: "var(--text-main)" }}>Diisi warga</strong> — jadi <strong>form di kiosk</strong> (kamu atur label & jenis input di sini).</span>
                    <span>• <strong style={{ color: "var(--text-main)" }}>Otomatis sistem</strong> — nomor surat, tanggal, kepala desa.</span>
                  </div>
                </div>
              </div>

              <div style={{ display: "flex", alignItems: "center", gap: "8px", margin: "4px 0 16px", color: "var(--text-muted)", fontSize: "13px" }}>
                <Check size={16} color="#22c55e" /> <strong style={{ color: "var(--text-main)" }}>{fileName}</strong>
                {pdfFile && <> · <span style={{ color: "#a78bfa" }}>PDF tampilan ✓</span></>} — {placeholders.length} penanda terdeteksi
              </div>

              {placeholders.length === 0 && <p style={{ color: "var(--text-muted)" }}>Tidak ada penanda untuk dipetakan.</p>}

              <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
                {placeholders.map((p, i) => (
                  <div key={p.key + i} style={card}>
                    <code style={tokenCode}>{`{{${p.key}}}`}</code>

                    <div style={{ display: "grid", gridTemplateColumns: "150px 1fr", gap: "10px", alignItems: "center", marginTop: "12px" }}>
                      <label style={lbl}>Sumber data</label>
                      <select className="form-control" style={inp} value={p.source}
                        onChange={(e) => patch(i, { source: e.target.value as PlaceholderDef["source"] })}>
                        {SOURCES.map((s) => <option key={s.value} value={s.value}>{s.label} — {s.desc}</option>)}
                      </select>

                      {p.source === "warga" && (
                        <>
                          <label style={lbl}>Field KTP</label>
                          <select className="form-control" style={inp} value={p.warga_field || ""}
                            onChange={(e) => patch(i, { warga_field: e.target.value })}>
                            <option value="">— pilih field KTP —</option>
                            {WARGA_FIELDS.map((f) => <option key={f.value} value={f.value}>{f.label}</option>)}
                          </select>
                        </>
                      )}

                      {p.source === "sistem" && (
                        <>
                          <label style={lbl}>Field sistem</label>
                          <select className="form-control" style={inp} value={p.sistem_field || ""}
                            onChange={(e) => patch(i, { sistem_field: e.target.value })}>
                            <option value="">— pilih field sistem —</option>
                            {SISTEM_FIELDS.map((f) => <option key={f.value} value={f.value}>{f.label}</option>)}
                          </select>
                        </>
                      )}

                      {p.source === "manual" && (
                        <>
                          <label style={lbl}>Label di form</label>
                          <input className="form-control" style={inp} placeholder="mis. Jenis Usaha"
                            value={p.label} onChange={(e) => patch(i, { label: e.target.value })} />

                          <label style={lbl}>Jenis input</label>
                          <select className="form-control" style={inp} value={p.type || "text"}
                            onChange={(e) => patch(i, { type: e.target.value })}>
                            {MANUAL_TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
                          </select>

                          {p.type === "select" && (
                            <>
                              <label style={lbl}>Pilihan (pisah koma)</label>
                              <input className="form-control" style={inp} placeholder="mis. Permanen, Sementara, Pinjam"
                                value={(p.options || []).join(", ")}
                                onChange={(e) => patch(i, { options: e.target.value.split(",").map((s) => s.trim()).filter(Boolean) })} />
                            </>
                          )}

                          <label style={lbl}>Wajib diisi</label>
                          <label style={{ display: "flex", alignItems: "center", gap: "6px", fontSize: "13px", color: "var(--text-muted)" }}>
                            <input type="checkbox" checked={!!p.required} onChange={(e) => patch(i, { required: e.target.checked })} /> ya, wajib
                          </label>

                          {p.type === "address" && (
                            <div style={{ gridColumn: "1 / -1", fontSize: "12px", color: "var(--text-muted)", background: "rgba(59,130,246,0.08)", padding: "8px 10px", borderRadius: "6px" }}>
                              📍 Di kiosk, warga memilih titik di <strong>peta</strong> (atau ketik manual saat offline). Alamat hasilnya otomatis masuk ke surat.
                            </div>
                          )}
                        </>
                      )}
                    </div>
                  </div>
                ))}
              </div>

              {manualFields.length > 0 && (
                <div style={{ marginTop: "16px", padding: "12px 14px", background: "rgba(34,197,94,0.08)", border: "1px solid rgba(34,197,94,0.3)", borderRadius: "8px", fontSize: "13px" }}>
                  <strong>Akan muncul sebagai form di kiosk ({manualFields.length}):</strong>{" "}
                  {manualFields.map((p) => p.label || p.key).join(", ")}
                </div>
              )}
            </>
          )}
        </div>

        {/* Footer */}
        <div style={{ display: "flex", justifyContent: "space-between", padding: "16px 24px", borderTop: "1px solid var(--border-color)" }}>
          <button className="btn btn-secondary" onClick={step === "mapping" ? () => setStep("upload") : onClose}>
            {step === "mapping" ? "← Ganti file" : "Batal"}
          </button>
          {step === "upload" && (
            <button className="btn btn-primary" onClick={handleUpload} disabled={!docxFile || uploading}>
              {uploading ? "Mengunggah..." : "Lanjut →"}
            </button>
          )}
          {step === "mapping" && (
            <button className="btn btn-primary" onClick={handleSave} disabled={saving}>
              <Check size={18} /> {saving ? "Menyimpan..." : "Simpan Pemetaan"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

const overlay: React.CSSProperties = {
  position: "fixed", inset: 0, background: "var(--overlay)",
  display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1200, padding: "20px",
};
const dropzone: React.CSSProperties = {
  border: "2px dashed var(--border-color)", borderRadius: "12px", padding: "50px 40px",
  textAlign: "center", cursor: "pointer", background: "rgba(0,0,0,0.2)",
};
const card: React.CSSProperties = {
  padding: "14px 16px", background: "rgba(0,0,0,0.2)", borderRadius: "10px",
  border: "1px solid var(--border-color)",
};
const tokenCode: React.CSSProperties = { color: "var(--primary)", fontSize: "14px", fontWeight: 700 };
const lbl: React.CSSProperties = { fontSize: "13px", color: "var(--text-muted)" };
const inp: React.CSSProperties = { width: "100%", padding: "8px 10px", fontSize: "13px" };
const pdfDrop: React.CSSProperties = {
  display: "flex", gap: "12px", alignItems: "center", padding: "14px 16px", marginTop: "14px",
  background: "rgba(124,58,237,0.08)", border: "1px dashed rgba(124,58,237,0.4)", borderRadius: "12px", cursor: "pointer",
};
const infoBox: React.CSSProperties = {
  display: "flex", gap: "10px", padding: "14px 16px", marginBottom: "20px",
  background: "rgba(59,130,246,0.08)", border: "1px solid rgba(59,130,246,0.3)",
  borderRadius: "10px", fontSize: "13px", lineHeight: 1.5,
};
const code: React.CSSProperties = {
  background: "rgba(0,0,0,0.3)", padding: "1px 6px", borderRadius: "4px", fontSize: "12px", color: "var(--primary)",
};
function banner(color: string): React.CSSProperties {
  return {
    padding: "12px 16px", border: `1px solid ${color}`, color, borderRadius: "8px",
    marginBottom: "16px", display: "flex", alignItems: "center", gap: "8px", fontSize: "13px", background: "rgba(0,0,0,0.2)",
  };
}

export default DocxTemplateWizard;
