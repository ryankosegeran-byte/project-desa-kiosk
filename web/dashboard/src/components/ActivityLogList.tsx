import React, { useEffect, useState } from "react";
import { request, getUser } from "../lib/api";

interface ActivityLog {
  id: number;
  user_id: string;
  desa_id?: string;
  action: string;
  entity_type?: string;
  entity_id?: string;
  detail?: string;
  ip_address?: string;
  created_at: string;
  // mapped properties
  username?: string;
}

interface Desa {
  id: string;
  nama: string;
}

export default function ActivityLogList() {
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [desas, setDesas] = useState<Desa[]>([]);
  const [selectedDesaId, setSelectedDesaId] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const user = getUser();

  useEffect(() => {
    async function loadData() {
      try {
        if (user?.role === "superadmin") {
          const desaData = await request("/api/desa");
          setDesas(desaData);
        }
        await fetchLogs(selectedDesaId);
      } catch (err: any) {
        setError(err.message || "Gagal mengambil data log aktivitas.");
        setLoading(false);
      }
    }
    loadData();
  }, [selectedDesaId]);

  async function fetchLogs(desaId: string) {
    setLoading(true);
    try {
      let endpoint = "/api/activity-log/desa/my";
      if (user?.role === "superadmin") {
        endpoint = "/api/activity-log";
        if (desaId) {
          endpoint += `?desa_id=${desaId}`;
        }
      }
      const data = await request(endpoint);
      setLogs(data);
    } catch (err: any) {
      setError(err.message || "Gagal memuat log.");
    } finally {
      setLoading(false);
    }
  }

  const getDesaName = (desaId?: string) => {
    if (!desaId) return "Global";
    const d = desas.find((x) => x.id === desaId);
    return d ? d.nama : "Lokal Desa";
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
          <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Log Aktivitas Audit</h1>
          <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
            Audit trail rekam jejak aksi yang dilakukan oleh petugas pelayanan dan sistem.
          </p>
        </div>

        {user?.role === "superadmin" && desas.length > 0 && (
          <div className="form-group" style={{ marginBottom: 0, minWidth: "200px" }}>
            <select
              className="form-control"
              value={selectedDesaId}
              onChange={(e) => setSelectedDesaId(e.target.value)}
            >
              <option value="">Semua Wilayah</option>
              {desas.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.nama}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>

      {error && (
        <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)", marginBottom: "24px" }}>
          ⚠️ {error}
        </div>
      )}

      {/* Logs Table */}
      <div className="table-container">
        <table className="premium-table">
          <thead>
            <tr>
              <th>ID Log</th>
              <th>Waktu Kejadian</th>
              <th>Nama Aksi</th>
              <th>Tipe Entitas</th>
              <th>Wilayah Desa</th>
              <th>IP Address</th>
              <th>Keterangan Detail</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((log) => (
              <tr key={log.id}>
                <td style={{ fontWeight: "600", fontFamily: "monospace" }}>#{log.id}</td>
                <td>{new Date(log.created_at).toLocaleString("id-ID", { dateStyle: "short", timeStyle: "medium" })}</td>
                <td>
                  <span className="badge badge-primary">{log.action}</span>
                </td>
                <td style={{ textTransform: "capitalize" }}>{log.entity_type || "-"}</td>
                <td>{getDesaName(log.desa_id)}</td>
                <td style={{ fontFamily: "monospace" }}>{log.ip_address || "-"}</td>
                <td style={{ fontSize: "13px", color: "var(--text-muted)", maxWidth: "300px", overflow: "hidden", textOverflow: "ellipsis" }}>
                  {log.detail || "-"}
                </td>
              </tr>
            ))}
            {logs.length === 0 && (
              <tr>
                <td colSpan={7} style={{ textAlign: "center", color: "var(--text-muted)", padding: "24px" }}>
                  Belum ada log aktivitas terekam
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
