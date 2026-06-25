import React, { useEffect, useState } from "react";
import { request } from "../lib/api";
import { useRFIDScanner } from "../hooks/useRFIDScanner";

interface WargaLite {
  id: string;
  nik: string;
  nama: string;
  desa_id?: string;
}

interface Props {
  warga: WargaLite;
  /** Fallback desa id (e.g. superadmin desa filter) when warga has none. */
  desaIdHint?: string;
  onClose: () => void;
  onLinked: () => void;
}

/**
 * LinkRFIDModal links an RFID card to an existing warga. While open it behaves
 * exactly like the registration step-3 scan: it listens to keyboard-wedge, the
 * local PC/SC bridge agent, AND "Scan via Kiosk" (server relay). It also opens a
 * kiosk registration session for the warga''s desa so a scan at the kiosk is
 * routed here.
 *
 * Crucially, all RFID side-effects live INSIDE this component, so they only run
 * while the modal is mounted (open). The parent list page is never affected.
 */
export default function LinkRFIDModal({ warga, desaIdHint, onClose, onLinked }: Props) {
  const [rfidUID, setRfidUID] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const desaId = warga.desa_id || desaIdHint || "";

  // Multi-source RFID input (active only while this modal is mounted).
  const { kioskRelayStatus } = useRFIDScanner(true, (uid) => setRfidUID(uid), desaId);

  // Open a kiosk registration session for this desa while the modal is open.
  useEffect(() => {
    if (!desaId) return;
    const q = `?desa_id=${encodeURIComponent(desaId)}`;
    const nama = (warga.nama || "").trim();
    const startQ = nama ? `${q}&name=${encodeURIComponent(nama)}` : q;
    request(`/api/rfid/session/start${startQ}`, { method: "POST" }).catch(() => {});
    return () => {
      request(`/api/rfid/session/stop${q}`, { method: "POST" }).catch(() => {});
    };
  }, [desaId, warga.nama]);

  const handleSave = async () => {
    if (!rfidUID) return;
    setLoading(true);
    setError("");
    try {
      await request(`/api/warga/${warga.id}/rfid`, {
        method: "PUT",
        body: JSON.stringify({ rfid_uid: rfidUID }),
      });
      onLinked();
    } catch (err: any) {
      setError(err.message || "Gagal menautkan kartu RFID.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "var(--overlay)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
      <div className="glass-card" style={{ maxWidth: "450px", width: "100%", padding: "30px" }}>
        <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "8px" }}>Tautkan Kartu KTP (RFID)</h3>
        <p style={{ color: "var(--text-muted)", fontSize: "14px", marginBottom: "12px" }}>
          Tautkan ID kartu RFID baru ke data warga: <strong>{warga.nama}</strong> ({warga.nik})
        </p>

        <div style={{ display: "inline-flex", alignItems: "center", gap: "6px", fontSize: "12px", fontWeight: 600, padding: "4px 10px", borderRadius: "999px", marginBottom: "16px", background: kioskRelayStatus === "online" ? "rgba(22,163,74,0.12)" : "var(--bg-inset)", color: kioskRelayStatus === "online" ? "#16a34a" : "var(--text-muted)" }}>
          <span style={{ width: "8px", height: "8px", borderRadius: "50%", background: "currentColor" }} />
          {kioskRelayStatus === "online" ? "Scan via Kiosk siap" : kioskRelayStatus === "connecting" ? "Menyambung ke kiosk..." : "Kiosk/Server offline"}
        </div>

        {error && (
          <div style={{ color: "var(--danger)", background: "hsla(355, 85%, 55%, 0.1)", padding: "10px", borderRadius: "var(--radius-sm)", marginBottom: "16px", fontSize: "13px" }}>
            {error}
          </div>
        )}

        <div style={{ textAlign: "center", padding: "30px 20px", background: "var(--bg-inset)", border: "1px dashed var(--border-color)", borderRadius: "var(--radius-sm)", marginBottom: "20px" }}>
          {rfidUID ? (
            <div>
              <span style={{ fontSize: "12px", color: "var(--success)", fontWeight: "600", textTransform: "uppercase" }}>Kartu Terbaca</span>
              <div style={{ fontSize: "24px", fontWeight: "800", color: "var(--text-main)", marginTop: "4px" }}>{rfidUID}</div>
              <button className="btn btn-secondary" style={{ marginTop: "12px", padding: "4px 10px", fontSize: "12px" }} onClick={() => setRfidUID("")}>Scan Ulang</button>
            </div>
          ) : (
            <div>
              <div className="spinner" style={{ margin: "0 auto 12px auto" }}></div>
              <span style={{ fontSize: "14px", color: "var(--text-muted)" }}>Tempelkan KTP pada reader (USB) atau scan di kiosk desa...</span>
            </div>
          )}
        </div>

        <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px" }}>
          <button className="btn btn-secondary" onClick={onClose} disabled={loading}>Batal</button>
          <button className="btn btn-primary" onClick={handleSave} disabled={!rfidUID || loading}>
            {loading ? "Menyimpan..." : "Simpan Tautan"}
          </button>
        </div>
      </div>
    </div>
  );
}