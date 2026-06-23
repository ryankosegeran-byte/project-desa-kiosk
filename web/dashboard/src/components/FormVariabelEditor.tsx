import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  X, Download, Save, ChevronUp, ChevronDown, User, Settings2,
  PencilLine, Info, ListChecks, AlertCircle, CheckCircle2, FileText, Plus, Trash2, Hash,
  HelpCircle, Lightbulb, Sparkles, MousePointerClick,
} from "lucide-react";
import { request, authFetch } from "../lib/api";
import { NomorSuratPanel } from "./NomorSuratPanel";

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
  jenisSuratId?: string;
  jenisSuratKode?: string;
  desaId?: string;
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
  templateId, placeholders, jenisSuratNama, jenisSuratId, jenisSuratKode, desaId, onClose, onSaved,
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

  const [loadingVars, setLoadingVars] = useState(true);
  const [activeTab, setActiveTab] = useState<"variabel" | "nomor">("variabel");
  const [showHelp, setShowHelp] = useState(false);
  const [showReminder, setShowReminder] = useState(false);
  const helpBtnRef = useRef<HTMLButtonElement>(null);
  const reminderRef = useRef<HTMLDivElement>(null);
  const helpPanelRef = useRef<HTMLDivElement>(null);
  const reminderShownOnce = useRef(false);
  const helpSeen = useRef(false);

  const applyPlaceholders = (ph: PlaceholderDef[]) => {
    const sorted = [...ph]
      .sort((a, b) => (a.urutan ?? 0) - (b.urutan ?? 0))
      .map((p) => ({ ...p, _id: newId() }));
    setRows(sorted);
    const initDummy: Record<string, string> = {};
    for (const r of sorted) initDummy[r._id] = getDefaultDummy(r);
    setDummy(initDummy);
    setDirty(false);
  };

  useEffect(() => {
    let alive = true;
    setLoadingVars(true);
    (async () => {
      try {
        const full = await request(`/api/templates/${templateId}`);
        if (!alive) return;
        const ph: PlaceholderDef[] = full.placeholders ?? [];
        applyPlaceholders(ph.length > 0 ? ph : placeholders);
      } catch {
        if (!alive) return;
        applyPlaceholders(placeholders);
      } finally {
        if (alive) setLoadingVars(false);
      }
    })();
    return () => { alive = false; };
  }, [templateId]);

  // Ambil PDF draft (tampilan). 404 = belum ada.
  useEffect(() => {
    let url: string | null = null;
    let alive = true;
    (async () => {
      setPdfStatus("loading");
      try {
        const res = await authFetch(`/api/templates/${templateId}/preview-pdf`);
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

  // Pengingat bantuan: muncul ~3 detik setelah menu dibuka, lalu mengingatkan
  // ulang setiap 20 detik selama belum membuka bantuan. Klik mana saja menutupnya.
  useEffect(() => {
    if (showHelp) { helpSeen.current = true; setShowReminder(false); return; }
    if (helpSeen.current) return; // PIC sudah buka bantuan, berhenti mengingatkan.
    const t = window.setTimeout(() => {
      setShowReminder(true);
      reminderShownOnce.current = true;
    }, reminderShownOnce.current ? 20000 : 3000);
    return () => window.clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showHelp, showReminder]);

  // Animasi "pop" elastis ala anime.js (pakai Web Animations API, tanpa dependensi).
  useEffect(() => {
    if (!showReminder || !reminderRef.current) return;
    const el = reminderRef.current;
    el.animate(
      [
        { transform: "scale(0.6) translateY(8px)", opacity: 0 },
        { transform: "scale(1.08) translateY(0)", opacity: 1, offset: 0.7 },
        { transform: "scale(0.97)", offset: 0.85 },
        { transform: "scale(1)", opacity: 1 },
      ],
      { duration: 620, easing: "cubic-bezier(0.22, 1.4, 0.36, 1)", fill: "forwards" }
    );
    // Denyut halus berulang pada ikon untuk menarik perhatian.
    const icon = el.querySelector("[data-pulse]") as HTMLElement | null;
    let pulse: Animation | undefined;
    if (icon) {
      pulse = icon.animate(
        [{ transform: "scale(1)" }, { transform: "scale(1.25)" }, { transform: "scale(1)" }],
        { duration: 1100, iterations: Infinity, easing: "ease-in-out" }
      );
    }
    return () => pulse?.cancel();
  }, [showReminder]);

  // Animasi panel bantuan saat dibuka (slide + fade dengan easing elastis lembut).
  useEffect(() => {
    if (!showHelp || !helpPanelRef.current) return;
    helpPanelRef.current.animate(
      [
        { transform: "translateY(24px) scale(0.96)", opacity: 0 },
        { transform: "translateY(0) scale(1)", opacity: 1 },
      ],
      { duration: 420, easing: "cubic-bezier(0.16, 1, 0.3, 1)", fill: "forwards" }
    );
  }, [showHelp]);

  const openHelp = () => { setShowReminder(false); setShowHelp(true); };
  const dismissReminder = () => setShowReminder(false);

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
      await request(`/api/templates/${templateId}/placeholders`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });
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
      const res = await authFetch(`/api/templates/${templateId}/preview`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
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
          <div style={{ display: "flex", alignItems: "center", gap: 8, position: "relative" }}>
            {/* Tombol bantuan */}
            <button
              ref={helpBtnRef}
              className="fve-help-btn"
              onClick={openHelp}
              title="Bantuan: cara kerja menu ini"
            >
              <HelpCircle size={16} /> Bantuan
            </button>

            {/* Pop-up pengingat beranimasi (bisa di-skip dengan klik) */}
            {showReminder && !showHelp && (
              <div
                ref={reminderRef}
                role="button"
                onClick={openHelp}
                style={S.reminder}
                title="Klik untuk buka bantuan / tutup pengingat"
              >
                <span data-pulse style={S.reminderIcon}><Lightbulb size={15} /></span>
                <span style={{ fontSize: 12.5, lineHeight: 1.35 }}>
                  Bingung mengisi variabel? Klik <b>Bantuan</b> dulu — cuma 1 menit untuk paham cara kerjanya.
                </span>
                <button
                  onClick={(e) => { e.stopPropagation(); dismissReminder(); }}
                  style={S.reminderClose}
                  title="Tutup pengingat"
                  aria-label="Tutup pengingat"
                >
                  <X size={13} />
                </button>
                <span style={S.reminderArrow} />
              </div>
            )}

            <button className="btn btn-secondary" onClick={onClose} style={{ padding: "6px 10px" }}><X size={18} /></button>
          </div>
        </div>

        {/* Tab navigation */}
        <div style={{ display: "flex", borderBottom: "1px solid var(--border-color)", flexShrink: 0 }}>
          <button
            onClick={() => setActiveTab("variabel")}
            style={{ ...S.tabBtn, ...(activeTab === "variabel" ? S.tabBtnActive : {}) }}
          >
            <ListChecks size={14} /> Form & Variabel
          </button>
          {jenisSuratId && (
            <button
              onClick={() => setActiveTab("nomor")}
              style={{ ...S.tabBtn, ...(activeTab === "nomor" ? S.tabBtnActive : {}) }}
            >
              <Hash size={14} /> Penomoran Surat
            </button>
          )}
        </div>

        {/* Tab: Legenda (only variabel tab) */}
        {activeTab === "variabel" && (
        <div style={S.legend}>
          <Info size={14} style={{ opacity: 0.6, flexShrink: 0 }} />
          <span style={{ fontSize: 12, color: "var(--text-muted)" }}>Sumber data tiap variabel:</span>
          {SOURCES.map((s) => (
            <span key={s.value} style={{ ...S.legendChip, color: s.color, background: s.bg }}>
              <s.icon size={12} /> {s.label}
            </span>
          ))}
        </div>
        )}

        {activeTab === "variabel" && error && <div style={S.errorBanner}><AlertCircle size={15} /> {error}</div>}

        {/* Tab: Penomoran Surat */}
        {activeTab === "nomor" && jenisSuratId && (
          <div style={{ flex: 1, minHeight: 0, overflowY: "auto", padding: 24 }}>
            <NomorSuratPanel
              jenisSuratId={jenisSuratId}
              jenisSuratKode={jenisSuratKode || ""}
              jenisSuratNama={jenisSuratNama}
              desaId={desaId || ""}
            />
          </div>
        )}

        {/* Tab: Form & Variabel */}
        {activeTab === "variabel" && (
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
                {loadingVars
                  ? "Memuat variabel dari template..."
                  : <>Belum ada variabel. Klik <b>Tambah Variabel</b>, atau unggah ulang DOCX dengan penanda <code>{"{{nama}}"}</code>.</>}
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
        )}

        {/* Footer (hanya tab Form & Variabel; tab Nomor punya tombol Simpan sendiri) */}
        {activeTab === "variabel" && (
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
        )}

        {/* ===== Panel Bantuan ===== */}
        {showHelp && (
          <div style={S.helpOverlay} onClick={() => setShowHelp(false)}>
            <div ref={helpPanelRef} style={S.helpPanel} onClick={(e) => e.stopPropagation()}>
              <div style={S.helpHeader}>
                <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
                  <span style={S.helpHeaderIcon}><Sparkles size={18} /></span>
                  <div>
                    <h3 style={{ fontSize: 16, fontWeight: 700 }}>Cara kerja menu Form &amp; Variabel</h3>
                    <p style={{ fontSize: 12.5, color: "var(--text-muted)", marginTop: 2 }}>
                      Panduan singkat supaya data warga muncul dengan benar di surat yang dicetak.
                    </p>
                  </div>
                </div>
                <button className="btn btn-secondary" style={{ padding: "6px 10px" }} onClick={() => setShowHelp(false)}><X size={18} /></button>
              </div>

              <div style={S.helpBody}>
                <p style={{ fontSize: 13.5, lineHeight: 1.7, marginBottom: 4 }}>
                  <b>Variabel</b> adalah penanda <code style={S.helpCode}>{"{{...}}"}</code> yang kamu tulis di dokumen Word —
                  misalnya <code style={S.helpCode}>{"{{nama}}"}</code> di tempat nama warga. Di halaman ini kamu cukup menentukan
                  <b> dari mana</b> isi tiap penanda diambil. Saat warga mencetak di kiosk, penanda otomatis berganti jadi data asli.
                </p>

                <div style={S.helpSourceGrid}>
                  <div style={{ ...S.helpSourceCard, borderColor: "#3b82f6" }}>
                    <span style={{ ...S.helpSourceBadge, color: "#3b82f6", background: "rgba(59,130,246,0.12)" }}><User size={13} /> Data KTP</span>
                    <p style={S.helpSourceText}>Terisi sendiri dari hasil scan KTP — nama, NIK, tempat/tanggal lahir, alamat, dan lainnya. Tidak perlu diketik ulang.</p>
                  </div>
                  <div style={{ ...S.helpSourceCard, borderColor: "#22c55e" }}>
                    <span style={{ ...S.helpSourceBadge, color: "#22c55e", background: "rgba(34,197,94,0.12)" }}><Settings2 size={13} /> Otomatis</span>
                    <p style={S.helpSourceText}>Diisi sistem secara otomatis: nomor surat, tanggal hari ini, serta nama dan NIP kepala desa.</p>
                  </div>
                  <div style={{ ...S.helpSourceCard, borderColor: "#f59e0b" }}>
                    <span style={{ ...S.helpSourceBadge, color: "#f59e0b", background: "rgba(245,158,11,0.12)" }}><PencilLine size={13} /> Form Kiosk</span>
                    <p style={S.helpSourceText}>Diketik warga lewat form di layar kiosk — untuk data yang tidak ada di KTP, seperti jenis usaha atau keperluan surat.</p>
                  </div>
                </div>

                <div style={S.helpSteps}>
                  <h4 style={S.helpStepTitle}><ListChecks size={15} /> Langkah singkat</h4>
                  <ol style={S.helpOl}>
                    <li>Cek tiap variabel — pilih sumbernya: <b>Data KTP</b>, <b>Otomatis</b>, atau <b>Form Kiosk</b>.</li>
                    <li>Khusus <b>Form Kiosk</b>: isi <b>Label</b> dan pilih tipe isian. Label inilah yang dibaca warga saat mengisi di kiosk.</li>
                    <li>Isi kolom <b>Contoh nilai</b>, lalu klik <b>“Cek data → Word”</b> untuk mengunduh surat contoh dan memastikan tata letaknya sudah pas.</li>
                    <li>Klik <b>Simpan</b>. Pengaturan langsung dipakai saat surat dicetak dari kiosk.</li>
                  </ol>
                </div>

                <div style={S.helpNomor}>
                  <h4 style={{ ...S.helpStepTitle, color: "#c4b5fd" }}><Hash size={15} /> Mengatur nomor surat</h4>
                  <p style={{ fontSize: 13, lineHeight: 1.7 }}>
                    Cukup tulis <code style={S.helpCode}>{"{{nomor_surat}}"}</code> di tempat nomor pada dokumen Word.
                    Penanda ini <b>langsung dikenali</b> dan otomatis diatur ke sumber <b>Otomatis → Nomor Surat</b>, jadi kamu
                    tidak perlu mengetik nomor satu per satu. Untuk mengubah format, nomor awal, dan batas maksimalnya, buka
                    tab <b>“Penomoran Surat”</b> di bagian atas. Contoh hasil cetak: <code style={S.helpCode}>12/SK_USAHA/08.10/VI/2026</code>.
                  </p>
                </div>

                <div style={S.helpTip}>
                  <MousePointerClick size={15} style={{ flexShrink: 0, marginTop: 1 }} />
                  <span><b>Aman:</b> kop, tanda tangan, dan tata letak surat dari Word <b>tidak diubah sama sekali</b>. Sistem hanya menukar penanda <code style={S.helpCode}>{"{{...}}"}</code> dengan data asli warga.</span>
                </div>
              </div>

              <div style={S.helpFooter}>
                <span style={{ fontSize: 12, color: "var(--text-muted)" }}>Butuh lihat lagi? Klik tombol <b>Bantuan</b> di pojok kanan atas kapan saja.</span>
                <button className="btn btn-primary" onClick={() => setShowHelp(false)}>Mengerti</button>
              </div>
            </div>
          </div>
        )}
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
.fve-help-btn { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; font-weight: 600; padding: 6px 12px; border-radius: 8px; border: 1px solid rgba(124,58,237,0.5); background: rgba(124,58,237,0.14); color: #c4b5fd; cursor: pointer; transition: all .14s; }
.fve-help-btn:hover { background: rgba(124,58,237,0.26); border-color: rgba(124,58,237,0.8); }
`;

const S: Record<string, React.CSSProperties> = {
  overlay: { position: "fixed", inset: 0, background: "var(--overlay)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1200, padding: 20 },
  modal: { position: "relative", maxWidth: 1180, width: "100%", height: "94vh", display: "flex", flexDirection: "column", overflow: "hidden", padding: 0, animation: "fveSlideIn 0.22s ease" },
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
  tabBtn: { display: "flex", alignItems: "center", gap: 6, padding: "10px 20px", fontSize: 13, fontWeight: 600, background: "transparent", border: "none", color: "var(--text-muted)", cursor: "pointer", borderBottom: "2px solid transparent", transition: "all .15s", flexShrink: 0 },
  tabBtnActive: { color: "var(--primary)", borderBottom: "2px solid var(--primary)", background: "rgba(0,0,0,0.15)" },
  // Reminder bubble
  reminder: { position: "absolute", top: "calc(100% + 12px)", right: 0, width: 260, display: "flex", alignItems: "flex-start", gap: 8, padding: "10px 12px", background: "linear-gradient(135deg, rgba(124,58,237,0.97), rgba(79,70,229,0.97))", color: "#fff", borderRadius: 12, boxShadow: "0 10px 30px rgba(79,70,229,0.45)", cursor: "pointer", zIndex: 30, border: "1px solid rgba(255,255,255,0.15)" },
  reminderIcon: { display: "inline-flex", flexShrink: 0, width: 24, height: 24, borderRadius: 7, alignItems: "center", justifyContent: "center", background: "rgba(255,255,255,0.18)", color: "#fde68a", transformOrigin: "center" },
  reminderClose: { position: "absolute", top: 6, right: 6, width: 18, height: 18, display: "flex", alignItems: "center", justifyContent: "center", borderRadius: 5, border: "none", background: "rgba(255,255,255,0.18)", color: "#fff", cursor: "pointer", padding: 0 },
  reminderArrow: { position: "absolute", top: -6, right: 18, width: 12, height: 12, background: "rgba(124,58,237,0.97)", transform: "rotate(45deg)", borderLeft: "1px solid rgba(255,255,255,0.15)", borderTop: "1px solid rgba(255,255,255,0.15)" },
  // Help panel
  helpOverlay: { position: "absolute", inset: 0, background: "var(--overlay)", backdropFilter: "blur(2px)", display: "flex", alignItems: "center", justifyContent: "center", zIndex: 50, padding: 24 },
  helpPanel: { width: "100%", maxWidth: 640, maxHeight: "90%", display: "flex", flexDirection: "column", background: "var(--bg-surface)", border: "1px solid var(--border-color)", borderRadius: 16, overflow: "hidden", boxShadow: "0 24px 70px var(--overlay)" },
  helpHeader: { display: "flex", justifyContent: "space-between", alignItems: "center", padding: "16px 20px", borderBottom: "1px solid var(--border-color)", flexShrink: 0 },
  helpHeaderIcon: { width: 36, height: 36, borderRadius: 10, display: "flex", alignItems: "center", justifyContent: "center", background: "linear-gradient(135deg, hsla(280,80%,55%,0.95), hsla(220,90%,55%,0.95))", color: "#fff", flexShrink: 0 },
  helpBody: { padding: "18px 20px", overflowY: "auto", display: "flex", flexDirection: "column", gap: 16 },
  helpCode: { background: "rgba(0,0,0,0.35)", padding: "1px 6px", borderRadius: 4, fontSize: 12, color: "#93c5fd", fontFamily: "monospace" },
  helpSourceGrid: { display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 10 },
  helpSourceCard: { padding: "10px 12px", borderRadius: 10, border: "1px solid var(--border-color)", borderLeftWidth: 3, background: "rgba(0,0,0,0.18)", display: "flex", flexDirection: "column", gap: 6 },
  helpSourceBadge: { display: "inline-flex", alignItems: "center", gap: 5, fontSize: 12, fontWeight: 700, padding: "3px 8px", borderRadius: 6, alignSelf: "flex-start" },
  helpSourceText: { fontSize: 12, color: "var(--text-muted)", lineHeight: 1.5 },
  helpSteps: { padding: "12px 14px", borderRadius: 10, background: "rgba(0,0,0,0.18)", border: "1px solid var(--border-color)" },
  helpStepTitle: { fontSize: 13.5, fontWeight: 700, display: "flex", alignItems: "center", gap: 6, marginBottom: 8 },
  helpOl: { margin: 0, paddingLeft: 20, display: "flex", flexDirection: "column", gap: 6, fontSize: 13, lineHeight: 1.6, color: "var(--text-muted)" },
  helpNomor: { padding: "12px 14px", borderRadius: 10, background: "rgba(124,58,237,0.10)", border: "1px solid rgba(124,58,237,0.35)" },
  helpTip: { display: "flex", gap: 8, alignItems: "flex-start", padding: "10px 12px", borderRadius: 10, background: "rgba(59,130,246,0.08)", border: "1px solid rgba(59,130,246,0.25)", fontSize: 12.5, lineHeight: 1.55, color: "var(--text-muted)" },
  helpFooter: { display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12, padding: "14px 20px", borderTop: "1px solid var(--border-color)", flexShrink: 0 },
};