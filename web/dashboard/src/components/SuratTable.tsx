import React, { useEffect, useState } from "react";
import { request } from "../lib/api";

interface Surat {
  id: string;
  nomor_surat?: string;
  jenis_surat_kode: string;
  jenis_surat_nama: string;
  nik_pemohon: string;
  nama_pemohon: string;
  status: string;
  created_at: string;
}

export default function SuratTable() {
  const [suratList, setSuratList] = useState<Surat[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function loadSurat() {
      try {
        const data = await request("/api/surat");
        setSuratList(data);
      } catch (err: any) {
        setError(err.message || "Gagal mengambil data arsip surat.");
      } finally {
        setLoading(false);
      }
    }
    loadSurat();
  }, []);

  const filteredList = suratList.filter(
    (s) =>
      s.nama_pemohon.toLowerCase().includes(search.toLowerCase()) ||
      s.nik_pemohon.includes(search) ||
      s.jenis_surat_nama.toLowerCase().includes(search.toLowerCase()) ||
      (s.nomor_surat && s.nomor_surat.toLowerCase().includes(search.toLowerCase()))
  );

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "PRINTED":
      case "SYNCED":
        return <span className="badge badge-success">{status}</span>;
      case "DRAFT":
        return <span className="badge badge-primary">{status}</span>;
      case "FAILED":
        return <span className="badge badge-danger">{status}</span>;
      default:
        return <span className="badge badge-warning">{status}</span>;
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
      <div style={{ marginBottom: "32px" }}>
        <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Arsip Surat</h1>
        <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
          Log cetak dan arsip surat yang diterbitkan secara mandiri oleh warga melalui kiosk.
        </p>
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
          placeholder="🔍 Cari nama pemohon, NIK, jenis surat, atau nomor surat..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {/* Table grid */}
      <div className="table-container">
        <table className="premium-table">
          <thead>
            <tr>
              <th>Nomor Surat</th>
              <th>Nama Pemohon</th>
              <th>NIK Pemohon</th>
              <th>Jenis Surat</th>
              <th>Tanggal Dibuat</th>
              <th>Status Sync</th>
            </tr>
          </thead>
          <tbody>
            {filteredList.map((s) => (
              <tr key={s.id}>
                <td style={{ fontWeight: "600" }}>{s.nomor_surat || "Belum Ada"}</td>
                <td>{s.nama_pemohon}</td>
                <td>{s.nik_pemohon}</td>
                <td>
                  <span className="badge badge-primary" style={{ fontSize: "11px" }}>
                    {s.jenis_surat_nama}
                  </span>
                </td>
                <td>{new Date(s.created_at).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" })}</td>
                <td>{getStatusBadge(s.status)}</td>
              </tr>
            ))}
            {filteredList.length === 0 && (
              <tr>
                <td colSpan={6} style={{ textAlign: "center", color: "var(--text-muted)", padding: "24px" }}>
                  Arsip surat tidak ditemukan
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
