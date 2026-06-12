import React, { useState, useEffect, useRef } from "react";
import { request } from "../lib/api";

interface WargaFormData {
  nik: string;
  nama: string;
  tempat_lahir: string;
  tanggal_lahir: string;
  jenis_kelamin: string;
  alamat: string;
  rt: string;
  rw: string;
  kelurahan: string;
  kecamatan: string;
  agama: string;
  status_kawin: string;
  pekerjaan: string;
  kewarganegaraan: string;
  foto_ktp_path: string;
}

export default function WargaDraftComplete() {
  const [token, setToken] = useState("");
  const [draftId, setDraftId] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const [formData, setFormData] = useState<WargaFormData>({
    nik: "",
    nama: "",
    tempat_lahir: "",
    tanggal_lahir: "",
    jenis_kelamin: "L",
    alamat: "",
    rt: "",
    rw: "",
    kelurahan: "",
    kecamatan: "",
    agama: "Islam",
    status_kawin: "Belum Kawin",
    pekerjaan: "",
    kewarganegaraan: "WNI",
    foto_ktp_path: "",
  });

  // RFID scan state
  const [rfidUID, setRfidUID] = useState("");
  const [showRFID, setShowRFID] = useState(false);
  const keypressBuffer = useRef<string[]>([]);
  const lastKeyTime = useRef<number>(0);

  // Extract token from URL query param (?token=xxx)
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const t = params.get("token") || "";
    setToken(t);
    if (!t) {
      setError("Token draft tidak ditemukan. Periksa kembali link yang Anda buka.");
      setLoading(false);
    }
  }, []);

  // Load draft data
  useEffect(() => {
    if (!token) return;

    async function loadDraft() {
      try {
        const data = await request(`/api/warga/draft/${token}`);
        setDraftId(data.id);
        setFormData({
          nik: data.nik || "",
          nama: data.nama || "",
          tempat_lahir: data.tempat_lahir || "",
          tanggal_lahir: data.tanggal_lahir || "",
          jenis_kelamin: data.jenis_kelamin || "L",
          alamat: data.alamat || "",
          rt: data.rt || "",
          rw: data.rw || "",
          kelurahan: data.kelurahan || "",
          kecamatan: data.kecamatan || "",
          agama: data.agama || "Islam",
          status_kawin: data.status_kawin || "Belum Kawin",
          pekerjaan: data.pekerjaan || "",
          kewarganegaraan: data.kewarganegaraan || "WNI",
          foto_ktp_path: data.foto_ktp_path || "",
        });
        if (data.rfid_uid) setRfidUID(data.rfid_uid);
      } catch (err: any) {
        setError(err.message || "Gagal memuat data draft. Link mungkin sudah kadaluarsa.");
      } finally {
        setLoading(false);
      }
    }
    loadDraft();
  }, [token]);

  // Keyboard wedge listener for RFID
  useEffect(() => {
    if (!showRFID) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      const currentTime = new Date().getTime();
      if (currentTime - lastKeyTime.current > 50) {
        keypressBuffer.current = [];
      }
      lastKeyTime.current = currentTime;

      if (e.key === "Enter") {
        if (keypressBuffer.current.length > 0) {
          setRfidUID(keypressBuffer.current.join(""));
          keypressBuffer.current = [];
        }
      } else if (e.key.length === 1) {
        keypressBuffer.current.push(e.key);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [showRFID]);

  const handleComplete = async () => {
    if (!formData.nik || !formData.nama) {
      setError("NIK dan Nama wajib diisi untuk menyelesaikan registrasi.");
      return;
    }

    setSaving(true);
    setError("");

    try {
      await request(`/api/warga/draft/${token}/complete`, {
        method: "PUT",
        body: JSON.stringify({
          ...formData,
          rfid_uid: rfidUID || undefined,
        }),
      });

      setSuccess(true);
      setTimeout(() => {
        window.location.href = "/warga";
      }, 2000);
    } catch (err: any) {
      setError(err.message || "Gagal menyelesaikan registrasi.");
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div style={{ maxWidth: "800px", margin: "0 auto", textAlign: "center", padding: "60px 20px" }}>
        <div className="spinner" style={{ margin: "0 auto 16px auto" }}></div>
        <p style={{ color: "var(--text-muted)" }}>Memuat data draft...</p>
      </div>
    );
  }

  if (success) {
    return (
      <div style={{ maxWidth: "600px", margin: "0 auto", textAlign: "center", padding: "60px 20px" }}>
        <div style={{ fontSize: "64px", marginBottom: "16px" }}>✅</div>
        <h2 style={{ fontSize: "24px", fontWeight: "700", marginBottom: "8px" }}>Registrasi Berhasil!</h2>
        <p style={{ color: "var(--text-muted)" }}>Data warga telah berhasil disimpan. Mengalihkan ke daftar warga...</p>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: "800px", margin: "0 auto" }}>
      <div style={{ marginBottom: "32px" }}>
        <a href="/warga" style={{ color: "var(--text-muted)", fontSize: "14px", textDecoration: "none" }}>
          ← Kembali ke Daftar Warga
        </a>
        <h1 style={{ fontSize: "28px", fontWeight: "700", marginTop: "8px" }}>Lengkapi Data Warga</h1>
        <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
          Lengkapi data berikut untuk menyelesaikan registrasi warga.
          <span style={{ display: "inline-block", marginLeft: "8px", padding: "2px 10px", background: "rgba(251,191,36,0.2)", color: "#fbbf24", borderRadius: "12px", fontSize: "12px", fontWeight: "600" }}>
            Draft
          </span>
        </p>
      </div>

      {error && (
        <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)", marginBottom: "24px" }}>
          ⚠️ {error}
        </div>
      )}

      {/* KTP Photo Preview */}
      {formData.foto_ktp_path && (
        <div className="glass-card" style={{ marginBottom: "24px", textAlign: "center" }}>
          <p style={{ fontSize: "13px", color: "var(--text-muted)", marginBottom: "8px" }}>Foto KTP yang diunggah sebelumnya:</p>
          <img
            src={formData.foto_ktp_path}
            alt="Foto KTP"
            style={{ maxWidth: "300px", borderRadius: "8px", border: "1px solid var(--border-color)" }}
          />
        </div>
      )}

      {/* Verification Form */}
      <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "24px" }}>
        <h3 style={{ fontSize: "18px", fontWeight: "700" }}>Data Penduduk</h3>

        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))", gap: "20px" }}>
          <div className="form-group">
            <label className="form-label">NIK (Nomor Induk Kependudukan) <span style={{ color: "var(--danger)" }}>*</span></label>
            <input type="text" className="form-control" value={formData.nik}
              onChange={(e) => setFormData({ ...formData, nik: e.target.value })} maxLength={16} />
          </div>

          <div className="form-group">
            <label className="form-label">Nama Lengkap <span style={{ color: "var(--danger)" }}>*</span></label>
            <input type="text" className="form-control" value={formData.nama}
              onChange={(e) => setFormData({ ...formData, nama: e.target.value })} />
          </div>

          <div className="form-group">
            <label className="form-label">Tempat Lahir</label>
            <input type="text" className="form-control" value={formData.tempat_lahir}
              onChange={(e) => setFormData({ ...formData, tempat_lahir: e.target.value })} />
          </div>

          <div className="form-group">
            <label className="form-label">Tanggal Lahir (YYYY-MM-DD)</label>
            <input type="text" className="form-control" value={formData.tanggal_lahir} placeholder="YYYY-MM-DD"
              onChange={(e) => setFormData({ ...formData, tanggal_lahir: e.target.value })} />
          </div>

          <div className="form-group">
            <label className="form-label">Jenis Kelamin</label>
            <select className="form-control" value={formData.jenis_kelamin}
              onChange={(e) => setFormData({ ...formData, jenis_kelamin: e.target.value })}>
              <option value="L">Laki-laki</option>
              <option value="P">Perempuan</option>
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Agama</label>
            <input type="text" className="form-control" value={formData.agama}
              onChange={(e) => setFormData({ ...formData, agama: e.target.value })} />
          </div>

          <div className="form-group" style={{ gridColumn: "span 2" }}>
            <label className="form-label">Alamat</label>
            <input type="text" className="form-control" value={formData.alamat}
              onChange={(e) => setFormData({ ...formData, alamat: e.target.value })} />
          </div>

          <div className="form-group">
            <label className="form-label">RT</label>
            <input type="text" className="form-control" value={formData.rt}
              onChange={(e) => setFormData({ ...formData, rt: e.target.value })} />
          </div>

          <div className="form-group">
            <label className="form-label">RW</label>
            <input type="text" className="form-control" value={formData.rw}
              onChange={(e) => setFormData({ ...formData, rw: e.target.value })} />
          </div>

          <div className="form-group">
            <label className="form-label">Kelurahan / Desa</label>
            <input type="text" className="form-control" value={formData.kelurahan}
              onChange={(e) => setFormData({ ...formData, kelurahan: e.target.value })} />
          </div>

          <div className="form-group">
            <label className="form-label">Kecamatan</label>
            <input type="text" className="form-control" value={formData.kecamatan}
              onChange={(e) => setFormData({ ...formData, kecamatan: e.target.value })} />
          </div>

          <div className="form-group">
            <label className="form-label">Pekerjaan</label>
            <input type="text" className="form-control" value={formData.pekerjaan}
              onChange={(e) => setFormData({ ...formData, pekerjaan: e.target.value })} />
          </div>

          <div className="form-group">
            <label className="form-label">Status Perkawinan</label>
            <input type="text" className="form-control" value={formData.status_kawin}
              onChange={(e) => setFormData({ ...formData, status_kawin: e.target.value })} />
          </div>
        </div>
      </div>

      {/* RFID Section */}
      <div className="glass-card" style={{ marginTop: "24px", display: "flex", flexDirection: "column", gap: "16px", alignItems: "center", padding: "32px" }}>
        <h3 style={{ fontSize: "18px", fontWeight: "700" }}>Tautkan Kartu RFID (Opsional)</h3>

        {!showRFID ? (
          <button className="btn btn-secondary" onClick={() => setShowRFID(true)}>
            + Tautkan Kartu RFID
          </button>
        ) : (
          <div style={{ width: "100%", maxWidth: "360px", padding: "32px 20px", background: "hsla(222,47%,7%,0.6)", border: "2px dashed var(--border-color)", borderRadius: "var(--radius-md)", textAlign: "center" }}>
            {rfidUID ? (
              <div>
                <span className="badge badge-success" style={{ marginBottom: "8px" }}>Kartu Terdeteksi</span>
                <h4 style={{ fontSize: "24px", fontWeight: "800", color: "var(--text-main)" }}>{rfidUID}</h4>
                <button className="btn btn-secondary" style={{ marginTop: "12px", padding: "6px 12px", fontSize: "13px" }} onClick={() => setRfidUID("")}>
                  Scan Ulang
                </button>
              </div>
            ) : (
              <div>
                <div className="spinner" style={{ margin: "0 auto 12px auto" }}></div>
                <span style={{ fontSize: "14px", color: "var(--text-muted)", fontWeight: "500" }}>Menunggu scan kartu...</span>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Actions */}
      <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px", marginTop: "24px", marginBottom: "40px" }}>
        <a href="/warga" className="btn btn-secondary" style={{ textDecoration: "none" }}>Batal</a>
        <button className="btn btn-primary" onClick={handleComplete} disabled={saving || !formData.nik || !formData.nama}>
          {saving ? "Menyimpan..." : "Selesaikan Registrasi"}
        </button>
      </div>
    </div>
  );
}
