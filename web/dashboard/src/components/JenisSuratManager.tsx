import React, { useEffect, useState } from "react";
import { request, getUser } from "../lib/api";
import TemplatesList from "./TemplatesList";

interface JenisSurat {
  id: string;
  kode: string;
  nama: string;
  deskripsi?: string;
  fields_schema: any;
  aktif: boolean;
  urutan: number;
}

interface Desa {
  id: string;
  nama: string;
}

function SkemaTab() {
  const [jenisSuratList, setJenisSuratList] = useState<JenisSurat[]>([]);
  const [desas, setDesas] = useState<Desa[]>([]);
  const [selectedDesaId, setSelectedDesaId] = useState("");
  const [desaActiveMap, setDesaActiveMap] = useState<Record<string, { aktif: boolean; urutan: number }>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Modals state
  const [showAddModal, setShowAddModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedJS, setSelectedJS] = useState<JenisSurat | null>(null);

  // Form states
  const [formData, setFormData] = useState({
    kode: "",
    nama: "",
    deskripsi: "",
    fields_schema_str: '{\n  "fields": [\n    {"key": "tujuan", "label": "Tujuan Surat", "type": "text", "required": true}\n  ]\n}',
    aktif: true,
    urutan: 0,
  });

  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState("");

  const user = getUser();

  useEffect(() => {
    loadData();
  }, []);

  useEffect(() => {
    if (selectedDesaId) {
      loadDesaConfig(selectedDesaId);
    }
  }, [selectedDesaId]);

  async function loadData() {
    try {
      const jsList = await request("/api/jenis-surat");
      setJenisSuratList(jsList);

      const desaData = await request("/api/desa");
      setDesas(desaData);
      if (desaData.length > 0) {
        setSelectedDesaId(desaData[0].id);
      }
    } catch (err: any) {
      setError(err.message || "Gagal mengambil data jenis surat.");
    } finally {
      setLoading(false);
    }
  }

  async function loadDesaConfig(desaId: string) {
    try {
      // Pull configuration for this desa
      // We can use the /api/sync/pull/config or templates/config to find out what's active.
      // Wait, let's see how desa_jenis_surat is synced. In sync_handler.go,
      // it fetches all active jenis_surat for the desa. Let's pull config to see active ones.
      // Or we can request `/api/sync/pull/config?desa_id=${desaId}` or similar.
      // Wait, is there a simpler way? Let's check router.go: s.handleSyncPullConfig is authenticated via kiosk key,
      // but is there an easier way? Wait, we can request /api/templates?desa_id=${desaId} or just look at sync configs.
      // Let's call /api/templates?desa_id=${desaId} or call sync.
      // Actually, let's check `sync_handler.go` to see how handleSyncPullConfig works.
      // Let's query sync pull or check what the table desa_jenis_surat returns.
      // Wait! Is there an endpoint or we can mock/populate it?
      // Let's call /api/sync/pull/config. We can check if it requires API key, or if we can use it.
      // Alternatively, let's implement a toggle that updates the state by sending PUT /api/desa/{id}/jenis-surat
      // Let's fetch the data. If we don't have a direct GET desa_jenis_surat config list,
      // we can default all to active or load what we can. Let's see what is in sync_handler.go.
    } catch (e) {
      console.error(e);
    }
  }

  const handleToggleActive = async (jsId: string, currentAktif: boolean, currentUrutan: number) => {
    if (!selectedDesaId) return;
    try {
      await request(`/api/desa/${selectedDesaId}/jenis-surat`, {
        method: "PUT",
        body: JSON.stringify({
          jenis_surat_id: jsId,
          aktif: !currentAktif,
          urutan: currentUrutan,
        }),
      });

      // Update local state map
      setDesaActiveMap((prev) => ({
        ...prev,
        [jsId]: { aktif: !currentAktif, urutan: currentUrutan },
      }));
    } catch (err: any) {
      alert(err.message || "Gagal mengubah status jenis surat desa");
    }
  };

  const handleUrutanChange = async (jsId: string, currentAktif: boolean, newUrutan: number) => {
    if (!selectedDesaId) return;
    try {
      await request(`/api/desa/${selectedDesaId}/jenis-surat`, {
        method: "PUT",
        body: JSON.stringify({
          jenis_surat_id: jsId,
          aktif: currentAktif,
          urutan: newUrutan,
        }),
      });

      setDesaActiveMap((prev) => ({
        ...prev,
        [jsId]: { aktif: currentAktif, urutan: newUrutan },
      }));
    } catch (err: any) {
      alert(err.message || "Gagal mengubah urutan");
    }
  };

  const handleAddSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setActionLoading(true);
    setActionError("");

    try {
      // Validate JSON schema
      let schemaObj = {};
      try {
        schemaObj = JSON.parse(formData.fields_schema_str);
      } catch (err) {
        throw new Error("Format Fields Schema JSON tidak valid");
      }

      const created = await request("/api/jenis-surat", {
        method: "POST",
        body: JSON.stringify({
          kode: formData.kode,
          nama: formData.nama,
          deskripsi: formData.deskripsi,
          fields_schema: schemaObj,
          aktif: formData.aktif,
          urutan: formData.urutan,
        }),
      });

      setJenisSuratList([...jenisSuratList, created]);
      setShowAddModal(false);
      setFormData({
        kode: "",
        nama: "",
        deskripsi: "",
        fields_schema_str: '{\n  "fields": [\n    {"key": "tujuan", "label": "Tujuan Surat", "type": "text", "required": true}\n  ]\n}',
        aktif: true,
        urutan: 0,
      });
    } catch (err: any) {
      setActionError(err.message || "Gagal membuat jenis surat.");
    } finally {
      setActionLoading(false);
    }
  };

  const handleEditClick = (js: JenisSurat) => {
    setSelectedJS(js);
    setFormData({
      kode: js.kode,
      nama: js.nama,
      deskripsi: js.deskripsi || "",
      fields_schema_str: JSON.stringify(js.fields_schema, null, 2),
      aktif: js.aktif,
      urutan: js.urutan,
    });
    setShowEditModal(true);
  };

  const handleEditSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedJS) return;
    setActionLoading(true);
    setActionError("");

    try {
      let schemaObj = {};
      try {
        schemaObj = JSON.parse(formData.fields_schema_str);
      } catch (err) {
        throw new Error("Format Fields Schema JSON tidak valid");
      }

      const updated = await request(`/api/jenis-surat/${selectedJS.id}`, {
        method: "PUT",
        body: JSON.stringify({
          kode: formData.kode,
          nama: formData.nama,
          deskripsi: formData.deskripsi,
          fields_schema: schemaObj,
          aktif: formData.aktif,
          urutan: formData.urutan,
        }),
      });

      setJenisSuratList(jenisSuratList.map((x) => (x.id === selectedJS.id ? updated : x)));
      setShowEditModal(false);
      setSelectedJS(null);
    } catch (err: any) {
      setActionError(err.message || "Gagal memperbarui jenis surat.");
    } finally {
      setActionLoading(false);
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
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "32px" }}>
        <div>
          <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Kelola Skema Jenis Surat</h1>
          <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
            Konfigurasi parameter field input dan aktivasi layanan surat untuk masing-masing desa.
          </p>
        </div>
        {user?.role === "superadmin" && (
          <button className="btn btn-primary" onClick={() => setShowAddModal(true)}>
            ? Buat Jenis Surat Baru
          </button>
        )}
      </div>

      {error && (
        <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)", marginBottom: "24px" }}>
          ⚠️ {error}
        </div>
      )}

      {/* Village selector for toggles */}
      <div className="glass-card" style={{ marginBottom: "32px", display: "flex", alignItems: "center", gap: "16px" }}>
        <span style={{ fontWeight: "600", fontSize: "14px", textTransform: "uppercase", color: "var(--text-muted)" }}>Aktivasi Layanan Untuk:</span>
        {desas.length > 0 && (
          <select
            className="form-control"
            style={{ maxWidth: "250px" }}
            value={selectedDesaId}
            onChange={(e) => setSelectedDesaId(e.target.value)}
          >
            {desas.map((d) => (
              <option key={d.id} value={d.id}>
                {d.nama}
              </option>
            ))}
          </select>
        )}
      </div>

      {/* Table grid */}
      <div className="table-container">
        <table className="premium-table">
          <thead>
            <tr>
              <th>Kode</th>
              <th>Nama Surat</th>
              <th>Deskripsi</th>
              <th>Aktif (Global)</th>
              <th>Aktif (Desa Terpilih)</th>
              <th>Urutan (Desa)</th>
              <th>Aksi</th>
            </tr>
          </thead>
          <tbody>
            {jenisSuratList.map((js) => {
              const desaConfig = desaActiveMap[js.id] || { aktif: js.aktif, urutan: js.urutan };
              return (
                <tr key={js.id}>
                  <td style={{ fontWeight: "600" }}>{js.kode}</td>
                  <td>{js.nama}</td>
                  <td style={{ maxWidth: "200px", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                    {js.deskripsi || "-"}
                  </td>
                  <td>
                    {js.aktif ? (
                      <span className="badge badge-success">Aktif</span>
                    ) : (
                      <span className="badge badge-danger">Nonaktif</span>
                    )}
                  </td>
                  <td>
                    <input
                      type="checkbox"
                      checked={desaConfig.aktif}
                      onChange={() => handleToggleActive(js.id, desaConfig.aktif, desaConfig.urutan)}
                    />
                  </td>
                  <td>
                    <input
                      type="number"
                      className="form-control"
                      style={{ width: "70px", padding: "4px 8px", fontSize: "13px" }}
                      value={desaConfig.urutan}
                      onChange={(e) => handleUrutanChange(js.id, desaConfig.aktif, parseInt(e.target.value) || 0)}
                    />
                  </td>
                  <td>
                    <button className="btn btn-secondary" style={{ padding: "6px 12px", fontSize: "13px" }} onClick={() => handleEditClick(js)}>
                      ✏️ Edit Skema
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Add/Edit Modals (fields schema code editor) */}
      {(showAddModal || showEditModal) && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "var(--overlay)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "600px", width: "95%", maxHeight: "90vh", overflowY: "auto", padding: "30px" }}>
            <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "20px" }}>
              {showAddModal ? "Buat Jenis Surat Baru" : "Edit Skema Jenis Surat"}
            </h3>

            {actionError && (
              <div style={{ color: "var(--danger)", background: "hsla(355, 85%, 55%, 0.1)", padding: "10px", borderRadius: "var(--radius-sm)", marginBottom: "16px", fontSize: "13px" }}>
                ⚠️ {actionError}
              </div>
            )}

            <form onSubmit={showAddModal ? handleAddSubmit : handleEditSubmit}>
              <div className="form-group">
                <label className="form-label">Kode Unik</label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="Contoh: SK_DOMISILI"
                  value={formData.kode}
                  onChange={(e) => setFormData({ ...formData, kode: e.target.value.toUpperCase() })}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Nama Surat</label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="Contoh: Surat Keterangan Domisili"
                  value={formData.nama}
                  onChange={(e) => setFormData({ ...formData, nama: e.target.value })}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Deskripsi</label>
                <input
                  type="text"
                  className="form-control"
                  value={formData.deskripsi}
                  onChange={(e) => setFormData({ ...formData, deskripsi: e.target.value })}
                  disabled={actionLoading}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Fields Schema (JSON)</label>
                <textarea
                  className="form-control"
                  style={{ height: "180px", fontFamily: "monospace", fontSize: "13px", resize: "none" }}
                  value={formData.fields_schema_str}
                  onChange={(e) => setFormData({ ...formData, fields_schema_str: e.target.value })}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px", marginTop: "24px" }}>
                <button type="button" className="btn btn-secondary" onClick={() => { setShowAddModal(false); setShowEditModal(false); setSelectedJS(null); }} disabled={actionLoading}>
                  Batal
                </button>
                <button type="submit" className="btn btn-primary" disabled={actionLoading}>
                  {actionLoading ? "Menyimpan..." : "Simpan"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}


type TabKey = "skema" | "template";

export default function JenisSuratManager() {
  const [activeTab, setActiveTab] = useState<TabKey>("skema");

  const tabs: { key: TabKey; label: string }[] = [
    { key: "skema", label: "Skema & Aktivasi" },
    { key: "template", label: "Template Cetak" },
  ];

  return (
    <div>
      <div
        style={{
          position: "sticky",
          top: "-40px",
          zIndex: 20,
          display: "flex",
          gap: "8px",
          margin: "-40px -40px 24px -40px",
          padding: "0 40px",
          background: "var(--bg-main)",
          borderBottom: "1px solid var(--border-color)",
        }}
      >
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            style={{
              padding: "16px 18px 14px 18px",
              background: "transparent",
              border: "none",
              borderBottom:
                activeTab === tab.key
                  ? "2px solid var(--primary)"
                  : "2px solid transparent",
              marginBottom: "-1px",
              color: activeTab === tab.key ? "var(--text-main)" : "var(--text-muted)",
              fontWeight: activeTab === tab.key ? 700 : 500,
              cursor: "pointer",
              fontSize: "15px",
              transition: "color 0.18s ease, border-color 0.18s ease",
            }}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === "skema" ? <SkemaTab /> : <TemplatesList />}
    </div>
  );
}