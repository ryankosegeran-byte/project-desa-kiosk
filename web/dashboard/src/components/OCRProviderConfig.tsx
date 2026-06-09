import React, { useState } from "react";
import { request } from "../lib/api";

export default function OCRProviderConfig() {
  const [file, setFile] = useState<File | null>(null);
  const [previewURL, setPreviewURL] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState("");

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const f = e.target.files[0];
      setFile(f);
      setPreviewURL(URL.createObjectURL(f));
      setResult(null);
      setError("");
    }
  };

  const handleTestOCR = async () => {
    if (!file) return;
    setLoading(true);
    setError("");
    setResult(null);

    try {
      const formData = new FormData();
      formData.append("foto_ktp", file);

      const data = await request("/api/ocr/ktp", {
        method: "POST",
        body: formData,
      });

      setResult(data);
    } catch (err: any) {
      setError(err.message || "Gagal melakukan pengetesan OCR.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div style={{ marginBottom: "32px" }}>
        <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Pengaturan & Tes AI OCR</h1>
        <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
          Monitor status engine AI OCR KTP dan lakukan uji coba ekstraksi data teks dari foto KTP.
        </p>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(320px, 1fr))", gap: "30px" }}>
        {/* Provider Status Details */}
        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "20px" }}>
          <h3 style={{ fontSize: "18px", fontWeight: "700" }}>Status Engine OCR</h3>
          
          <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", borderBottom: "1px solid var(--border-color)", paddingBottom: "10px" }}>
              <span>Failover Strategy</span>
              <span className="badge badge-success">AKTIF</span>
            </div>

            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", borderBottom: "1px solid var(--border-color)", paddingBottom: "10px" }}>
              <span>Gemini AI Provider</span>
              <span className="badge badge-primary">CONFIGURED</span>
            </div>

            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", borderBottom: "1px solid var(--border-color)", paddingBottom: "10px" }}>
              <span>Mistral AI Provider</span>
              <span className="badge badge-primary">CONFIGURED</span>
            </div>

            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", borderBottom: "1px solid var(--border-color)", paddingBottom: "10px" }}>
              <span>Groq Cloud Provider</span>
              <span className="badge badge-primary">CONFIGURED</span>
            </div>

            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", paddingBottom: "10px" }}>
              <span>Offline Mock Fallback</span>
              <span className="badge badge-warning">READY</span>
            </div>
          </div>

          <div style={{ background: "hsla(210, 100%, 55%, 0.05)", border: "1px solid var(--border-glow)", padding: "16px", borderRadius: "var(--radius-sm)", fontSize: "13px", color: "var(--text-muted)" }}>
            💡 <strong>Info:</strong> API Keys dan model AI dikonfigurasi melalui file environment variable (<code>.env</code>) pada server untuk menjamin keamanan credential.
          </div>
        </div>

        {/* OCR Test Playground */}
        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "20px" }}>
          <h3 style={{ fontSize: "18px", fontWeight: "700" }}>Playground Tes OCR</h3>

          <div style={{ width: "100%", height: "180px", border: "2px dashed var(--border-color)", borderRadius: "var(--radius-sm)", display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", cursor: "pointer", position: "relative", overflow: "hidden", background: "hsla(222,47%,7%,0.4)" }}>
            {previewURL ? (
              <img src={previewURL} alt="Tes Preview" style={{ width: "100%", height: "100%", objectFit: "contain" }} />
            ) : (
              <div style={{ textAlign: "center", padding: "10px" }}>
                <span style={{ fontSize: "32px" }}>📷</span>
                <p style={{ marginTop: "8px", fontSize: "14px", fontWeight: "600" }}>Pilih Foto KTP Uji Coba</p>
              </div>
            )}
            <input type="file" accept="image/*" onChange={handleFileChange} style={{ position: "absolute", top: 0, left: 0, right: 0, bottom: 0, opacity: 0, cursor: "pointer" }} />
          </div>

          <button className="btn btn-primary" onClick={handleTestOCR} disabled={!file || loading}>
            {loading ? "Mengekstrak dengan AI..." : "Uji Coba Kirim Foto"}
          </button>

          {error && (
            <div style={{ color: "var(--danger)", background: "hsla(355, 85%, 55%, 0.1)", padding: "10px", borderRadius: "var(--radius-sm)", fontSize: "13px" }}>
              ⚠️ {error}
            </div>
          )}

          {result && (
            <div style={{ background: "hsla(222,47%,7%,0.8)", border: "1px solid var(--border-color)", borderRadius: "var(--radius-sm)", padding: "16px", height: "200px", overflowY: "auto" }}>
              <span style={{ fontSize: "11px", fontWeight: "600", color: "var(--text-muted)" }}>HASIL EKSTRAKSI AI</span>
              <pre style={{ fontFamily: "monospace", fontSize: "12px", marginTop: "8px", color: "var(--success)" }}>
                {JSON.stringify(result, null, 2)}
              </pre>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
