import React, { useEffect, useMemo, useState } from "react";
import { X, Hash, Save, Info, CheckCircle2, AlertCircle } from "lucide-react";
import { request } from "../lib/api";

interface JenisSurat {
  id: string;
  kode: string;
  nama: string;
}

interface Props {
  jenisSurat: JenisSurat[];
  desaId: string;
  desaNama?: string;
  onClose: () => void;
}

interface Cfg {
  nomor_mulai: number;
  batas_atas: number;
  nomor_terakhir: number;
  format_nomor: string;
  saving?: boolean;
  saved?: boolean;
  dirty?: boolean;
}

const DEFAULT_FORMAT = "{nomor}/{kode_surat}/{kode_desa}/{bulan_romawi}/{tahun}";

const TOKENS = [
  { t: "{nomor}", d: "Nomor urut (1, 2, 3…)" },
  { t: "{kode_surat}", d: "Kode jenis surat (mis. SKU)" },
  { t: "{kode_desa}", d: "Kode desa" },
  { t: "{bulan_romawi}", d: "Bulan angka romawi (VI)" },
  { t: "{bulan}", d: "Bulan 2 digit (06)" },
  { t: "{tahun}", d: "Tahun (2026)" },
];

const ROMAN = ["", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"];

// Mirror dari kiosk FormatNomorSurat untuk live preview.
function renderPreview(fmt: string, nomor: number, kodeSurat: string, kodeDesa: string): string {
  const now = new Date();
  return (fmt || "")
    .replaceAll("{nomor}", String(nomor))
    .replaceAll("{kode_surat}", kodeSurat || "—")
    .replaceAll("{kode_desa}", kodeDesa || "—")
    .replaceAll("{bulan_romawi}", ROMAN[now.getMonth() + 1])
    .replaceAll("{bulan}", String(now.getMonth() + 1).padStart(2, "0"))
    .replaceAll("{tahun}", String(now.getFullYear()));
}

export const NomorSuratConfig: React.FC<Props> = ({ jenisSurat, desaId, desaNama, onClose }) => {
  const [cfgs, setCfgs] = useState<Record<string, Cfg>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let alive = true;
    // Tampilkan default dulu agar baris selalu muncul, lalu timpa dengan data server.
    const base: Record<string, Cfg> = {};
    for (const js of jenisSurat) {
      base[js.id] = { nomor_mulai: 1, batas_atas: 1000, nomor_terakhir: 0, format_nomor: DEFAULT_FORMAT };
    }
    setCfgs(base);

    (async () => {
      setLoading(true); setError("");
      try {
        const q = desaId ? `?desa_id=${encodeURIComponent(desaId)}` : "";
        const existing: any[] = await request(`/api/nomor-surat${q}`);
        if (!alive) return;
        const merged = { ...base };
        for (const c of existing || []) {
          if (!merged[c.jenis_surat_id]) continue;
          merged[c.jenis_surat_id] = {
            nomor_mulai: c.nomor_mulai ?? 1,
            batas_atas: c.batas_atas ?? 1000,
            nomor_terakhir: c.nomor_terakhir ?? 0,
            format_nomor: c.format_nomor || DEFAULT_FORMAT,
          };
        }
        setCfgs(merged);
      } catch (e: any) {
        if (alive) setError(e.message || "Gagal memuat konfigurasi (memakai nilai default)");
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => { alive = false; };
  }, [desaId, jenisSurat]);

  const patch = (id: string, upd: Partial<Cfg>) =>
    setCfgs((prev) => ({ ...prev, [id]: { ...prev[id], ...upd, dirty: true, saved: false } }));

  const save = async (js: JenisSurat) => {
    const c = cfgs[js.id];
    if (!c) return;
    if (c.batas_atas < c.nomor_mulai) {
      patch(js.id, {});
      setError(`${js.nama}: batas atas harus ≥ nomor mulai`);
      return;
    }
    setError("");
    patch(js.id, { saving: true });
    try {
      await request(`/api/nomor-surat/${js.id}`, {
        method: "PUT",
        body: JSON.stringify({
          desa_id: desaId,
          nomor_mulai: c.nomor_mulai,
          batas_atas: c.batas_atas,
          format_nomor: c.format_nomor,
        }),
      });
      setCfgs((prev) => ({ ...prev, [js.id]: { ...prev[js.id], saving: false, saved: true, dirty: false } }));
      setTimeout(() => setCfgs((prev) => ({ ...prev, [js.id]: { ...prev[js.id], saved: false } })), 2000);
    } catch (e: any) {
      setCfgs((prev) => ({ ...prev, [js.id]: { ...prev[js.id], saving: false } }));
      setError(e.message || "Gagal menyimpan");
    }
  };

  const totalConfigured = useMemo(
    () => jenisSurat.filter((js) => cfgs[js.id] && (cfgs[js.id].nomor_terakhir > 0 || !cfgs[js.id].dirty)).length,
    [cfgs, jenisSurat]
  );

  return (
    <div style={S.overlay}>
      <style>{CSS}</style>
      <div className="glass-card" style={S.modal}>
        {/* Header */}
        <div style={S.header}>
          <div style={{ display: "flex", gap: 12, alignItems: "center" }}>
            <div style={S.headerIcon}><Hash size={20} /></div>
            <div>
              <h2 style={{ fontSize: 18, fontWeight: 700 }}>Pengaturan Nomor Surat</h2>
              <p style={{ color: "var(--text-muted)", fontSize: 13, marginTop: 2 }}>
                {desaNama ? `Desa ${desaNama} · ` : ""}{jenisSurat.length} jenis surat
              </p>
            </div>
          </div>
          <button className="btn btn-secondary" onClick={onClose} style={{ padding: "6px 10px" }}><X size={18} /></button>
        </div>

        {/* Token legend */}
        <div style={S.legend}>
          <Info size={14} style={{ opacity: 0.6, flexShrink: 0 }} />
          <span style={{ fontSize: 12, color: "var(--text-muted)" }}>Token format yang bisa dipakai:</span>
          {TOKENS.map((tk) => (
            <span key={tk.t} title={tk.d} style={S.tokenChip}><code>{tk.t}</code></span>
          ))}
        </div>

        {error && <div style={S.errorBanner}><AlertCircle size={15} /> {error}</div>}

        {/* Body */}
        <div style={S.body}>
          {loading ? (
            <div style={{ display: "flex", justifyContent: "center", padding: 60 }}><div className="spinner" /></div>
          ) : jenisSurat.length === 0 ? (
            <div style={S.empty}>Belum ada jenis surat. Tambahkan jenis surat dulu.</div>
          ) : (
            jenisSurat.map((js) => {
              const c = cfgs[js.id];
              if (!c) return null;
              const full = c.nomor_terakhir >= c.batas_atas && c.batas_atas > 0;
              const usagePct = c.batas_atas > 0 ? Math.min(100, Math.round((c.nomor_terakhir / c.batas_atas) * 100)) : 0;
              const previewNomor = Math.max(c.nomor_mulai, c.nomor_terakhir + 1);
              return (
                <div key={js.id} className="nsc-card" style={S.card}>
                  <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 12 }}>
                    <span className="badge badge-primary" style={{ fontSize: 11 }}>{js.kode}</span>
                    <h3 style={{ fontSize: 15, fontWeight: 700 }}>{js.nama}</h3>
                    <span style={{ marginLeft: "auto", fontSize: 12, color: full ? "#ef4444" : "var(--text-muted)" }}>
                      Terpakai {c.nomor_terakhir}/{c.batas_atas}{full ? " · PENUH" : ""}
                    </span>
                  </div>

                  {/* usage bar */}
                  <div style={S.usageTrack}>
                    <div style={{ ...S.usageFill, width: `${usagePct}%`, background: full ? "#ef4444" : usagePct > 80 ? "#f59e0b" : "#22c55e" }} />
                  </div>

                  <div style={S.grid}>
                    <Field label="Nomor mulai">
                      <input className="nsc-input" type="number" min={0} value={c.nomor_mulai}
                        onChange={(e) => patch(js.id, { nomor_mulai: parseInt(e.target.value) || 0 })} />
                    </Field>
                    <Field label="Batas atas">
                      <input className="nsc-input" type="number" min={0} value={c.batas_atas}
                        onChange={(e) => patch(js.id, { batas_atas: parseInt(e.target.value) || 0 })} />
                    </Field>
                    <Field label="Format nomor" grow>
                      <input className="nsc-input" value={c.format_nomor}
                        onChange={(e) => patch(js.id, { format_nomor: e.target.value })}
                        placeholder={DEFAULT_FORMAT} />
                    </Field>
                  </div>

                  <div style={{ display: "flex", alignItems: "center", gap: 12, marginTop: 12 }}>
                    <div style={{ flex: 1, fontSize: 13, color: "var(--text-muted)" }}>
                      Contoh hasil:{" "}
                      <code style={S.preview}>{renderPreview(c.format_nomor, previewNomor, js.kode, "08.10") || "—"}</code>
                    </div>
                    {c.saved && <span style={{ color: "#22c55e", display: "flex", gap: 4, alignItems: "center", fontSize: 13 }}><CheckCircle2 size={15} /> Tersimpan</span>}
                    <button className="btn btn-primary" style={{ padding: "7px 16px", fontSize: 13 }}
                      onClick={() => save(js)} disabled={c.saving || !c.dirty}>
                      <Save size={14} /> {c.saving ? "Menyimpan…" : "Simpan"}
                    </button>
                  </div>
                </div>
              );
            })
          )}
        </div>

        <div style={S.footer}>
          <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
            <Info size={13} style={{ verticalAlign: "-2px", marginRight: 4, opacity: 0.6 }} />
            Nomor & batas tersinkron ke kiosk. Saat batas penuh, kiosk berhenti menerbitkan sampai diperbarui.
          </span>
          <button className="btn btn-secondary" onClick={onClose}>Selesai</button>
        </div>
      </div>
    </div>
  );
};

const Field: React.FC<{ label: string; grow?: boolean; children: React.ReactNode }> = ({ label, grow, children }) => (
  <div style={{ display: "flex", flexDirection: "column", gap: 4, flex: grow ? 1 : "0 0 110px" }}>
    <span style={{ fontSize: 11, color: "var(--text-muted)", fontWeight: 600 }}>{label}</span>
    {children}
  </div>
);

export default NomorSuratConfig;

const CSS = `
@keyframes nscIn { from { opacity:0; transform: translateY(16px) scale(0.98);} to { opacity:1; transform: translateY(0) scale(1);} }
.nsc-card { animation: nscIn 0.18s ease; }
.nsc-input { background: rgba(0,0,0,0.28); border: 1px solid var(--border-color); border-radius: 7px; color: inherit; font-size: 13px; padding: 7px 10px; width: 100%; outline: none; transition: border-color .15s, box-shadow .15s; }
.nsc-input:focus { border-color: var(--primary); box-shadow: 0 0 0 3px hsla(220,100%,60%,0.15); }
`;

const S: Record<string, React.CSSProperties> = {
  overlay: { position: "fixed", inset: 0, background: "rgba(0,0,0,0.85)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1200, padding: 20 },
  modal: { maxWidth: 920, width: "100%", maxHeight: "92vh", display: "flex", flexDirection: "column", overflow: "hidden", padding: 0, animation: "nscIn 0.22s ease" },
  header: { display: "flex", justifyContent: "space-between", alignItems: "center", padding: "18px 24px", borderBottom: "1px solid var(--border-color)", flexShrink: 0 },
  headerIcon: { width: 40, height: 40, borderRadius: 10, display: "flex", alignItems: "center", justifyContent: "center", background: "linear-gradient(135deg, hsla(280,80%,55%,0.9), hsla(220,90%,55%,0.9))", color: "#fff", flexShrink: 0 },
  legend: { display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap", padding: "10px 24px", background: "rgba(0,0,0,0.18)", borderBottom: "1px solid var(--border-color)", flexShrink: 0 },
  tokenChip: { fontSize: 11, padding: "2px 7px", borderRadius: 5, background: "rgba(99,102,241,0.12)", color: "#a5b4fc", cursor: "help" },
  errorBanner: { margin: "12px 24px 0", padding: "10px 14px", background: "rgba(239,68,68,0.12)", border: "1px solid rgba(239,68,68,0.3)", borderRadius: 8, color: "#ef4444", fontSize: 13, display: "flex", gap: 8, alignItems: "center", flexShrink: 0 },
  body: { flex: 1, minHeight: 0, overflowY: "auto", padding: 20, display: "flex", flexDirection: "column", gap: 14 },
  empty: { color: "var(--text-muted)", fontSize: 13, padding: 40, textAlign: "center" },
  card: { padding: "14px 16px", borderRadius: 10, background: "rgba(0,0,0,0.18)", border: "1px solid var(--border-color)" },
  usageTrack: { height: 4, borderRadius: 4, background: "rgba(255,255,255,0.08)", overflow: "hidden", marginBottom: 12 },
  usageFill: { height: "100%", borderRadius: 4, transition: "width .3s" },
  grid: { display: "flex", gap: 12, flexWrap: "wrap", alignItems: "flex-end" },
  preview: { color: "var(--primary)", fontWeight: 700, background: "rgba(0,0,0,0.25)", padding: "2px 8px", borderRadius: 5 },
  footer: { display: "flex", justifyContent: "space-between", alignItems: "center", gap: 16, padding: "14px 24px", borderTop: "1px solid var(--border-color)", flexShrink: 0 },
};
