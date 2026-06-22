import React, { useEffect, useState } from "react";
import { Save, Info, CheckCircle2, AlertCircle, RefreshCw } from "lucide-react";
import { request } from "../lib/api";

interface Props {
  jenisSuratId: string;
  jenisSuratKode: string;
  jenisSuratNama: string;
  desaId: string;
}

interface Cfg {
  nomor_mulai: number;
  batas_atas: number;
  nomor_terakhir: number;
  format_nomor: string;
}

const DEFAULT_FORMAT = "{nomor}/{kode_surat}/{kode_desa}/{bulan_romawi}/{tahun}";

const TOKENS = [
  { t: "{nomor}", d: "Nomor urut" },
  { t: "{kode_surat}", d: "Kode jenis surat" },
  { t: "{kode_desa}", d: "Kode desa" },
  { t: "{bulan_romawi}", d: "Bulan romawi (VI)" },
  { t: "{bulan}", d: "Bulan 2 digit (06)" },
  { t: "{tahun}", d: "Tahun (2026)" },
];

const ROMAN = ["", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"];

function renderPreview(fmt: string, nomor: number, kodeSurat: string, kodeDesa: string): string {
  const now = new Date();
  return (fmt || "")
    .replaceAll("{nomor}", String(nomor))
    .replaceAll("{kode_surat}", kodeSurat || "-")
    .replaceAll("{kode_desa}", kodeDesa || "-")
    .replaceAll("{bulan_romawi}", ROMAN[now.getMonth() + 1])
    .replaceAll("{bulan}", String(now.getMonth() + 1).padStart(2, "0"))
    .replaceAll("{tahun}", String(now.getFullYear()));
}

export const NomorSuratPanel: React.FC<Props> = ({
  jenisSuratId, jenisSuratKode, desaId,
}) => {
  const [cfg, setCfg] = useState<Cfg>({
    nomor_mulai: 1, batas_atas: 1000, nomor_terakhir: 0, format_nomor: DEFAULT_FORMAT,
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState("");

  const load = async () => {
    setLoading(true); setError("");
    try {
      const q = desaId ? `?desa_id=${encodeURIComponent(desaId)}` : "";
      const list: any[] = await request(`/api/nomor-surat${q}`);
      const found = (list || []).find((c: any) => c.jenis_surat_id === jenisSuratId);
      if (found) {
        setCfg({
          nomor_mulai: found.nomor_mulai ?? 1,
          batas_atas: found.batas_atas ?? 1000,
          nomor_terakhir: found.nomor_terakhir ?? 0,
          format_nomor: found.format_nomor || DEFAULT_FORMAT,
        });
      }
    } catch (e: any) {
      setError(e.message || "Gagal memuat konfigurasi nomor surat");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [jenisSuratId, desaId]);

  const patch = (upd: Partial<Cfg>) => {
    setCfg((prev) => ({ ...prev, ...upd }));
    setDirty(true);
    setSaved(false);
  };

  const handleSave = async () => {
    if (cfg.batas_atas < cfg.nomor_mulai) {
      setError("Batas atas harus lebih besar atau sama dengan nomor mulai");
      return;
    }
    setError(""); setSaving(true);
    try {
      await request(`/api/nomor-surat/${jenisSuratId}`, {
        method: "PUT",
        body: JSON.stringify({
          desa_id: desaId,
          nomor_mulai: cfg.nomor_mulai,
          batas_atas: cfg.batas_atas,
          format_nomor: cfg.format_nomor,
        }),
      });
      setDirty(false); setSaved(true);
      setTimeout(() => setSaved(false), 2500);
    } catch (e: any) {
      setError(e.message || "Gagal menyimpan");
    } finally {
      setSaving(false);
    }
  };

  const previewNomor = cfg.nomor_terakhir > 0 ? cfg.nomor_terakhir + 1 : cfg.nomor_mulai;
  const usedCount = Math.max(0, cfg.nomor_terakhir - cfg.nomor_mulai + 1);
  const totalSlot = cfg.batas_atas - cfg.nomor_mulai + 1;
  const usagePct = totalSlot > 0 ? Math.min(100, Math.round((usedCount / totalSlot) * 100)) : 0;
  const usageColor = usagePct >= 90 ? "#ef4444" : usagePct >= 70 ? "#f59e0b" : "#22c55e";

  if (loading) {
    return (
      <div style={{ display: "flex", alignItems: "center", gap: 10, padding: "32px 0", color: "var(--text-muted)", fontSize: 13 }}>
        <div className="spinner" style={{ width: 18, height: 18 }} />
        Memuat konfigurasi nomor surat...
      </div>
    );
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 18 }}>
      <style>{CSS}</style>

      {/* Token legend */}
      <div style={S.legendBox}>
        <Info size={13} style={{ opacity: 0.6, flexShrink: 0, marginTop: 1 }} />
        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          <span style={{ fontSize: 12, color: "var(--text-muted)", fontWeight: 600 }}>
            Token yang bisa dipakai di format nomor (klik untuk sisipkan):
          </span>
          <div style={{ display: "flex", flexWrap: "wrap", gap: 6 }}>
            {TOKENS.map((tk) => (
              <span
                key={tk.t}
                title={tk.d}
                style={S.tokenChip}
                onClick={() => patch({ format_nomor: cfg.format_nomor + tk.t })}
              >
                {tk.t}
              </span>
            ))}
          </div>
        </div>
      </div>

      {/* Usage bar */}
      <div style={S.usageCard}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
          <span style={{ fontSize: 13, fontWeight: 600 }}>Pemakaian nomor</span>
          <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
            Terpakai <b style={{ color: usageColor }}>{usedCount}</b> / {totalSlot}
          </span>
        </div>
        <div style={{ height: 6, borderRadius: 4, background: "rgba(255,255,255,0.08)", overflow: "hidden" }}>
          <div style={{ width: `${usagePct}%`, height: "100%", borderRadius: 4, background: usageColor, transition: "width .3s" }} />
        </div>
        {usagePct >= 90 && (
          <p style={{ fontSize: 12, color: "#ef4444", marginTop: 8 }}>
            Hampir penuh! Segera perbarui batas atas agar kiosk tetap bisa menerbitkan surat.
          </p>
        )}
      </div>

      {/* Config fields */}
      <div style={S.configCard}>
        <div style={{ display: "flex", gap: 12, flexWrap: "wrap", alignItems: "flex-end" }}>
          <div style={S.fieldBox}>
            <label style={S.fieldLabel}>Nomor mulai</label>
            <input
              className="nsp-input nsp-num"
              type="number" min={1}
              value={cfg.nomor_mulai}
              onChange={(e) => patch({ nomor_mulai: Math.max(1, Number(e.target.value)) })}
            />
          </div>
          <div style={S.fieldBox}>
            <label style={S.fieldLabel}>Batas atas</label>
            <input
              className="nsp-input nsp-num"
              type="number" min={1}
              value={cfg.batas_atas}
              onChange={(e) => patch({ batas_atas: Math.max(1, Number(e.target.value)) })}
            />
          </div>
          <div style={{ ...S.fieldBox, flex: 1, minWidth: 200 }}>
            <label style={S.fieldLabel}>Format nomor</label>
            <input
              className="nsp-input"
              value={cfg.format_nomor}
              placeholder={DEFAULT_FORMAT}
              onChange={(e) => patch({ format_nomor: e.target.value })}
            />
          </div>
        </div>

        {/* Live preview */}
        <div style={S.previewRow}>
          <span style={{ fontSize: 12, color: "var(--text-muted)" }}>Contoh hasil:</span>
          <code style={S.previewCode}>
            {renderPreview(cfg.format_nomor, previewNomor, jenisSuratKode, "08.10") || "-"}
          </code>
          <span style={{ fontSize: 11, color: "var(--text-muted)", marginLeft: "auto" }}>
            nomor ke-{previewNomor}, kode desa contoh: 08.10
          </span>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div style={S.errorBanner}>
          <AlertCircle size={15} /> {error}
        </div>
      )}

      {/* Actions */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <div style={{ display: "flex", gap: 8, alignItems: "center", fontSize: 13 }}>
          {saved && (
            <span style={{ color: "#22c55e", display: "flex", gap: 4, alignItems: "center" }}>
              <CheckCircle2 size={15} /> Tersimpan
            </span>
          )}
          {!saved && dirty && <span style={{ color: "var(--text-muted)" }}>Ada perubahan belum disimpan</span>}
          {!dirty && !saved && (
            <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
              Nomor & batas tersinkron ke kiosk. Saat batas penuh, kiosk berhenti menerbitkan.
            </span>
          )}
        </div>
        <div style={{ display: "flex", gap: 8 }}>
          <button
            className="btn btn-secondary"
            style={{ fontSize: 12, padding: "6px 10px" }}
            onClick={load}
            disabled={loading || saving}
            title="Muat ulang dari server"
          >
            <RefreshCw size={13} />
          </button>
          <button
            className="btn btn-primary"
            style={{ fontSize: 13, padding: "7px 18px" }}
            onClick={handleSave}
            disabled={saving || !dirty}
          >
            <Save size={14} /> {saving ? "Menyimpan..." : "Simpan"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default NomorSuratPanel;

const CSS = `
.nsp-input { background: rgba(0,0,0,0.28); border: 1px solid var(--border-color); border-radius: 7px; color: inherit; font-size: 13px; padding: 7px 10px; width: 100%; outline: none; transition: border-color .15s, box-shadow .15s; box-sizing: border-box; }
.nsp-input:focus { border-color: var(--primary); box-shadow: 0 0 0 3px hsla(220,100%,60%,0.15); }
.nsp-num { width: 100px !important; }
`;

const S: Record<string, React.CSSProperties> = {
  legendBox: { display: "flex", gap: 10, padding: "12px 14px", background: "rgba(0,0,0,0.18)", border: "1px solid var(--border-color)", borderRadius: 10, alignItems: "flex-start" },
  tokenChip: { fontSize: 11, padding: "3px 9px", borderRadius: 5, background: "rgba(99,102,241,0.12)", color: "#a5b4fc", cursor: "pointer", border: "1px solid rgba(99,102,241,0.2)", transition: "background .12s", userSelect: "none" as const },
  usageCard: { padding: "14px 16px", borderRadius: 10, background: "rgba(0,0,0,0.18)", border: "1px solid var(--border-color)" },
  configCard: { padding: "16px", borderRadius: 10, background: "rgba(0,0,0,0.18)", border: "1px solid var(--border-color)" },
  fieldBox: { display: "flex", flexDirection: "column" as const, gap: 5 },
  fieldLabel: { fontSize: 11, fontWeight: 600, color: "var(--text-muted)" },
  previewRow: { marginTop: 14, padding: "10px 14px", background: "rgba(0,0,0,0.2)", borderRadius: 8, display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" as const },
  previewCode: { color: "var(--primary)", fontWeight: 700, background: "rgba(0,0,0,0.25)", padding: "2px 10px", borderRadius: 5, fontSize: 13 },
  errorBanner: { padding: "10px 14px", background: "rgba(239,68,68,0.12)", border: "1px solid rgba(239,68,68,0.3)", borderRadius: 8, color: "#ef4444", fontSize: 13, display: "flex", gap: 8, alignItems: "center" },
};