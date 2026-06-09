import React, { useEffect, useState } from "react";
import { request, getUser } from "../lib/api";

interface Stats {
  total_warga: number;
  total_surat: number;
  total_kiosks: number;
  active_jenis_surat: number;
}

export default function DashboardOverview() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const user = getUser();

  useEffect(() => {
    async function loadStats() {
      try {
        const data = await request("/api/dashboard/stats");
        setStats(data);
      } catch (err: any) {
        setError(err.message || "Gagal mengambil data statistik.");
      } finally {
        setLoading(false);
      }
    }
    loadStats();
  }, []);

  if (loading) {
    return (
      <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "400px" }}>
        <div className="spinner"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)" }}>
        ⚠️ {error}
      </div>
    );
  }

  return (
    <div>
      <div style={{ marginBottom: "40px" }}>
        <h1 style={{ fontSize: "32px", fontWeight: "700" }}>Halo, {user?.nama || "User"}!</h1>
        <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
          Selamat datang di panel administrasi Kiosk Desa. Berikut rangkuman status sistem hari ini.
        </p>
      </div>

      {/* Stats Grid */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))", gap: "24px", marginBottom: "40px" }}>
        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
          <span style={{ fontSize: "14px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase" }}>Warga Terdaftar</span>
          <span style={{ fontSize: "40px", fontWeight: "800", color: "var(--primary)" }}>{stats?.total_warga || 0}</span>
          <span style={{ fontSize: "12px", color: "var(--text-muted)" }}>Total profil di database</span>
        </div>

        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
          <span style={{ fontSize: "14px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase" }}>Surat Dicetak</span>
          <span style={{ fontSize: "40px", fontWeight: "800", color: "var(--secondary)" }}>{stats?.total_surat || 0}</span>
          <span style={{ fontSize: "12px", color: "var(--text-muted)" }}>Total pengajuan cetak kiosk</span>
        </div>

        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
          <span style={{ fontSize: "14px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase" }}>Kiosk Aktif</span>
          <span style={{ fontSize: "40px", fontWeight: "800", color: "var(--success)" }}>{stats?.total_kiosks || 0}</span>
          <span style={{ fontSize: "12px", color: "var(--text-muted)" }}>Unit terminal terdaftar</span>
        </div>

        <div className="glass-card" style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
          <span style={{ fontSize: "14px", fontWeight: "600", color: "var(--text-muted)", textTransform: "uppercase" }}>Jenis Layanan</span>
          <span style={{ fontSize: "40px", fontWeight: "800", color: "var(--warning)" }}>{stats?.active_jenis_surat || 0}</span>
          <span style={{ fontSize: "12px", color: "var(--text-muted)" }}>Layanan surat yang aktif</span>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="glass-card" style={{ marginBottom: "40px" }}>
        <h3 style={{ fontSize: "18px", fontWeight: "700", marginBottom: "20px" }}>Akses Cepat Layanan</h3>
        <div style={{ display: "flex", flexWrap: "wrap", gap: "16px" }}>
          <a href="/warga/register" className="btn btn-primary">
            ➕ Registrasi Warga Baru
          </a>
          <a href="/warga" className="btn btn-secondary">
            👥 Cari Profil Warga
          </a>
          <a href="/surat" className="btn btn-secondary">
            📄 Lihat Arsip Cetak
          </a>
          {user?.role === "superadmin" && (
            <a href="/admin/users" className="btn btn-secondary">
              👤 Kelola Pengguna PIC
            </a>
          )}
        </div>
      </div>
    </div>
  );
}
