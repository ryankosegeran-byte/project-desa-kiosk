import React, { useEffect, useState, useRef } from "react";
import { request } from "../lib/api";

interface Warga {
  id: string;
  nik: string;
  nama: string;
  tempat_lahir?: string;
  tanggal_lahir?: string;
  jenis_kelamin?: string;
  alamat?: string;
  rt?: string;
  rw?: string;
  rfid_uid?: string;
  status?: string;
  draft_token?: string;
}

export default function WargaList() {
  const [warga, setWarga] = useState<Warga[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  
  // RFID linking modal state
  const [linkingWarga, setLinkingWarga] = useState<Warga | null>(null);
  const [rfidUID, setRfidUID] = useState("");
  const [linkLoading, setLinkLoading] = useState(false);
  const [linkError, setLinkError] = useState("");

  const keypressBuffer = useRef<string[]>([]);
  const lastKeyTime = useRef<number>(0);

  useEffect(() => {
    loadWarga();
  }, []);

  // Keyboard wedge listener for RFID reader
  useEffect(() => {
    if (!linkingWarga) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      const currentTime = timeNow();
      if (currentTime - lastKeyTime.current > 50) {
        keypressBuffer.current = [];
      }
      lastKeyTime.current = currentTime;

      if (e.key === "Enter") {
        if (keypressBuffer.current.length > 0) {
          const scanned = keypressBuffer.current.join("");
          setRfidUID(scanned);
          keypressBuffer.current = [];
        }
      } else if (e.key.length === 1) {
        keypressBuffer.current.push(e.key);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [linkingWarga]);

  const timeNow = () => new Date().getTime();

  async function loadWarga() {
    try {
      const data = await request("/api/warga");
      setWarga(data);
    } catch (err: any) {
      setError(err.message || "Gagal mengambil data warga.");
    } finally {
      setLoading(false);
    }
  }

  const handleLinkRFID = async () => {
    if (!linkingWarga || !rfidUID) return;
    setLinkLoading(true);
    setLinkError("");

    try {
      await request(`/api/warga/${linkingWarga.id}/rfid`, {
        method: "PUT",
        body: JSON.stringify({ rfid_uid: rfidUID }),
      });
      // reload warga
      await loadWarga();
      setLinkingWarga(null);
      setRfidUID("");
    } catch (err: any) {
      setLinkError(err.message || "Gagal menautkan kartu RFID.");
    } finally {
      setLinkLoading(false);
    }
  };

  const filteredWarga = warga.filter((w) =>
    (w.nama || "").toLowerCase().includes(search.toLowerCase()) ||
    (w.nik || "").includes(search)
  );

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
          <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Data Warga Desa</h1>
          <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
            Kelola profil kependudukan warga desa yang sinkron dengan kiosk pelayanan mandiri.
          </p>
        </div>
        <a href="/warga/register" className="btn btn-primary">
          ➕ Registrasi Warga Baru
        </a>
      </div>

      {error && (
        <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)", marginBottom: "24px" }}>
          ⚠️ {error}
        </div>
      )}

      {/* Filter / Search Bar */}
      <div className="glass-card" style={{ padding: "16px", marginBottom: "24px" }}>
        <input
          type="text"
          className="form-control"
          placeholder="🔍 Cari nama atau NIK warga..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {/* Table grid */}
      <div className="table-container">
        <table className="premium-table">
          <thead>
            <tr>
              <th>NIK</th>
              <th>Nama Lengkap</th>
              <th>Tempat / Tgl Lahir</th>
              <th>Jenis Kelamin</th>
              <th>Alamat Lengkap</th>
              <th>Kartu RFID / KTP</th>
              <th>Status</th>
              <th>Aksi</th>
            </tr>
          </thead>
          <tbody>
            {filteredWarga.map((w) => (
              <tr key={w.id} style={w.status === "draft" ? { opacity: 0.75 } : {}}>
                <td style={{ fontWeight: "600" }}>{w.nik || <em style={{ color: "var(--text-muted)" }}>belum diisi</em>}</td>
                <td>{w.nama || <em style={{ color: "var(--text-muted)" }}>belum diisi</em>}</td>
                <td>{w.tempat_lahir}, {w.tanggal_lahir}</td>
                <td>{w.jenis_kelamin === "L" ? "Laki-laki" : w.jenis_kelamin === "P" ? "Perempuan" : "-"}</td>
                <td style={{ maxWidth: "250px", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                  {w.alamat} {w.rt ? `RT ${w.rt}` : ""} {w.rw ? `/ RW ${w.rw}` : ""}
                </td>
                <td>
                  {w.rfid_uid ? (
                    <span className="badge badge-success">{w.rfid_uid}</span>
                  ) : (
                    <span className="badge badge-warning">Belum Ada</span>
                  )}
                </td>
                <td>
                  {w.status === "draft" ? (
                    <span style={{ display: "inline-block", padding: "3px 10px", background: "rgba(251,191,36,0.15)", color: "#fbbf24", borderRadius: "12px", fontSize: "12px", fontWeight: "600" }}>Draft</span>
                  ) : (
                    <span style={{ display: "inline-block", padding: "3px 10px", background: "rgba(34,197,94,0.15)", color: "#22c55e", borderRadius: "12px", fontSize: "12px", fontWeight: "600" }}>Lengkap</span>
                  )}
                </td>
                <td>
                  {w.status === "draft" && w.draft_token ? (
                    <a href={`/warga/draft?token=${w.draft_token}`} className="btn btn-primary" style={{ padding: "6px 12px", fontSize: "13px", textDecoration: "none" }}>
                      ✏️ Lanjutkan
                    </a>
                  ) : (
                    <button className="btn btn-secondary" style={{ padding: "6px 12px", fontSize: "13px" }} onClick={() => setLinkingWarga(w)}>
                      💳 Tautkan Kartu
                    </button>
                  )}
                </td>
              </tr>
            ))}
            {filteredWarga.length === 0 && (
              <tr>
                <td colSpan={8} style={{ textAlign: "center", color: "var(--text-muted)", padding: "24px" }}>
                  Warga tidak ditemukan
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* RFID Link Modal */}
      {linkingWarga && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "rgba(0,0,0,0.7)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "450px", width: "100%", padding: "30px" }}>
            <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "8px" }}>Tautkan Kartu KTP (RFID)</h3>
            <p style={{ color: "var(--text-muted)", fontSize: "14px", marginBottom: "20px" }}>
              Tautkan ID kartu RFID baru ke data warga: <strong>{linkingWarga.nama}</strong> ({linkingWarga.nik})
            </p>

            {linkError && (
              <div style={{ color: "var(--danger)", background: "hsla(355, 85%, 55%, 0.1)", padding: "10px", borderRadius: "var(--radius-sm)", marginBottom: "16px", fontSize: "13px" }}>
                ⚠️ {linkError}
              </div>
            )}

            <div style={{ textAlign: "center", padding: "30px 20px", background: "hsla(222,47%,7%,0.8)", border: "1px dashed var(--border-color)", borderRadius: "var(--radius-sm)", marginBottom: "20px" }}>
              {rfidUID ? (
                <div>
                  <span style={{ fontSize: "12px", color: "var(--success)", fontWeight: "600", textTransform: "uppercase" }}>Kartu Terbaca</span>
                  <div style={{ fontSize: "24px", fontWeight: "800", color: "var(--text-main)", marginTop: "4px" }}>{rfidUID}</div>
                </div>
              ) : (
                <div>
                  <div className="spinner" style={{ margin: "0 auto 12px auto" }}></div>
                  <span style={{ fontSize: "14px", color: "var(--text-muted)" }}>Tempelkan Kartu KTP pada RFID Reader...</span>
                </div>
              )}
            </div>

            <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px" }}>
              <button className="btn btn-secondary" onClick={() => { setLinkingWarga(null); setRfidUID(""); }} disabled={linkLoading}>
                Batal
              </button>
              <button className="btn btn-primary" onClick={handleLinkRFID} disabled={!rfidUID || linkLoading}>
                {linkLoading ? "Menyimpan..." : "Simpan Tautan"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
