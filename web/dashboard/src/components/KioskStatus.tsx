import React, { useEffect, useState } from "react";
import { request, getUser } from "../lib/api";

interface Kiosk {
  id: string;
  desa_id: string;
  nama: string;
  api_key?: string;
  last_seen_at?: string;
  last_sync_at?: string;
  status: string;
  ip_address?: string;
  created_at: string;
}

interface Desa {
  id: string;
  nama: string;
}

export default function KioskStatus() {
  const [kiosks, setKiosks] = useState<Kiosk[]>([]);
  const [desas, setDesas] = useState<Desa[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  
  // Registration modal states
  const [showModal, setShowModal] = useState(false);
  const [newKioskName, setNewKioskName] = useState("");
  const [newKioskDesaId, setNewKioskDesaId] = useState("");
  const [registeredKiosk, setRegisteredKiosk] = useState<Kiosk | null>(null);
  const [registerLoading, setRegisterLoading] = useState(false);
  const [registerError, setRegisterError] = useState("");

  // Detail modal states
  const [detailKiosk, setDetailKiosk] = useState<Kiosk | null>(null);
  const [showApiKey, setShowApiKey] = useState(false);
  const [copyFeedback, setCopyFeedback] = useState<string | null>(null);

  const user = getUser();

  useEffect(() => {
    async function loadData() {
      try {
        const data = await request("/api/kiosks");
        setKiosks(data);

        if (user?.role === "superadmin") {
          const desaData = await request("/api/desa");
          setDesas(desaData);
          if (desaData.length > 0) {
            setNewKioskDesaId(desaData[0].id);
          }
        } else {
          setNewKioskDesaId(user?.desa_id || "");
        }
      } catch (err: any) {
        setError(err.message || "Gagal mengambil data kiosk.");
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, []);

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setRegisterLoading(true);
    setRegisterError("");

    try {
      const created = await request("/api/kiosks", {
        method: "POST",
        body: JSON.stringify({
          nama: newKioskName,
          desa_id: newKioskDesaId,
        }),
      });

      setKiosks([...kiosks, created]);
      setRegisteredKiosk(created);
      setNewKioskName("");
    } catch (err: any) {
      setRegisterError(err.message || "Gagal mendaftarkan kiosk.");
    } finally {
      setRegisterLoading(false);
    }
  };

  const getStatusBadge = (k: Kiosk) => {
    if (!k.last_seen_at) return <span className="badge badge-warning">OFFLINE</span>;
    
    const lastSeen = new Date(k.last_seen_at).getTime();
    const now = new Date().getTime();
    
    // Online if pinged in the last 2 minutes
    const isOnline = now - lastSeen < 120 * 1000;
    
    if (isOnline) {
      return <span className="badge badge-success">ONLINE</span>;
    }
    return <span className="badge badge-danger">OFFLINE</span>;
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
          <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Status Kiosk Desa</h1>
          <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
            Monitor terminal kiosk pelayanan mandiri yang terpasang di kantor desa.
          </p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          ➕ Daftarkan Kiosk Baru
        </button>
      </div>

      {error && (
        <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)", marginBottom: "24px" }}>
          ⚠️ {error}
        </div>
      )}

      {/* Kiosks Table */}
      <div className="table-container">
        <table className="premium-table">
          <thead>
            <tr>
              <th>ID Kiosk</th>
              <th>Nama Kiosk</th>
              <th>IP Address</th>
              <th>Status Koneksi</th>
              <th>Heartbeat Terakhir</th>
              <th>Sinkronisasi Terakhir</th>
            </tr>
          </thead>
          <tbody>
            {kiosks.map((k) => (
              <tr
                key={k.id}
                onClick={() => { setDetailKiosk(k); setShowApiKey(false); }}
                style={{ cursor: "pointer" }}
              >
                <td style={{ fontWeight: "600", fontFamily: "monospace" }}>{k.id.substring(0, 8)}...</td>
                <td>{k.nama}</td>
                <td style={{ fontFamily: "monospace" }}>{k.ip_address || "-"}</td>
                <td>{getStatusBadge(k)}</td>
                <td>
                  {k.last_seen_at
                    ? new Date(k.last_seen_at).toLocaleString("id-ID", { dateStyle: "short", timeStyle: "short" })
                    : "-"}
                </td>
                <td>
                  {k.last_sync_at
                    ? new Date(k.last_sync_at).toLocaleString("id-ID", { dateStyle: "short", timeStyle: "short" })
                    : "-"}
                </td>
              </tr>
            ))}
            {kiosks.length === 0 && (
              <tr>
                <td colSpan={6} style={{ textAlign: "center", color: "var(--text-muted)", padding: "24px" }}>
                  Belum ada terminal kiosk terdaftar
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Registration Modal */}
      {showModal && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "rgba(0,0,0,0.7)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "500px", width: "100%", padding: "30px" }}>
            <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "16px" }}>
              {registeredKiosk ? "Pendaftaran Berhasil!" : "Daftarkan Kiosk Baru"}
            </h3>

            {registerError && (
              <div style={{ color: "var(--danger)", background: "hsla(355, 85%, 55%, 0.1)", padding: "10px", borderRadius: "var(--radius-sm)", marginBottom: "16px", fontSize: "13px" }}>
                ⚠️ {registerError}
              </div>
            )}

            {registeredKiosk ? (
              <div>
                <p style={{ fontSize: "14px", color: "var(--text-muted)", marginBottom: "16px" }}>
                  Kiosk telah terdaftar. Gunakan API Key rahasia di bawah ini untuk mengonfigurasi kiosk lokal Anda (simpan file <code>.env</code>):
                </p>
                <div style={{ background: "var(--bg-main)", padding: "16px", borderRadius: "var(--radius-sm)", border: "1px solid var(--border-color)", marginBottom: "20px" }}>
                  <span style={{ fontSize: "11px", fontWeight: "600", color: "var(--text-muted)" }}>API KEY KIOSK (RAHASIA)</span>
                  <div style={{ fontSize: "16px", fontWeight: "800", color: "var(--success)", wordBreak: "break-all", marginTop: "4px", fontFamily: "monospace" }}>
                    {registeredKiosk.api_key}
                  </div>
                </div>
                <div style={{ display: "flex", justifyContent: "flex-end" }}>
                  <button className="btn btn-primary" onClick={() => { setShowModal(false); setRegisteredKiosk(null); }}>
                    Selesai
                  </button>
                </div>
              </div>
            ) : (
              <form onSubmit={handleRegister}>
                <div className="form-group">
                  <label className="form-label">Nama Terminal Kiosk</label>
                  <input
                    type="text"
                    className="form-control"
                    placeholder="Contoh: Kiosk Kantor Desa R1"
                    value={newKioskName}
                    onChange={(e) => setNewKioskName(e.target.value)}
                    required
                    disabled={registerLoading}
                  />
                </div>

                {user?.role === "superadmin" && desas.length > 0 && (
                  <div className="form-group">
                    <label className="form-label">Desa Penempatan</label>
                    <select
                      className="form-control"
                      value={newKioskDesaId}
                      onChange={(e) => setNewKioskDesaId(e.target.value)}
                      disabled={registerLoading}
                    >
                      {desas.map((d) => (
                        <option key={d.id} value={d.id}>
                          {d.nama}
                        </option>
                      ))}
                    </select>
                  </div>
                )}

                <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px", marginTop: "24px" }}>
                  <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)} disabled={registerLoading}>
                    Batal
                  </button>
                  <button type="submit" className="btn btn-primary" disabled={registerLoading}>
                    {registerLoading ? "Mendaftarkan..." : "Daftarkan"}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}

      {/* Detail Modal */}
      {detailKiosk && (
        <div
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: "rgba(0,0,0,0.7)",
            backdropFilter: "blur(4px)",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            zIndex: 1000,
          }}
          onClick={() => setDetailKiosk(null)}
        >
          <div
            className="glass-card"
            style={{ maxWidth: "520px", width: "100%", padding: "30px" }}
            onClick={(e) => e.stopPropagation()}
          >
            <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "20px" }}>Detail Kiosk</h3>

            <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
              {/* Nama Kiosk */}
              <div>
                <div style={{ fontSize: "11px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase", letterSpacing: "0.05em", marginBottom: "4px" }}>
                  Nama Kiosk
                </div>
                <div style={{ fontSize: "16px", fontWeight: "600" }}>{detailKiosk.nama}</div>
              </div>

              {/* Full ID Kiosk */}
              <div>
                <div style={{ fontSize: "11px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase", letterSpacing: "0.05em", marginBottom: "4px" }}>
                  ID Kiosk
                </div>
                <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                  <code style={{ fontSize: "14px", fontFamily: "monospace", wordBreak: "break-all" }}>{detailKiosk.id}</code>
                  <button
                    className="btn btn-secondary"
                    style={{ padding: "4px 10px", fontSize: "12px" }}
                    onClick={() => {
                      navigator.clipboard.writeText(detailKiosk.id);
                      setCopyFeedback("id");
                      setTimeout(() => setCopyFeedback((prev) => (prev === "id" ? null : prev)), 1500);
                    }}
                  >
                    {copyFeedback === "id" ? "Tersalin!" : "Salin"}
                  </button>
                </div>
              </div>

              {/* API Key — admin only */}
              {user?.role === "superadmin" && (
                <div>
                  <div style={{ fontSize: "11px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase", letterSpacing: "0.05em", marginBottom: "4px" }}>
                    API Key
                  </div>
                  <div
                    style={{
                      background: "var(--bg-main)",
                      padding: "12px 14px",
                      borderRadius: "var(--radius-sm)",
                      border: "1px solid var(--border-color)",
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "space-between",
                      gap: "10px",
                    }}
                  >
                    <code
                      style={{
                        fontSize: "14px",
                        fontFamily: "monospace",
                        wordBreak: "break-all",
                        color: showApiKey ? "var(--success)" : "inherit",
                      }}
                    >
                      {showApiKey
                        ? detailKiosk.api_key || "-"
                        : detailKiosk.api_key
                        ? detailKiosk.api_key.substring(0, 6) + "****...****" + detailKiosk.api_key.slice(-4)
                        : "-"}
                    </code>
                    <div style={{ display: "flex", gap: "6px", flexShrink: 0 }}>
                      <button
                        className="btn btn-secondary"
                        style={{ padding: "4px 10px", fontSize: "12px" }}
                        onClick={() => setShowApiKey((prev) => !prev)}
                        title={showApiKey ? "Sembunyikan" : "Tampilkan"}
                      >
                        {showApiKey ? "🙈" : "👁️"}
                      </button>
                      <button
                        className="btn btn-secondary"
                        style={{ padding: "4px 10px", fontSize: "12px" }}
                        onClick={() => {
                          if (detailKiosk.api_key) {
                            navigator.clipboard.writeText(detailKiosk.api_key);
                            setCopyFeedback("apiKey");
                            setTimeout(() => setCopyFeedback((prev) => (prev === "apiKey" ? null : prev)), 1500);
                          }
                        }}
                        disabled={!detailKiosk.api_key}
                      >
                        {copyFeedback === "apiKey" ? "Tersalin!" : "Salin"}
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* Status */}
              <div>
                <div style={{ fontSize: "11px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase", letterSpacing: "0.05em", marginBottom: "4px" }}>
                  Status
                </div>
                <div>{getStatusBadge(detailKiosk)}</div>
              </div>

              {/* Last Heartbeat */}
              <div>
                <div style={{ fontSize: "11px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase", letterSpacing: "0.05em", marginBottom: "4px" }}>
                  Heartbeat Terakhir
                </div>
                <div style={{ fontSize: "14px" }}>
                  {detailKiosk.last_seen_at
                    ? new Date(detailKiosk.last_seen_at).toLocaleString("id-ID", { dateStyle: "long", timeStyle: "medium" })
                    : "-"}
                </div>
              </div>

              {/* Last Sync */}
              <div>
                <div style={{ fontSize: "11px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase", letterSpacing: "0.05em", marginBottom: "4px" }}>
                  Sinkronisasi Terakhir
                </div>
                <div style={{ fontSize: "14px" }}>
                  {detailKiosk.last_sync_at
                    ? new Date(detailKiosk.last_sync_at).toLocaleString("id-ID", { dateStyle: "long", timeStyle: "medium" })
                    : "-"}
                </div>
              </div>
            </div>

            <div style={{ display: "flex", justifyContent: "flex-end", marginTop: "24px" }}>
              <button className="btn btn-primary" onClick={() => setDetailKiosk(null)}>
                Tutup
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
