import React, { useState } from "react";
import { request } from "../lib/api";

export default function LoginForm() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const data = await request("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      });

      localStorage.setItem("token", data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);
      localStorage.setItem("user", JSON.stringify(data.user));

      window.location.href = "/dashboard";
    } catch (err: any) {
      setError(err.message || "Gagal masuk, periksa koneksi Anda.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "100vh", padding: "20px" }}>
      <div className="glass-card" style={{ maxWidth: "400px", width: "100%", padding: "40px" }}>
        <div style={{ textAlign: "center", marginBottom: "32px" }}>
          <div className="brand-logo" style={{ margin: "0 auto 16px auto", width: "60px", height: "60px", fontSize: "30px" }}>KD</div>
          <h2 style={{ fontSize: "24px", fontWeight: "700" }}>MASUK KE HUB DESA</h2>
          <p style={{ color: "var(--text-muted)", fontSize: "14px", marginTop: "8px" }}>
            Kelola data warga, print surat, dan sinkronisasi kiosk
          </p>
        </div>

        {error && (
          <div style={{ background: "hsla(355, 85%, 55%, 0.15)", border: "1px solid var(--danger)", color: "var(--danger)", padding: "12px", borderRadius: "var(--radius-sm)", marginBottom: "20px", fontSize: "14px" }}>
            ⚠️ {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Username</label>
            <input
              type="text"
              className="form-control"
              placeholder="Masukkan username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={loading}
            />
          </div>

          <div className="form-group">
            <label className="form-label">Password</label>
            <input
              type="password"
              className="form-control"
              placeholder="Masukkan password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={loading}
            />
          </div>

          <button type="submit" className="btn btn-primary" style={{ width: "100%", marginTop: "12px" }} disabled={loading}>
            {loading ? "Menghubungkan..." : "Masuk ke Sistem"}
          </button>
        </form>
      </div>
    </div>
  );
}
