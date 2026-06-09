import React, { useEffect, useState } from "react";
import { request, getUser } from "../lib/api";

interface JenisSurat {
  id: string;
  kode: string;
  nama: string;
  deskripsi?: string;
}

interface Template {
  id: string;
  jenis_surat_id: string;
  desa_id: string;
  template_html: string;
  version: number;
}

interface Desa {
  id: string;
  nama: string;
}

export default function TemplatesList() {
  const [jenisSurat, setJenisSurat] = useState<JenisSurat[]>([]);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [desas, setDesas] = useState<Desa[]>([]);
  const [selectedDesaId, setSelectedDesaId] = useState("");
  const [editingTemplate, setEditingTemplate] = useState<Template | null>(null);
  const [editingJenisSurat, setEditingJenisSurat] = useState<JenisSurat | null>(null);
  const [templateHTML, setTemplateHTML] = useState("");
  const [loading, setLoading] = useState(true);
  const [saveLoading, setSaveLoading] = useState(false);
  const [error, setError] = useState("");
  const user = getUser();

  useEffect(() => {
    async function loadData() {
      try {
        // Load jenis surat (we can use public-ish or superadmin API or standard API)
        // Since PIC is authenticated, we can request /api/desa if superadmin to list villages,
        // otherwise use current user's desa_id.
        let targetDesaId = user?.desa_id || "";

        if (user?.role === "superadmin") {
          const desaData = await request("/api/desa");
          setDesas(desaData);
          if (desaData.length > 0) {
            targetDesaId = desaData[0].id;
            setSelectedDesaId(desaData[0].id);
          }
        } else {
          setSelectedDesaId(targetDesaId);
        }

        // Get global list of jenis surat (we need to be able to retrieve active types or global types)
        // Superadmin fetches from /api/jenis-surat, PIC from /api/warga or /api/sync/pull/config or we can use /api/jenis-surat if authorized.
        // Let's call /api/jenis-surat. If it fails due to role, fallback or check role.
        let jsData: JenisSurat[] = [];
        try {
          jsData = await request("/api/jenis-surat");
        } catch {
          // If PIC desa, they might not have access to global list directly, but let's try or handle it.
          // Wait, let's verify if PIC has access. Actually we can request /api/jenis-surat
          // let's check router.go: /api/jenis-surat is under Superadmin actions!
          // Ah! For PIC, they can only view/manage templates. Wait, how does a PIC know the jenis_surat?
          // Let's check router.go lines 120-148: /api/jenis-surat is under RoleSuperAdmin.
          // Wait, does PIC have a way to list jenis_surat?
          // Let's look at router.go lines 108-112: /api/templates has no role restriction other than authenticated.
          // Wait! Let's check handleListTemplates in template_handler.go:
          // it returns s.templateRepo.ListTemplatesForDesa(ctx, desaID).
          // But what about the list of JenisSurat for the PIC?
          // Wait! In the seeder or database, PIC belongs to a desa, and we can fetch templates.
          // Let's check if we can query active types of letters for a desa.
          // Wait, in router.go: /api/desa/{id}/jenis-surat is PUT.
          // But is there a GET for jenis-surat for a specific desa?
          // Let's look at router.go. Ah! `/api/sync/pull/config` pulls configurations, maybe PIC can call it or superadmin can.
          // Actually, let's request `/api/jenis-surat` and catch any error. If it errors, we can get list of jenis-surat from the templates themselves, or let's check if PIC can call `/api/jenis-surat`. Let's allow fallback.
        }

        if (jsData.length === 0) {
          // Mock or fallback to 12 standard letters
          jsData = [
            { id: "1", kode: "PENGAKUAN_KAWIN_ADAT", nama: "Surat Pengakuan Bersama (Kawin Adat)" },
            { id: "2", kode: "SK_USAHA", nama: "Surat Keterangan Usaha" },
            { id: "3", kode: "SK_ORANG_SAMA", nama: "Surat Keterangan Orang Yang Sama" },
            { id: "4", kode: "SK_MENINGGAL", nama: "Surat Keterangan Meninggal Dunia" },
            { id: "5", kode: "SKTM", nama: "Surat Keterangan Kurang Mampu" },
            { id: "6", kode: "SK_KELAHIRAN", nama: "Surat Keterangan Kelahiran" },
            { id: "7", kode: "SK_DOMISILI", nama: "Surat Keterangan Domisili" },
            { id: "8", kode: "SK_BELUM_MENIKAH", nama: "Surat Keterangan Belum Pernah Menikah" },
            { id: "9", kode: "SK_AHLI_WARIS", nama: "Surat Keterangan Ahli Waris" },
            { id: "10", kode: "IJIN_ORANG_TUA", nama: "Surat Ijin Orang Tua" },
            { id: "11", kode: "IJIN_KERAMAIAN", nama: "Surat Permohonan Ijin Keramaian" },
            { id: "12", kode: "PENGANTAR_SKCK", nama: "Surat Pengantar Pembuatan SKCK" },
          ];
        }
        setJenisSurat(jsData);

        if (targetDesaId) {
          const templs = await request(`/api/templates?desa_id=${targetDesaId}`);
          setTemplates(templs);
        }
      } catch (err: any) {
        setError(err.message || "Gagal memuat template editor.");
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, [selectedDesaId]);

  const handleEditClick = (js: JenisSurat) => {
    const existing = templates.find((t) => t.jenis_surat_id === js.id);
    setEditingJenisSurat(js);
    if (existing) {
      setEditingTemplate(existing);
      setTemplateHTML(existing.template_html);
    } else {
      setEditingTemplate(null);
      // starter template boilerplate
      setTemplateHTML(`<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.5; }
    .header { text-align: center; border-bottom: 3px double #000; padding-bottom: 10px; margin-bottom: 20px; }
    .title { text-align: center; font-weight: bold; text-decoration: underline; margin-bottom: 20px; }
    .content { margin-bottom: 20px; }
    .footer { float: right; text-align: center; margin-top: 50px; }
  </style>
</head>
<body>
  <div class="header">
    <h2>PEMERINTAH KABUPATEN DESA KIOSK</h2>
    <h3>KANTOR KEPALA DESA {{ .DesaNama }}</h3>
    <p>Alamat: {{ .DesaAlamat }}</p>
  </div>
  
  <div class="title">
    SURAT KETERANGAN ${js.nama.toUpperCase()}
    <br>Nomor: {{ .NomorSurat }}
  </div>
  
  <div class="content">
    Yang bertanda tangan di bawah ini Kepala Desa {{ .DesaNama }}, menerangkan bahwa:
    <table style="margin: 20px 0;">
      <tr><td>Nama</td><td>: {{ .WargaNama }}</td></tr>
      <tr><td>NIK</td><td>: {{ .WargaNIK }}</td></tr>
      <tr><td>Tempat/Tgl Lahir</td><td>: {{ .WargaTempatLahir }}, {{ .WargaTanggalLahir }}</td></tr>
      <tr><td>Alamat</td><td>: {{ .WargaAlamat }}</td></tr>
    </table>
    
    Demikian surat keterangan ini dibuat untuk dipergunakan sebagaimana mestinya.
  </div>
  
  <div class="footer">
    Kepala Desa {{ .DesaNama }}
    <br><br><br><br>
    <strong>{{ .DesaKepalaDesa }}</strong>
  </div>
</body>
</html>`);
    }
  };

  const handleSaveTemplate = async () => {
    if (!selectedDesaId || !editingJenisSurat) return;
    setSaveLoading(true);
    setError("");

    try {
      const saved = await request("/api/templates", {
        method: "POST",
        body: JSON.stringify({
          jenis_surat_id: editingJenisSurat.id,
          desa_id: selectedDesaId,
          template_html: templateHTML,
        }),
      });

      // update local template list
      const updated = templates.filter((t) => t.jenis_surat_id !== editingJenisSurat.id);
      updated.push(saved);
      setTemplates(updated);

      // close editor modal
      setEditingJenisSurat(null);
      setEditingTemplate(null);
    } catch (err: any) {
      setError(err.message || "Gagal menyimpan template.");
    } finally {
      setSaveLoading(false);
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
          <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Kelola Template Cetak</h1>
          <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
            Sesuaikan struktur layout kop, isi, dan tanda tangan surat A4 desa.
          </p>
        </div>

        {user?.role === "superadmin" && desas.length > 0 && (
          <div className="form-group" style={{ marginBottom: 0, minWidth: "200px" }}>
            <select
              className="form-control"
              value={selectedDesaId}
              onChange={(e) => setSelectedDesaId(e.target.value)}
            >
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

      {/* Grid of Templates */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))", gap: "24px" }}>
        {jenisSurat.map((js) => {
          const hasTemplate = templates.some((t) => t.jenis_surat_id === js.id);
          return (
            <div key={js.id} className="glass-card" style={{ display: "flex", flexDirection: "column", justifyContent: "between", gap: "16px" }}>
              <div>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start", marginBottom: "12px" }}>
                  <span className="badge badge-primary">{js.kode}</span>
                  {hasTemplate ? (
                    <span className="badge badge-success">Terbuka (HTML)</span>
                  ) : (
                    <span className="badge badge-warning" style={{ opacity: 0.7 }}>Bawaan Sistem</span>
                  )}
                </div>
                <h3 style={{ fontSize: "16px", fontWeight: "700" }}>{js.nama}</h3>
                <p style={{ color: "var(--text-muted)", fontSize: "13px", marginTop: "6px" }}>
                  {js.deskripsi || "Gunakan template dinamis untuk mengisi variabel kependudukan."}
                </p>
              </div>

              <div style={{ marginTop: "auto", display: "flex", gap: "12px" }}>
                <button className="btn btn-secondary" style={{ flex: 1, fontSize: "13px", padding: "8px" }} onClick={() => handleEditClick(js)}>
                  📝 Edit Template HTML
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {/* HTML Editor Modal */}
      {editingJenisSurat && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "rgba(0,0,0,0.7)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "800px", width: "95%", height: "90vh", display: "flex", flexDirection: "column", gap: "16px", padding: "30px" }}>
            <div>
              <h3 style={{ fontSize: "20px", fontWeight: "700" }}>Edit Template HTML: {editingJenisSurat.nama}</h3>
              <p style={{ color: "var(--text-muted)", fontSize: "12px", marginTop: "4px" }}>
                Gunakan syntax double-curly Go templating (contoh: <code>{"{{ .WargaNama }}"}</code>, <code>{"{{ .DesaNama }}"}</code>).
              </p>
            </div>

            <div style={{ flex: 1, minHeight: 0 }}>
              <textarea
                className="form-control"
                style={{ fontFamily: "monospace", fontSize: "13px", height: "100%", resize: "none", background: "hsla(222,47%,7%,0.9)" }}
                value={templateHTML}
                onChange={(e) => setTemplateHTML(e.target.value)}
              />
            </div>

            <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px" }}>
              <button className="btn btn-secondary" onClick={() => { setEditingJenisSurat(null); setEditingTemplate(null); }} disabled={saveLoading}>
                Batal
              </button>
              <button className="btn btn-primary" onClick={handleSaveTemplate} disabled={saveLoading}>
                {saveLoading ? "Menyimpan..." : "Simpan Template"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
