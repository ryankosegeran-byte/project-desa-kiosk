import React, { useEffect, useState } from "react";
import { request, getUser } from "../lib/api";

interface Desa {
  id: string;
  nama: string;
  kode_desa: string;
  kecamatan?: string;
  kabupaten?: string;
  provinsi?: string;
  kepala_desa?: string;
  nip_kepala_desa?: string;
  alamat_kantor?: string;
  theme?: string;
}

const THEME_OPTIONS = [
  { value: "merah-putih", label: "Merah Putih" },
  { value: "dark-blue", label: "Dark Blue" },
];

export default function DesaManager() {
  const user = getUser();
  const [desas, setDesas] = useState<Desa[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Create modal state
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    nama: "",
    kode_desa: "",
    kecamatan: "",
    kabupaten: "",
    provinsi: "",
    kepala_desa: "",
    nip_kepala_desa: "",
    alamat_kantor: "",
    theme: "merah-putih",
  });
  const [themeSavingId, setThemeSavingId] = useState<string | null>(null);
  const [saveLoading, setSaveLoading] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleCopyId = async (id: string) => {
    try {
      await navigator.clipboard.writeText(id);
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 1500);
    } catch {
      // fallback
      const ta = document.createElement("textarea");
      ta.value = id;
      document.body.appendChild(ta);
      ta.select();
      document.execCommand("copy");
      document.body.removeChild(ta);
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 1500);
    }
  };

  useEffect(() => {
    loadDesa();
  }, []);

  async function loadDesa() {
    try {
      const data = await request("/api/desa");
      setDesas(data);
    } catch (err: any) {
      setError(err.message || "Gagal mengambil data desa.");
    } finally {
      setLoading(false);
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaveLoading(true);
    setSaveError("");

    try {
      const created = await request("/api/desa", {
        method: "POST",
        body: JSON.stringify(formData),
      });

      setDesas([...desas, created]);
      setShowModal(false);
      setFormData({
        nama: "",
        kode_desa: "",
        kecamatan: "",
        kabupaten: "",
        provinsi: "",
        kepala_desa: "",
        nip_kepala_desa: "",
        alamat_kantor: "",
        theme: "merah-putih",
      });
    } catch (err: any) {
      setSaveError(err.message || "Gagal membuat profil desa.");
    } finally {
      setSaveLoading(false);
    }
  };

  const handleThemeChange = async (id: string, theme: string) => {
    setThemeSavingId(id);
    setDesas((prev) => prev.map((d) => (d.id === id ? { ...d, theme } : d)));
    try {
      await request(`/api/desa/${id}/theme`, {
        method: "PUT",
        body: JSON.stringify({ theme }),
      });
    } catch (err: any) {
      setError(err.message || "Gagal memperbarui tema kiosk.");
    } finally {
      setThemeSavingId(null);
    }
  };

  if (loading) {
    return (
      <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "400px" }}>
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "32px" }}>
        <div>
          <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Kelola Desa</h1>
          <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
            Daftarkan dan perbarui wilayah desa pengguna kiosk pelayanan terdistribusi.
          </p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          ➕ Daftarkan Desa Baru
        </button>
      </div>

      {error && (
        <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)", marginBottom: "24px" }}>
          ⚠️ {error}
        </div>
      )}

      {/* Villages Table */}
      <div className="table-container">
        <table className="premium-table">
          <thead>
            <tr>
              {user?.role === "superadmin" && <th>ID Desa</th>}
              <th>Kode Desa</th>
              <th>Nama Desa</th>
              <th>Kecamatan</th>
              <th>Kabupaten</th>
              <th>Kepala Desa</th>
              <th>NIP Kepala Desa</th>
              <th>Tema Kiosk</th>
            </tr>
          </thead>
          <tbody>
            {desas.map((d) => (
              <tr key={d.id}>
                {user?.role === "superadmin" && (
                  <td style={{ fontFamily: "monospace", fontSize: "13px", whiteSpace: "nowrap" }}>
                    <span style={{ color: "var(--text-muted)" }}>{d.id.slice(0, 8)}...</span>
                    <button
                      onClick={() => handleCopyId(d.id)}
                      title={copiedId === d.id ? "Tersalin!" : "Salin UUID"}
                      style={{
                        background: "none",
                        border: "none",
                        cursor: "pointer",
                        padding: "2px 6px",
                        marginLeft: "4px",
                        borderRadius: "4px",
                        color: copiedId === d.id ? "var(--success, #10b981)" : "var(--text-muted)",
                        fontSize: "14px",
                        transition: "color 0.2s",
                        verticalAlign: "middle",
                      }}
                    >
                      {copiedId === d.id ? "✓" : "📋"}
                    </button>
                    {copiedId === d.id && (
                      <span style={{
                        fontSize: "11px",
                        color: "var(--success, #10b981)",
                        marginLeft: "4px",
                        verticalAlign: "middle",
                      }}>
                        Tersalin!
                      </span>
                    )}
                  </td>
                )}
                <td style={{ fontWeight: "600", fontFamily: "monospace" }}>{d.kode_desa}</td>
                <td>{d.nama}</td>
                <td>{d.kecamatan || "-"}</td>
                <td>{d.kabupaten || "-"}</td>
                <td>{d.kepala_desa || "-"}</td>
                <td>{d.nip_kepala_desa || "-"}</td>
                <td>
                  <select
                    className="form-control"
                    style={{ minWidth: "140px", padding: "6px 10px" }}
                    value={d.theme || "merah-putih"}
                    disabled={themeSavingId === d.id}
                    onChange={(e) => handleThemeChange(d.id, e.target.value)}
                  >
                    {THEME_OPTIONS.map((t) => (
                      <option key={t.value} value={t.value}>
                        {t.label}
                      </option>
                    ))}
                  </select>
                </td>
              </tr>
            ))}
            {desas.length === 0 && (
              <tr>
                <td colSpan={user?.role === "superadmin" ? 8 : 7} style={{ textAlign: "center", color: "var(--text-muted)", padding: "24px" }}>
                  Belum ada profil desa terdaftar
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Register Village Modal */}
      {showModal && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "var(--overlay)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "600px", width: "95%", maxHeight: "90vh", overflowY: "auto", padding: "30px" }}>
            <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "20px" }}>Daftarkan Desa Baru</h3>

            {saveError && (
              <div style={{ color: "var(--danger)", background: "hsla(355, 85%, 55%, 0.1)", padding: "10px", borderRadius: "var(--radius-sm)", marginBottom: "16px", fontSize: "13px" }}>
                ⚠️ {saveError}
              </div>
            )}

            <form onSubmit={handleSubmit}>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px" }}>
                <div className="form-group">
                  <label className="form-label">Nama Desa</label>
                  <input
                    type="text"
                    className="form-control"
                    placeholder="Contoh: Desa Cibunar"
                    value={formData.nama}
                    onChange={(e) => setFormData({ ...formData, nama: e.target.value })}
                    required
                    disabled={saveLoading}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Kode Desa</label>
                  <input
                    type="text"
                    className="form-control"
                    placeholder="Contoh: 32.05.11.2001"
                    value={formData.kode_desa}
                    onChange={(e) => setFormData({ ...formData, kode_desa: e.target.value })}
                    required
                    disabled={saveLoading}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Kecamatan</label>
                  <input
                    type="text"
                    className="form-control"
                    value={formData.kecamatan}
                    onChange={(e) => setFormData({ ...formData, kecamatan: e.target.value })}
                    disabled={saveLoading}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Kabupaten</label>
                  <input
                    type="text"
                    className="form-control"
                    value={formData.kabupaten}
                    onChange={(e) => setFormData({ ...formData, kabupaten: e.target.value })}
                    disabled={saveLoading}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Provinsi</label>
                  <input
                    type="text"
                    className="form-control"
                    value={formData.provinsi}
                    onChange={(e) => setFormData({ ...formData, provinsi: e.target.value })}
                    disabled={saveLoading}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Kepala Desa (Kades)</label>
                  <input
                    type="text"
                    className="form-control"
                    value={formData.kepala_desa}
                    onChange={(e) => setFormData({ ...formData, kepala_desa: e.target.value })}
                    disabled={saveLoading}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">NIP Kepala Desa</label>
                  <input
                    type="text"
                    className="form-control"
                    value={formData.nip_kepala_desa}
                    onChange={(e) => setFormData({ ...formData, nip_kepala_desa: e.target.value })}
                    disabled={saveLoading}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Tema Kiosk</label>
                  <select
                    className="form-control"
                    value={formData.theme}
                    onChange={(e) => setFormData({ ...formData, theme: e.target.value })}
                    disabled={saveLoading}
                  >
                    {THEME_OPTIONS.map((t) => (
                      <option key={t.value} value={t.value}>
                        {t.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="form-group">
                <label className="form-label">Alamat Kantor Desa</label>
                <textarea
                  className="form-control"
                  style={{ height: "80px", resize: "none" }}
                  value={formData.alamat_kantor}
                  onChange={(e) => setFormData({ ...formData, alamat_kantor: e.target.value })}
                  disabled={saveLoading}
                />
              </div>

              <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px", marginTop: "24px" }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)} disabled={saveLoading}>
                  Batal
                </button>
                <button type="submit" className="btn btn-primary" disabled={saveLoading}>
                  {saveLoading ? "Menyimpan..." : "Daftarkan"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
