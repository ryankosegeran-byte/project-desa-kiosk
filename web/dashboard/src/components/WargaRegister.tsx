import React, { useState, useEffect, useRef } from "react";
import { request, getUser } from "../lib/api";

interface KTPData {
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
  confidence: number;
}

interface Desa {
  id: string;
  nama: string;
  kode_desa: string;
  kecamatan?: string;
  kabupaten?: string;
}

export default function WargaRegister() {
  const user = getUser();
  const [step, setStep] = useState(1);
  const [file, setFile] = useState<File | null>(null);
  const [previewURL, setPreviewURL] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Desa selection modal state
  const [showDesaModal, setShowDesaModal] = useState(true);
  const [desaList, setDesaList] = useState<Desa[]>([]);
  const [desaLoading, setDesaLoading] = useState(true);
  const [selectedDesaId, setSelectedDesaId] = useState("");
  const [selectedDesaName, setSelectedDesaName] = useState("");

  // Draft link state
  const [draftLink, setDraftLink] = useState("");
  const [copyFeedback, setCopyFeedback] = useState("");

  // Extracted verified warga fields
  const [wargaData, setWargaData] = useState<KTPData>({
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
    confidence: 0,
  });

  // RFID scan state
  const [rfidUID, setRfidUID] = useState("");
  const keypressBuffer = useRef<string[]>([]);
  const lastKeyTime = useRef<number>(0);

  // Fetch desa list on mount
  useEffect(() => {
    async function fetchDesa() {
      try {
        if (user?.role === "pic_desa") {
          // PIC desa: auto-select their desa (no API call needed, /api/desa is superadmin only)
          setSelectedDesaId(user?.desa_id || "");
          setSelectedDesaName(user?.nama || "Desa PIC");
          setShowDesaModal(false);
        } else {
          // Superadmin: fetch all desa list
          const data = await request("/api/desa");
          setDesaList(Array.isArray(data) ? data : []);
        }
      } catch (err: any) {
        setError("Gagal memuat daftar desa: " + err.message);
      } finally {
        setDesaLoading(false);
      }
    }
    fetchDesa();
  }, []);

  // Keyboard wedge listener for linking RFID card
  useEffect(() => {
    if (step !== 3) return;

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
  }, [step]);

  const handleDesaSelect = (desaId: string) => {
    const desa = desaList.find((d) => d.id === desaId);
    if (desa) {
      setSelectedDesaId(desa.id);
      setSelectedDesaName(desa.nama);
      setShowDesaModal(false);
    }
  };

  const handleDesaModalClose = () => {
    if (!selectedDesaId) {
      window.location.href = "/warga";
      return;
    }
    setShowDesaModal(false);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const f = e.target.files[0];
      setFile(f);
      setPreviewURL(URL.createObjectURL(f));
      setError("");
    }
  };

  const handleUploadAndOCR = async () => {
    if (!file) return;
    setLoading(true);
    setError("");

    try {
      const formData = new FormData();
      formData.append("foto_ktp", file);

      // Call OCR endpoint
      const data = await request("/api/ocr/ktp", {
        method: "POST",
        body: formData,
      });

      // Populate extracted fields
      setWargaData({
        nik: data.nik || "",
        nama: data.nama || "",
        tempat_lahir: data.tempat_lahir || "",
        tanggal_lahir: data.tanggal_lahir || "",
        jenis_kelamin: data.jenis_kelamin === "P" ? "P" : "L",
        alamat: data.alamat || "",
        rt: data.rt || "",
        rw: data.rw || "",
        kelurahan: data.kelurahan || "",
        kecamatan: data.kecamatan || "",
        agama: data.agama || "Islam",
        status_kawin: data.status_kawin || "Belum Kawin",
        pekerjaan: data.pekerjaan || "",
        kewarganegaraan: data.kewarganegaraan || "WNI",
        confidence: data.confidence || 0,
      });

      setStep(2);
    } catch (err: any) {
      setError(err.message || "Gagal melakukan OCR. Silakan coba kembali atau isi data secara manual.");
    } finally {
      setLoading(false);
    }
  };

  const handleSaveWarga = async () => {
    setLoading(true);
    setError("");

    try {
      await request("/api/warga", {
        method: "POST",
        body: JSON.stringify({
          nik: wargaData.nik,
          nama: wargaData.nama,
          tempat_lahir: wargaData.tempat_lahir,
          tanggal_lahir: wargaData.tanggal_lahir,
          jenis_kelamin: wargaData.jenis_kelamin,
          alamat: wargaData.alamat,
          rt: wargaData.rt,
          rw: wargaData.rw,
          kelurahan: wargaData.kelurahan,
          kecamatan: wargaData.kecamatan,
          agama: wargaData.agama,
          status_kawin: wargaData.status_kawin,
          pekerjaan: wargaData.pekerjaan,
          kewarganegaraan: wargaData.kewarganegaraan,
          desa_id: selectedDesaId,
          rfid_uid: rfidUID || undefined,
        }),
      });

      window.location.href = "/warga";
    } catch (err: any) {
      setError(err.message || "Gagal menyimpan data warga.");
      setLoading(false);
    }
  };

  const handleSaveDraft = async () => {
    setLoading(true);
    setError("");

    try {
      const result = await request("/api/warga/draft", {
        method: "POST",
        body: JSON.stringify({
          desa_id: selectedDesaId,
          nik: wargaData.nik || undefined,
          nama: wargaData.nama || undefined,
          foto_ktp_path: previewURL || undefined,
        }),
      });

      const link = `${window.location.origin}${result.url}`;
      setDraftLink(link);
    } catch (err: any) {
      setError(err.message || "Gagal menyimpan draft.");
      setLoading(false);
    }
  };

  const handleCopyLink = async () => {
    if (!draftLink) return;
    try {
      await navigator.clipboard.writeText(draftLink);
      setCopyFeedback("Link berhasil disalin!");
      setTimeout(() => setCopyFeedback(""), 2500);
    } catch {
      setCopyFeedback("Gagal menyalin, silakan copy manual");
    }
  };

  return (
    <div style={{ maxWidth: "800px", margin: "0 auto" }}>
      {/* ====== MODAL PEMILIHAN DESA ====== */}
      {showDesaModal && (
        <div
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: "rgba(0,0,0,0.7)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 9999,
          }}
        >
          <div
            className="glass-card"
            style={{
              maxWidth: "480px",
              width: "90%",
              padding: "32px",
              display: "flex",
              flexDirection: "column",
              gap: "20px",
            }}
          >
            <div style={{ textAlign: "center" }}>
              <div style={{ fontSize: "40px", marginBottom: "8px" }}>🏘️</div>
              <h2 style={{ fontSize: "20px", fontWeight: "700" }}>Pilih Desa</h2>
              <p style={{ color: "var(--text-muted)", fontSize: "14px", marginTop: "4px" }}>
                Tentukan desa tempat warga ini akan didaftarkan.
              </p>
            </div>

            {desaLoading ? (
              <div style={{ textAlign: "center", padding: "24px" }}>
                <div className="spinner" style={{ margin: "0 auto 12px auto" }}></div>
                <span style={{ color: "var(--text-muted)", fontSize: "14px" }}>Memuat daftar desa...</span>
              </div>
            ) : (
              <>
                <div className="form-group">
                  <label className="form-label">Desa / Kelurahan</label>
                  <select
                    className="form-control"
                    value={selectedDesaId}
                    onChange={(e) => setSelectedDesaId(e.target.value)}
                    style={{ fontSize: "15px", padding: "10px 12px" }}
                  >
                    <option value="">-- Pilih Desa --</option>
                    {desaList.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.nama} ({d.kode_desa}){d.kecamatan ? ` - ${d.kecamatan}` : ""}
                      </option>
                    ))}
                  </select>
                </div>

                <div style={{ display: "flex", gap: "12px" }}>
                  <button
                    className="btn btn-secondary"
                    style={{ flex: 1 }}
                    onClick={handleDesaModalClose}
                  >
                    Batal
                  </button>
                  <button
                    className="btn btn-primary"
                    style={{ flex: 1 }}
                    disabled={!selectedDesaId}
                    onClick={() => handleDesaSelect(selectedDesaId)}
                  >
                    Lanjutkan
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* ====== DRAFT LINK SUCCESS BANNER ====== */}
      {draftLink && (
        <div
          className="glass-card"
          style={{
            borderLeft: "4px solid var(--success, #22c55e)",
            marginBottom: "24px",
            display: "flex",
            flexDirection: "column",
            gap: "12px",
          }}
        >
          <div style={{ fontWeight: "700", color: "var(--success, #22c55e)" }}>
            ✅ Draft berhasil disimpan!
          </div>
          <p style={{ fontSize: "14px", color: "var(--text-muted)", margin: 0 }}>
            Bagikan link berikut agar warga bisa melengkapi data nanti:
          </p>
          <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
            <input
              type="text"
              readOnly
              value={draftLink}
              className="form-control"
              style={{ fontFamily: "monospace", fontSize: "13px", flex: 1 }}
              onFocus={(e) => e.target.select()}
            />
            <button className="btn btn-primary" style={{ padding: "8px 16px", whiteSpace: "nowrap" }} onClick={handleCopyLink}>
              {copyFeedback || "Salin Link"}
            </button>
          </div>
          <a href="/warga" className="btn btn-secondary" style={{ textAlign: "center", textDecoration: "none" }}>
            Kembali ke Daftar Warga
          </a>
        </div>
      )}

      <div style={{ marginBottom: "32px" }}>
        <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Registrasi Warga Baru</h1>
        <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
          Pendaftaran warga baru menggunakan teknologi kecerdasan buatan (AI OCR) untuk membaca foto KTP.
          {selectedDesaName && (
            <span style={{ display: "inline-block", marginLeft: "8px", padding: "2px 10px", background: "var(--primary)", borderRadius: "12px", fontSize: "12px", fontWeight: "600", color: "#000" }}>
              🏘️ {selectedDesaName}
            </span>
          )}
        </p>
      </div>

      {error && (
        <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)", marginBottom: "24px" }}>
          ⚠️ {error}
          {step === 1 && (
            <button className="btn btn-secondary" style={{ display: "block", marginTop: "10px", padding: "6px 12px" }} onClick={() => setStep(2)}>
              Lewati & Isi Manual
            </button>
          )}
        </div>
      )}

      {/* Progress Steps Indicators */}
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "40px", position: "relative" }}>
        <div style={{ position: "absolute", top: "50%", left: 0, right: 0, height: "2px", background: "var(--border-color)", zIndex: 0 }}></div>
        <div style={{ zIndex: 1, display: "flex", flexDirection: "column", alignItems: "center" }}>
          <div style={{ width: "36px", height: "36px", borderRadius: "50%", background: step >= 1 ? "var(--primary)" : "var(--border-color)", color: step >= 1 ? "var(--text-dark)" : "var(--text-muted)", display: "flex", alignItems: "center", justifyContent: "center", fontWeight: "700" }}>1</div>
          <span style={{ fontSize: "12px", color: step >= 1 ? "var(--text-main)" : "var(--text-muted)", marginTop: "8px", fontWeight: "600" }}>Upload KTP</span>
        </div>
        <div style={{ zIndex: 1, display: "flex", flexDirection: "column", alignItems: "center" }}>
          <div style={{ width: "36px", height: "36px", borderRadius: "50%", background: step >= 2 ? "var(--primary)" : "var(--border-color)", color: step >= 2 ? "var(--text-dark)" : "var(--text-muted)", display: "flex", alignItems: "center", justifyContent: "center", fontWeight: "700" }}>2</div>
          <span style={{ fontSize: "12px", color: step >= 2 ? "var(--text-main)" : "var(--text-muted)", marginTop: "8px", fontWeight: "600" }}>Verifikasi Data</span>
        </div>
        <div style={{ zIndex: 1, display: "flex", flexDirection: "column", alignItems: "center" }}>
          <div style={{ width: "36px", height: "36px", borderRadius: "50%", background: step >= 3 ? "var(--primary)" : "var(--border-color)", color: step >= 3 ? "var(--text-dark)" : "var(--text-muted)", display: "flex", alignItems: "center", justifyContent: "center", fontWeight: "700" }}>3</div>
          <span style={{ fontSize: "12px", color: step >= 3 ? "var(--text-main)" : "var(--text-muted)", marginTop: "8px", fontWeight: "600" }}>Link RFID Kartu</span>
        </div>
      </div>

      {/* STEP 1: UPLOAD AND OCR */}
      {step === 1 && (
        <div className="glass-card" style={{ display: "flex", flexDirection: "column", alignItems: "center", gap: "24px", padding: "40px" }}>
          <div style={{ width: "100%", maxWidth: "400px", height: "240px", border: "2px dashed var(--border-color)", borderRadius: "var(--radius-md)", display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", cursor: "pointer", position: "relative", overflow: "hidden", background: "hsla(222,47%,7%,0.4)" }}>
            {previewURL ? (
              <img src={previewURL} alt="Preview KTP" style={{ width: "100%", height: "100%", objectFit: "contain" }} />
            ) : (
              <div style={{ textAlign: "center", padding: "20px" }}>
                <span style={{ fontSize: "40px" }}>📷</span>
                <p style={{ marginTop: "12px", fontSize: "15px", fontWeight: "600" }}>Pilih atau Seret Foto KTP</p>
                <p style={{ fontSize: "12px", color: "var(--text-muted)", marginTop: "4px" }}>Mendukung format JPG, PNG (Maksimal 5MB)</p>
              </div>
            )}
            <input type="file" accept="image/*" onChange={handleFileChange} style={{ position: "absolute", top: 0, left: 0, right: 0, bottom: 0, opacity: 0, cursor: "pointer" }} />
          </div>

          <div style={{ display: "flex", gap: "16px", flexWrap: "wrap", justifyContent: "center" }}>
            <button className="btn btn-secondary" onClick={() => setStep(2)}>
              Isi Manual Saja
            </button>
            <button className="btn btn-primary" onClick={handleUploadAndOCR} disabled={!file || loading}>
              {loading ? "Mengekstrak Data AI..." : "Proses OCR Foto KTP"}
            </button>
          </div>
        </div>
      )}

      {/* STEP 2: VERIFICATION FORM */}
      {step === 2 && (
        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "24px" }}>
          <h3 style={{ fontSize: "18px", fontWeight: "700" }}>Verifikasi Data Penduduk</h3>

          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))", gap: "20px" }}>
            <div className="form-group">
              <label className="form-label">NIK (Nomor Induk Kependudukan)</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.nik}
                onChange={(e) => setWargaData({ ...wargaData, nik: e.target.value })}
                maxLength={16}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Nama Lengkap</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.nama}
                onChange={(e) => setWargaData({ ...wargaData, nama: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Tempat Lahir</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.tempat_lahir}
                onChange={(e) => setWargaData({ ...wargaData, tempat_lahir: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Tanggal Lahir (YYYY-MM-DD)</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.tanggal_lahir}
                placeholder="YYYY-MM-DD"
                onChange={(e) => setWargaData({ ...wargaData, tanggal_lahir: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Jenis Kelamin</label>
              <select
                className="form-control"
                value={wargaData.jenis_kelamin}
                onChange={(e) => setWargaData({ ...wargaData, jenis_kelamin: e.target.value })}
              >
                <option value="L">Laki-laki</option>
                <option value="P">Perempuan</option>
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Agama</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.agama}
                onChange={(e) => setWargaData({ ...wargaData, agama: e.target.value })}
              />
            </div>

            <div className="form-group" style={{ gridColumn: "span 2" }}>
              <label className="form-label">Alamat</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.alamat}
                onChange={(e) => setWargaData({ ...wargaData, alamat: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label className="form-label">RT</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.rt}
                onChange={(e) => setWargaData({ ...wargaData, rt: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label className="form-label">RW</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.rw}
                onChange={(e) => setWargaData({ ...wargaData, rw: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Kelurahan / Desa</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.kelurahan}
                onChange={(e) => setWargaData({ ...wargaData, kelurahan: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Kecamatan</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.kecamatan}
                onChange={(e) => setWargaData({ ...wargaData, kecamatan: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Pekerjaan</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.pekerjaan}
                onChange={(e) => setWargaData({ ...wargaData, pekerjaan: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label className="form-label">Status Perkawinan</label>
              <input
                type="text"
                className="form-control"
                value={wargaData.status_kawin}
                onChange={(e) => setWargaData({ ...wargaData, status_kawin: e.target.value })}
              />
            </div>
          </div>

          <div style={{ display: "flex", justifyContent: "space-between", marginTop: "20px" }}>
            <button className="btn btn-secondary" onClick={() => setStep(1)}>
              Kembali
            </button>
            <button className="btn btn-primary" onClick={() => setStep(3)}>
              Lanjutkan ke RFID Link
            </button>
          </div>
        </div>
      )}

      {/* STEP 3: LINK RFID */}
      {step === 3 && (
        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "24px", alignItems: "center", padding: "40px" }}>
          <div style={{ textAlign: "center" }}>
            <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "8px" }}>Tautkan Kartu KTP (NFC/RFID)</h3>
            <p style={{ color: "var(--text-muted)", fontSize: "14px" }}>
              Scan kartu KTP pada alat pembaca RFID USB untuk menautkan UID chip fisik.
            </p>
          </div>

          <div style={{ width: "100%", maxWidth: "400px", padding: "40px 20px", background: "hsla(222,47%,7%,0.6)", border: "2px dashed var(--border-color)", borderRadius: "var(--radius-md)", textAlign: "center" }}>
            {rfidUID ? (
              <div>
                <span className="badge badge-success" style={{ marginBottom: "8px" }}>Kartu Terdeteksi</span>
                <h4 style={{ fontSize: "28px", fontWeight: "800", color: "var(--text-main)" }}>{rfidUID}</h4>
                <button className="btn btn-secondary" style={{ marginTop: "16px", padding: "6px 12px", fontSize: "13px" }} onClick={() => setRfidUID("")}>
                  Scan Ulang
                </button>
              </div>
            ) : (
              <div>
                <div className="spinner" style={{ margin: "0 auto 16px auto" }}></div>
                <span style={{ fontSize: "15px", color: "var(--text-muted)", fontWeight: "500" }}>Menunggu scan kartu KTP...</span>
              </div>
            )}
          </div>

          <div style={{ display: "flex", justifyContent: "space-between", width: "100%", marginTop: "20px", flexWrap: "wrap", gap: "12px" }}>
            <button className="btn btn-secondary" onClick={() => setStep(2)}>
              Kembali
            </button>
            <div style={{ display: "flex", gap: "12px", flexWrap: "wrap" }}>
              <button
                className="btn btn-secondary"
                onClick={handleSaveDraft}
                disabled={loading || !!draftLink}
                style={{ borderColor: "var(--primary)", color: "var(--primary)" }}
              >
                {loading ? "Menyimpan Draft..." : "📎 Simpan Draft & Ambil Link"}
              </button>
              <button className="btn btn-secondary" onClick={handleSaveWarga} disabled={loading}>
                Lewati & Simpan
              </button>
              <button className="btn btn-primary" onClick={handleSaveWarga} disabled={!rfidUID || loading}>
                {loading ? "Menyimpan..." : "Simpan Data & Kartu"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
