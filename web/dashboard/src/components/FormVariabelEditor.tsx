import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  X, Download, Save, ChevronUp, ChevronDown, User, Settings2,
  PencilLine, Info, ListChecks, AlertCircle, CheckCircle2, FileText, Plus, Trash2,
} from "lucide-react";
import { API_BASE } from "../lib/api";

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

type Row = PlaceholderDef & { _id: string };

interface Props {
  templateId: string;
  placeholders: PlaceholderDef[];
  jenisSuratNama: string;
  onClose: () => void;
  onSaved?: (updated: PlaceholderDef[]) => void;
}

const WARGA_FIELDS = [
  { value: "Nama", label: "Nama Lengkap" },
  { value: "NIK", label: "NIK" },
  { value: "TempatLahir", label: "Tempat Lahir" },
  { value: "TanggalLahir", label: "Tanggal Lahir" },
  { value: "JenisKelamin", label: "Jenis Kelamin" },
  { value: "Alamat", label: "Alamat" },
  { value: "RT", label: "RT" },
  { value: "RW", label: "RW" },
  { value: "Kelurahan", label: "Kelurahan / Desa" },
  { value: "Kecamatan", label: "Kecamatan" },
  { value: "Kabupaten", label: "Kabupaten" },
  { value: "Provinsi", label: "Provinsi" },
  { value: "Agama", label: "Agama" },
  { value: "StatusKawin", label: "Status Perkawinan" },
  { value: "Pekerjaan", label: "Pekerjaan" },
  { value: "Kewarganegaraan", label: "Kewarganegaraan" },
];

const SISTEM_FIELDS = [
  { value: "NomorSurat", label: "Nomor Surat (otomatis)" },
  { value: "DateToday", label: "Tanggal Surat (hari ini)" },
  { value: "DesaKepalaDesa", label: "Nama Kepala Desa" },
  { value: "DesaNIP", label: "NIP Kepala Desa" },
];

const MANUAL_TYPES = [
  { value: "text", label: "Teks singkat" },
  { value: "textarea", label: "Teks panjang" },
  { value: "number", label: "Angka" },
  { value: "date", label: "Tanggal" },
  { value: "select", label: "Pilihan (dropdown)" },
];

const SOURCES = [
  { value: "warga", label: "Data KTP", icon: User, color: "#3b82f6", bg: "rgba(59,130,246,0.12)", hint: "Terisi otomatis dari data warga (hasil scan KTP)." },
  { value: "sistem", label: "Otomatis", icon: Settings2, color: "#22c55e", bg: "rgba(34,197,94,0.12)", hint: "Diisi otomatis oleh sistem (nomor surat, tanggal, kepala desa)." },
  { value: "manual", label: "Form Kiosk", icon: PencilLine, color: "#f59e0b", bg: "rgba(245,158,11,0.12)", hint: "Diketik warga lewat form di kiosk." },
] as const;

const sourceMeta = (s: string) => SOURCES.find((x) => x.value === s) || SOURCES[2];
let idSeq = 0;
const newId = () => `r${Date.now()}_${idSeq++}`;

// Nilai contoh default untuk fitur "Cek data → Word".
const WARGA_DEF: Record<string, string> = {
  Nama: "BUDI SANTOSO", NIK: "3201234567890001", TempatLahir: "Manado", TanggalLahir: "17 Juni 1990",
  JenisKelamin: "Laki-laki", Alamat: "Jl. Sudirman No. 10", RT: "001", RW: "002", Kelurahan: "Kalawat",
  Kecamatan: "Kalawat", Kabupaten: "Minahasa Utara", Provinsi: "Sulawesi Utara", Agama: "Kristen",
  StatusKawin: "Kawin", Pekerjaan: "Wiraswasta", Kewarganegaraan: "Indonesia",
};
const todayID = new Date().toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" });
const SISTEM_DEF: Record<string, string> = {
  NomorSurat: "001/SKU/08.10/VI/2026", DateToday: todayID,
  DesaKepalaDesa: "ALFRIDA, A.Md.Kes", DesaNIP: "196902061993032004",
};
const KEY_DEF: Record<string, string> = {
  jenis_usaha: "Toko Kelontong", merk_usaha: "Toko Budi Makmur", nama_usaha: "Toko Budi Makmur",
  alamat_usaha: "Jl. Desa Kalawat No. 5", tahun_kewajiban: "2026", tahun_mulai_usaha: "2020",
  penghasilan: "Rp 1.500.000", keterangan: "Untuk keperluan administrasi.", keperluan: "Administrasi",
};
const getDefaultDummy = (r: Row): string => {
  if (r.source === "warga") return WARGA_DEF[r.warga_field || ""] || `[${r.label || r.key}]`;
  if (r.source === "sistem") return SISTEM_DEF[r.sistem_field || ""] || `[${r.label || r.key}]`;
  if (r.type === "select" && r.options && r.options.length) return r.options[0];
  return KEY_DEF[r.key] || `[${r.label || r.key}]`;
};

export const FormVariabelEditor: React.FC<Props> = ({
  templateId, placeholders, jenisSuratNama, onClose, onSaved,
}) => {
  const [rows, setRows] = useState<Row[]>([]);
  const [saving, setSaving] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");
  const [dirty, setDirty] = useState(false);
  const [dummy, setDummy] = useState<Record<string, string>>({});
  const [pdfUrl, setPdfUrl] = useState<string | null>(null);
  const [pdfStatus, setPdfStatus] = useState<"loading" | "ready" | "none" | "error">("loading");
  const optionDraftRef = useRef<Record<string, string>>({});

  useEffect(() => {
    const sorted = [...placeholders]
      .sort((a, b) => (a.urutan ?? 0) - (b.urutan ?? 0))
      .map((p) => ({ ...p, _id: newId() }));
    setRows(sorted);
    const initDummy: Record<string, string> = {};
    for (const r of sorted) initDummy[r._id] = getDefaultDummy(r);
    setDummy(initDummy);
    setDirty(false);
  }, [placeholders]);

  // Ambil PDF draft (tampilan). 404 = belum ada.
  useEffect(() => {
    let url: string | null = null;
    let alive = true;
    (async () => {
      setPdfStatus("loading");
      try {
        const token = localStorage.getItem("token");
        const res = await fetch(`${API_BASE}/api/templates/${templateId}/preview-pdf`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
        });
        if (!alive) return;
        if (res.status === 404) { setPdfStatus("none"); return; }
        if (!res.ok) { setPdfStatus("error"); return; }
        const blob = await res.blob();
        if (!alive) return;
        url = URL.createObjectURL(blob);
        setPdfUrl(url);
        setPdfStatus("ready");
      } catch {
        if (alive) setPdfStatus("error");
      }
    })();
    return () => { alive = false; if (url) URL.revokeObjectURL(url); };
  }, [templateId]);

  const patch = (id: string, upd: Partial<Row>) => {
    setRows((prev) => prev.map((p) => (p._id === id ? { ...p, ...upd } : p)));
    setDirty(true);
  };

  const move = (idx: number, dir: -1 | 1) => {
    setRows((prev) => {
      const next = [...prev];
      const j = idx + dir;
      if (j < 0 || j >= next.length) return prev;
      [next[idx], next[j]] = [next[j], next[idx]];
      return next;
    });
    setDirty(true);
  };

  const addRow = () => {
    const r: Row = { _id: newId(), key: "", label: "", source: "manual", type: "text", required: true };
    setRows((prev) => [...prev, r]);
    setDummy((d) => ({ ...d, [r._id]: "" }));
    setDirty(true);
  };

  const removeRow = (id: string) => {
    setRows((prev) => prev.filter((p) => p._id !== id));
    setDirty(true);
  };

  const changeSource = (p: Row, source: Row["source"]) => {
    const upd: Partial<Row> = { source };
    if (source === "warga") upd.warga_field = p.warga_field || WARGA_FIELDS[0].value;
    else if (source === "sistem") upd.sistem_field = p.sistem_field || SISTEM_FIELDS[0].value;
    else { upd.type = p.type || "text"; if (p.required === undefined) upd.required = true; }
    patch(p._id, upd);
  };

  const manualCount = useMemo(() => rows.filter((p) => p.source === "manual").length, [rows]);

  const handleSave = async () => {
    // Validasi: key wajib & unik.
    const cleaned = rows.map((r) => ({ ...r, key: (r.key || "").trim() }));
    if (cleaned.some((r) => !r.key)) { setError("Ada variabel tanpa nama (key). Isi atau hapus dulu."); return; }
    const keys = cleaned.map((r) => r.key);
    const dup = keys.find((k, i) => keys.indexOf(k) !== i);
    if (dup) { setError(`Nama variabel duplikat: {{${dup}}}`); return; }

    setSaving(true); setError(""); setMsg("");
    try {
      const payload: PlaceholderDef[] = cleaned.map((r, i) => {
        const { _id, ...rest } = r;
        return { ...rest, urutan: i };
      });
      const token = localStorage.getItem("token");
      const res = await fetch(`${API_BASE}/api/templates/${templateId}/placeholders`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}) },
        body: JSON.stringify(payload),
      });
      if (!res.ok) {
        const d = await res.json().catch(() => ({}));
        throw new Error(d.error || "Gagal menyimpan");
      }
      setMsg("Tersimpan!");
      setDirty(false);
      setTimeout(() => setMsg(""), 2200);
      onSaved?.(payload);
    } catch (e: any) {
      setError(e.message || "Gagal menyimpan perubahan");
    } finally {
      setSaving(false);
    }
  };

  const handleDownload = async () => {
    setDownloading(true); setError("");
    try {
      // Kirim nilai contoh yang diisi admin (keyed by token key).
      const dummyValues: Record<string, string> = {};
      for (const r of rows) {
        const k = (r.key || "").trim();
        if (k) dummyValues[k] = dummy[r._id] ?? "";
      }
      const token = localStorage.getItem("token");
      const res = await fetch(`${API_BASE}/api/templates/${templateId}/preview`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}) },
        body: JSON.stringify({ dummy_values: dummyValues }),
      });
      if (!res.ok) {
        const d = await res.json().catch(() => ({}));
        throw new Error(d.error || "Gagal membuat contoh");
      }
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `contoh_${jenisSuratNama.replace(/\s+/g, "_")}.docx`;
      document.body.appendChild(a); a.click(); document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (e: any) {
      setError(e.message || "Gagal mengunduh contoh");
    } finally {
      setDownloading(false);
    }
  };

  return (
    <div style={S.overlay}>
      <style>{CSS}</style>
      <div className="glass-card" style={S.modal}>
        {/* Header */}
        <div style={S.header}>
          <div style={{ display: "flex", gap: 12, alignItems: "center" }}>
            <div style={S.headerIcon}><Settings2 size={20} /></div>
            <div>
              <h2 style={{ fontSize: 18, fontWeight: 700 }}>Form &amp; Variabel</h2>
              <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 2 }}>
                {jenisSuratNama} &nbsp;·&nbsp; {rows.length} variabel &nbsp;·&nbsp; {manualCount} field diisi warga
              </p>
            </div>
          </div>
          <button className="btn btn-secondary" onClick={onClose} style={{ padding: "6px 10px" }}><X size={18} /></button>
        </div>

        {/* Legenda */}
        <div style={S.legend}>
          <Info size={14} style={{ opacity: 0.6, flexShrink: 0 }} />
          <span style={{ fontSize: 12, color: "var(--text-muted)" }}>Sumber data tiap variabel:</span>
          {SOURCES.map((s) => (
            <span key={s.value} style={{ ...S.legendChip, color: s.color, background: s.bg }}>
              <s.icon size={12} /> {s.label}
            </span>
          ))}
        </div>

        {error && <div style={S.errorBanner}><AlertCircle size={15} /> {error}</div>}

        {/* Body: kiri mapping, kanan PDF */}
        <div style={S.body}>
          {/* KIRI — variable mapping */}
          <div style={S.leftCol}>
            <div style={{ display: "flex", alignItems: "center", marginBottom: 4 }}>
              <p style={{ ...S.colTitle, marginBottom: 0 }}><ListChecks size={13} /> VARIABEL</p>
              <button className="fve-add-btn" onClick={addRow} style={{ marginLeft: "auto" }}>
                <Plus size={14} /> Tambah Variabel
              </button>
            </div>

            {rows.length === 0 && (
              <div style={S.emptyState}>
                Belum ada variabel. Klik <b>Tambah Variabel</b>, atau unggah ulang DOCX dengan penanda <code>{"{{nama}}"}</code>.
              </div>
            )}

            {rows.map((p, idx) => {
              const meta = sourceMeta(p.source);
              return (
                <div key={p._id} className="fve-card" style={{ ...S.card, borderLeft: `3px solid ${meta.color}` }}>
                  {/* token editable + aksi */}
                  <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                    <span style={{ ...S.brace, color: meta.color }}>{"{{"}</span>
                    <input
                      className="fve-key-input"
                      style={{ color: meta.color }}
                      value={p.key}
                      placeholder="nama_variabel"
                      onChange={(e) => patch(p._id, { key: e.target.value.replace(/\s+/g, "_") })}
                    />
                    <span style={{ ...S.brace, color: meta.color }}>{"}}"}</span>
                    <div style={{ marginLeft: "auto", display: "flex", gap: 2 }}>
                      <button className="fve-icon-btn" title="Naik" disabled={idx === 0} onClick={() => move(idx, -1)}><ChevronUp size={15} /></button>
                      <button className="fve-icon-btn" title="Turun" disabled={idx === rows.length - 1} onClick={() => move(idx, 1)}><ChevronDown size={15} /></button>
                      <button className="fve-icon-btn fve-del" title="Hapus variabel" onClick={() => removeRow(p._id)}><Trash2 size={15} /></button>
                    </div>
                  </div>

                  {/* segmented sumber */}
                  <div style={S.segmented}>
                    {SOURCES.map((s) => {
                      const active = p.source === s.value;
                      return (
                        <button key={s.value} onClick={() => changeSource(p, s.value)} title={s.hint}
                          style={{ ...S.segBtn, ...(active ? { background: s.bg, color: s.color, borderColor: s.color } : {}) }}>
                          <s.icon size={13} /> {s.label}
                        </button>
                      );
                    })}
                  </div>

                  {/* konfigurasi */}
                  {p.source === "warga" && (
                    <Field label="Ambil dari data KTP">
                      <select className="fve-input" value={p.warga_field || ""} onChange={(e) => patch(p._id, { warga_field: e.target.value })}>
                        {WARGA_FIELDS.map((f) => <option key={f.value} value={f.value}>{f.label}</option>)}
                      </select>
                    </Field>
                  )}
                  {p.source === "sistem" && (
                    <Field label="Ambil dari sistem">
                      <select className="fve-input" value={p.sistem_field || ""} onChange={(e) => patch(p._id, { sistem_field: e.target.value })}>
                        {SISTEM_FIELDS.map((f) => <option key={f.value} value={f.value}>{f.label}</option>)}
                      </select>
                    </Field>
                  )}
                  {p.source === "manual" && (
                    <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                      <div style={{ display: "flex", gap: 10 }}>
                        <Field label="Label di form" grow>
                          <input className="fve-input" value={p.label} placeholder="mis. Jenis Usaha"
                            onChange={(e) => patch(p._id, { label: e.target.value })} />
                        </Field>
                        <Field label="Tipe input">
                          <select className="fve-input" value={p.type || "text"} onChange={(e) => patch(p._id, { type: e.target.value })}>
                            {MANUAL_TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
                          </select>
                        </Field>
                      </div>
                      {p.type === "select" && (
                        <Field label="Opsi pilihan (pisahkan dengan koma)">
                          <input className="fve-input" placeholder="mis. Permanen, Sementara, Pinjam"
                            defaultValue={(p.options || []).join(", ")}
                            onChange={(e) => { optionDraftRef.current[p._id] = e.target.value; }}
                            onBlur={(e) => patch(p._id, { options: e.target.value.split(",").map((s) => s.trim()).filter(Boolean) })} />
                        </Field>
                      )}
                      <label style={S.checkRow}>
                        <input type="checkbox" checked={!!p.required} onChange={(e) => patch(p._id, { required: e.target.checked })} />
                        <span>Wajib diisi warga</span>
                      </label>
                    </div>
                  )}

                  {/* Contoh nilai untuk fitur "Cek data → Word" */}
                  <div style={S.dummyRow}>
                    <span style={S.dummyLabel}>Contoh nilai</span>
                    <input className="fve-input" style={{ flex: 1 }} value={dummy[p._id] ?? ""}
                      placeholder="nilai contoh untuk Cek data"
                      onChange={(e) => setDummy((d) => ({ ...d, [p._id]: e.target.value }))} />
                  </div>
                </div>
              );
            })}
          </div>

          {/* KANAN — preview PDF draft */}
          <div style={S.rightCol}>
            <p style={S.colTitle}><FileText size={13} /> TAMPILAN SURAT (PDF DRAFT)</p>
            <div style={S.pdfWrap}>
              {pdfStatus === "loading" && <div style={S.pdfMsg}><div className="spinner" style={{ width: 18, height: 18 }} /><span>Memuat tampilan…</span></div>}
              {pdfStatus === "none" && (
                <div style={S.pdfMsg}>
                  <Info size={26} style={{ opacity: 0.4 }} />
                  <span style={{ textAlign: "center", lineHeight: 1.6 }}>
                    Belum ada PDF tampilan.<br />Unggah PDF (export dari Word) lewat wizard <b>📄 DOCX</b>.
                  </span>
                </div>
              )}
              {pdfStatus === "error" && <div style={S.pdfMsg}><AlertCircle size={26} style={{ opacity: 0.5 }} /><span>Gagal memuat tampilan PDF.</span></div>}
              {pdfStatus === "ready" && pdfUrl && (
                <iframe src={pdfUrl} title="Tampilan Surat" style={{ width: "100%", height: "100%", border: "none", borderRadius: 10, background: "#fff" }} />
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div style={S.footer}>
          <div style={{ display: "flex", alignItems: "center", gap: 10, fontSize: 13 }}>
            {msg && <span style={{ color: "#22c55e", display: "flex", gap: 4, alignItems: "center" }}><CheckCircle2 size={15} /> {msg}</span>}
            {dirty && !msg && <span style={{ color: "var(--text-muted)" }}>Ada perubahan belum disimpan</span>}
          </div>
          <div style={{ display: "flex", gap: 10 }}>
            <button className="btn btn-secondary" onClick={handleDownload} disabled={downloading || rows.length === 0}
              title="Unduh contoh DOCX terisi untuk cek data di Word">
              <Download size={15} /> {downloading ? "Menyiapkan…" : "Cek data → Word"}
            </button>
            <button className="btn btn-primary" onClick={handleSave} disabled={saving || !dirty}>
              <Save size={15} /> {saving ? "Menyimpan…" : "Simpan"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

const Field: React.FC<{ label: string; grow?: boolean; children: React.ReactNode }> = ({ label, grow, children }) => (
  <div style={{ display: "flex", flexDirection: "column", gap: 4, flex: grow ? 1 : undefined }}>
    <span style={S.fieldLabel}>{label}</span>
    {children}
  </div>
);

export default FormVariabelEditor;

const CSS = `
@keyframes fveSlideIn { from { opacity:0; transform: translateY(16px) scale(0.98);} to { opacity:1; transform: translateY(0) scale(1);} }
.fve-card { animation: fveSlideIn 0.18s ease; }
.fve-card:hover { background: rgba(255,255,255,0.025); }
.fve-input { background: rgba(0,0,0,0.28); border: 1px solid var(--border-color); border-radius: 7px; color: inherit; font-size: 13px; padding: 7px 10px; width: 100%; outline: none; transition: border-color .15s, box-shadow .15s; }
.fve-input:focus { border-color: var(--primary); box-shadow: 0 0 0 3px hsla(220,100%,60%,0.15); }
.fve-key-input { background: rgba(0,0,0,0.28); border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; font-weight: 700; padding: 5px 8px; outline: none; width: 100%; font-family: monospace; }
.fve-key-input:focus { border-color: var(--primary); box-shadow: 0 0 0 3px hsla(220,100%,60%,0.15); }
.fve-icon-btn { background: rgba(255,255,255,0.05); border: 1px solid var(--border-color); border-radius: 6px; color: var(--text-muted); cursor: pointer; padding: 3px; display: flex; transition: all .12s; }
.fve-icon-btn:hover:not(:disabled) { color: var(--primary); border-color: var(--primary); }
.fve-icon-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.fve-del:hover:not(:disabled) { color: #ef4444 !important; border-color: #ef4444 !important; }
.fve-add-btn { display: inline-flex; align-items: center; gap: 5px; font-size: 12px; font-weight: 600; padding: 5px 11px; border-radius: 7px; border: 1px solid var(--primary); background: hsla(220,100%,60%,0.12); color: var(--primary); cursor: pointer; transition: all .12s; }
.fve-add-btn:hover { background: hsla(220,100%,60%,0.22); }
`;

const S: Record<string, React.CSSProperties> = {
  overlay: { position: "fixed", inset: 0, background: "rgba(0,0,0,0.85)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1200, padding: 20 },
  modal: { maxWidth: 1180, width: "100%", height: "94vh", display: "flex", flexDirection: "column", overflow: "hidden", padding: 0, animation: "fveSlideIn 0.22s ease" },
  header: { display: "flex", justifyContent: "space-between", alignItems: "center", padding: "18px 24px", borderBottom: "1px solid var(--border-color)", flexShrink: 0 },
  headerIcon: { width: 40, height: 40, borderRadius: 10, display: "flex", alignItems: "center", justifyContent: "center", background: "linear-gradient(135deg, hsla(220,90%,55%,0.9), hsla(260,80%,55%,0.9))", color: "#fff", flexShrink: 0 },
  legend: { display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap", padding: "10px 24px", background: "rgba(0,0,0,0.18)", borderBottom: "1px solid var(--border-color)", flexShrink: 0 },
  legendChip: { display: "inline-flex", alignItems: "center", gap: 4, fontSize: 11, fontWeight: 700, padding: "3px 9px", borderRadius: 6 },
  errorBanner: { margin: "12px 24px 0", padding: "10px 14px", background: "rgba(239,68,68,0.12)", border: "1px solid rgba(239,68,68,0.3)", borderRadius: 8, color: "#ef4444", fontSize: 13, display: "flex", gap: 8, alignItems: "center", flexShrink: 0 },
  body: { display: "flex", flex: 1, minHeight: 0, overflow: "hidden" },
  leftCol: { width: "50%", borderRight: "1px solid var(--border-color)", padding: 20, overflowY: "auto", display: "flex", flexDirection: "column", gap: 12 },
  rightCol: { flex: 1, padding: 20, display: "flex", flexDirection: "column", minWidth: 0, background: "rgba(0,0,0,0.12)" },
  colTitle: { fontSize: 11, fontWeight: 700, letterSpacing: "0.08em", color: "var(--text-muted)", marginBottom: 10, textTransform: "uppercase", display: "flex", alignItems: "center", gap: 6 },
  emptyState: { color: "var(--text-muted)", fontSize: 13, padding: "20px", background: "rgba(0,0,0,0.2)", borderRadius: 8, lineHeight: 1.6 },
  card: { padding: "12px 14px", borderRadius: 10, background: "rgba(0,0,0,0.18)", display: "flex", flexDirection: "column", gap: 10 },
  brace: { fontSize: 14, fontWeight: 700, fontFamily: "monospace" },
  segmented: { display: "flex", gap: 6 },
  segBtn: { flex: 1, display: "flex", alignItems: "center", justifyContent: "center", gap: 5, fontSize: 12, fontWeight: 600, padding: "6px 4px", borderRadius: 7, border: "1px solid var(--border-color)", background: "rgba(255,255,255,0.03)", color: "var(--text-muted)", cursor: "pointer", transition: "all .12s" },
  fieldLabel: { fontSize: 11, color: "var(--text-muted)", fontWeight: 600 },
  checkRow: { display: "flex", alignItems: "center", gap: 8, fontSize: 13, color: "var(--text-muted)", cursor: "pointer", userSelect: "none" },
  dummyRow: { display: "flex", alignItems: "center", gap: 8, marginTop: 4, paddingTop: 10, borderTop: "1px dashed var(--border-color)" },
  dummyLabel: { fontSize: 11, color: "var(--text-muted)", fontWeight: 600, flexShrink: 0 },
  pdfWrap: { flex: 1, minHeight: 0, borderRadius: 10, border: "1px solid var(--border-color)", background: "rgba(0,0,0,0.25)", overflow: "hidden", display: "flex" },
  pdfMsg: { flex: 1, display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", gap: 12, color: "var(--text-muted)", fontSize: 13, padding: 24 },
  footer: { display: "flex", justifyContent: "space-between", alignItems: "center", padding: "14px 24px", borderTop: "1px solid var(--border-color)", flexShrink: 0 },
};
