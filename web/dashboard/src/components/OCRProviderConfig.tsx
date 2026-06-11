import React, { useState, useEffect } from "react";
import { request } from "../lib/api";

interface ProviderInfo {
  name: string;
  configured: boolean;
  model?: string;
}

interface OCRStatus {
  strategy: string;
  providers: ProviderInfo[];
}

interface TestResult {
  provider: string;
  data?: Record<string, unknown>;
  error?: string;
}

const PROVIDER_LABELS: Record<string, string> = {
  gemini: "Gemini AI",
  mistral: "Mistral AI",
  groq: "Groq Cloud",
  mock: "Offline Mock",
};

export default function OCRProviderConfig() {
  const [status, setStatus] = useState<OCRStatus | null>(null);
  const [statusLoading, setStatusLoading] = useState(true);

  const [file, setFile] = useState<File | null>(null);
  const [previewURL, setPreviewURL] = useState("");
  const [loading, setLoading] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState("auto");
  const [result, setResult] = useState<TestResult | null>(null);
  const [error, setError] = useState("");

  // Fetch OCR status on mount
  useEffect(() => {
    (async () => {
      try {
        const data = await request("/api/ocr/status");
        setStatus(data);
      } catch {
        // silently ignore — UI will show loading state
      } finally {
        setStatusLoading(false);
      }
    })();
  }, []);

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
      if (selectedProvider && selectedProvider !== "auto") {
        formData.append("provider", selectedProvider);
      }

      const data = await request("/api/ocr/test", {
        method: "POST",
        body: formData,
      });

      setResult(data);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Gagal melakukan pengetesan OCR.";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  const configuredCount = status?.providers.filter((p) => p.configured).length ?? 0;

  return (
    <div>
      <div style={{ marginBottom: "32px" }}>
        <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Pengaturan & Tes AI OCR</h1>
        <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
          Monitor status engine AI OCR KTP dan lakukan uji coba ekstraksi data teks dari foto KTP.
        </p>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(320px, 1fr))", gap: "30px" }}>
        {/* Provider Status Panel */}
        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "20px" }}>
          <h3 style={{ fontSize: "18px", fontWeight: "700" }}>Status Engine OCR</h3>

          {statusLoading ? (
            <p style={{ color: "var(--text-muted)", fontSize: "13px" }}>Memuat status...</p>
          ) : (
            <>
              <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
                {/* Strategy row */}
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", borderBottom: "1px solid var(--border-color)", paddingBottom: "10px" }}>
                  <span>Failover Strategy</span>
                  <span className="badge badge-success">
                    {status?.strategy === "round_robin" ? "ROUND ROBIN" : "FAILOVER"}
                  </span>
                </div>

                {/* Provider rows */}
                {status?.providers.map((p) => (
                  <div
                    key={p.name}
                    style={{ display: "flex", justifyContent: "space-between", alignItems: "center", borderBottom: "1px solid var(--border-color)", paddingBottom: "10px" }}
                  >
                    <div>
                      <span>{PROVIDER_LABELS[p.name] ?? p.name}</span>
                      {p.model && p.model !== "offline-fallback" && (
                        <span style={{ display: "block", fontSize: "11px", color: "var(--text-muted)" }}>
                          {p.model}
                        </span>
                      )}
                    </div>
                    <span className={`badge ${p.configured ? "badge-primary" : "badge-secondary"}`}>
                      {p.configured ? "CONFIGURED" : "NOT CONFIGURED"}
                    </span>
                  </div>
                ))}
              </div>

              <div style={{ background: "hsla(210, 100%, 55%, 0.05)", border: "1px solid var(--border-glow)", padding: "16px", borderRadius: "var(--radius-sm)", fontSize: "13px", color: "var(--text-muted)" }}>
                💡 <strong>Info:</strong> API Keys dan model AI dikonfigurasi melalui file environment variable (<code>.env</code>) pada server untuk menjamin keamanan credential.
                {configuredCount === 0 && (
                  <span style={{ display: "block", marginTop: "8px", color: "var(--warning)" }}>
                    ⚠️ Belum ada API key yang diatur. OCR akan menggunakan <strong>Mock Fallback</strong> saja.
                  </span>
                )}
              </div>
            </>
          )}
        </div>

        {/* OCR Test Playground */}
        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "20px" }}>
          <h3 style={{ fontSize: "18px", fontWeight: "700" }}>Playground Tes OCR</h3>

          {/* Provider selector */}
          <div>
            <label style={{ display: "block", fontSize: "13px", fontWeight: "600", marginBottom: "6px" }}>
              Provider
            </label>
            <select
              value={selectedProvider}
              onChange={(e) => setSelectedProvider(e.target.value)}
              style={{
                width: "100%",
                padding: "8px 12px",
                borderRadius: "var(--radius-sm)",
                border: "1px solid var(--border-color)",
                background: "var(--bg-card)",
                color: "var(--text-primary)",
                fontSize: "13px",
              }}
            >
              <option value="auto">Auto Failover (semua provider)</option>
              {status?.providers.map((p) => (
                <option key={p.name} value={p.name} disabled={!p.configured}>
                  {PROVIDER_LABELS[p.name] ?? p.name}
                  {!p.configured ? " (tidak aktif)" : ""}
                </option>
              ))}
            </select>
          </div>

          {/* File drop zone */}
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
            <div style={{ background: "hsla(222,47%,7%,0.8)", border: "1px solid var(--border-color)", borderRadius: "var(--radius-sm)", padding: "16px", maxHeight: "280px", overflowY: "auto" }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "8px" }}>
                <span style={{ fontSize: "11px", fontWeight: "600", color: "var(--text-muted)" }}>HASIL EKSTRAKSI AI</span>
                <span className={`badge ${result.error ? "badge-danger" : "badge-success"}`}>
                  {result.error ? "GAGAL" : "BERHASIL"}
                  {" · "}
                  {PROVIDER_LABELS[result.provider] ?? result.provider}
                </span>
              </div>
              {result.error ? (
                <p style={{ color: "var(--danger)", fontFamily: "monospace", fontSize: "12px" }}>{result.error}</p>
              ) : (
                <pre style={{ fontFamily: "monospace", fontSize: "12px", color: "var(--success)", whiteSpace: "pre-wrap" }}>
                  {JSON.stringify(result.data, null, 2)}
                </pre>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
