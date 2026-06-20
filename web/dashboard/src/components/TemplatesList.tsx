import React, { useEffect, useState, useRef } from "react";
import { request, getUser, API_BASE } from "../lib/api";
import { DOCXImportWizard } from "./DOCXImportWizard";
import { DocxTemplateWizard } from "./DocxTemplateWizard";
import { VersionHistory } from "./VersionHistory";

// Import Quill dynamically for WYSIWYG editor
let Quill: any = null;
if (typeof window !== "undefined") {
  import("quill").then((module) => {
    Quill = module.default;
  });
}

// docx-preview: renders DOCX → HTML in browser (no server tools needed)
let renderDocxAsync: ((data: ArrayBuffer, el: HTMLElement) => Promise<void>) | null = null;
if (typeof window !== "undefined") {
  import("docx-preview").then((m) => {
    renderDocxAsync = (data, el) =>
      m.renderAsync(data, el, undefined, {
        className: "docx-preview-page",
        inWrapper: true,
        ignoreWidth: false,
        ignoreHeight: false,
        breakPages: true,
        renderHeaders: true,
        renderFooters: true,
        renderFootnotes: true,
      });
  });
}

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
  is_general: boolean;
  format_kertas: string;
  version: number;
}

interface Desa {
  id: string;
  nama: string;
}

// Available variables for template
const TEMPLATE_VARIABLES = [
  // Data Pemohon
  { key: "{{.Warga.Nama}}", desc: "Nama Lengkap" },
  { key: "{{.Warga.NIK}}", desc: "NIK" },
  { key: "{{.Warga.TempatLahir}}", desc: "Tempat Lahir" },
  { key: "{{.Warga.TanggalLahir}}", desc: "Tanggal Lahir" },
  { key: "{{.Warga.JenisKelamin}}", desc: "L / P" },
  { key: "{{.Warga.Agama}}", desc: "Agama" },
  { key: "{{.Warga.StatusKawin}}", desc: "Status Kawin" },
  { key: "{{.Warga.Alamat}}", desc: "Alamat Lengkap" },
  // Data Desa
  { key: "{{.Warga.Kelurahan}}", desc: "Desa" },
  { key: "{{.Warga.Kecamatan}}", desc: "Kecamatan" },
  { key: "{{.Warga.Kabupaten}}", desc: "Kabupaten" },
  { key: "{{.DesaNama}}", desc: "Nama Desa" },
  { key: "{{.DesaKepalaDesa}}", desc: "Nama Kepala Desa" },
  { key: "{{.DesaNIP}}", desc: "NIP Kepala Desa" },
  // Surat
  { key: "{{.NomorSurat}}", desc: "Nomor Surat" },
  { key: "{{.DateToday}}", desc: "Tanggal Surat" },
  // Data Surat Fields (SKU)
  { key: "{{index .DataSurat \"tahun_kewajiban\"}}", desc: "Tahun Kewajiban" },
  { key: "{{index .DataSurat \"tahun_mulai_usaha\"}}", desc: "Tahun Mulai Usaha" },
  { key: "{{index .DataSurat \"jenis_usaha\"}}", desc: "Jenis Usaha" },
  { key: "{{index .DataSurat \"merk_usaha\"}}", desc: "Merk/Nama Usaha" },
  { key: "{{index .DataSurat \"alamat_usaha\"}}", desc: "Alamat Usaha" },
  { key: "{{index .DataSurat \"batas_utara\"}}", desc: "Batas Utara" },
  { key: "{{index .DataSurat \"batas_selatan\"}}", desc: "Batas Selatan" },
  { key: "{{index .DataSurat \"batas_timur\"}}", desc: "Batas Timur" },
  { key: "{{index .DataSurat \"batas_barat\"}}", desc: "Batas Barat" },
  { key: "{{index .DataSurat \"sifat_tempat_usaha\"}}", desc: "Sifat Usaha" },
];

// Default template starter - Based on reference SKU Kalawat
const getDefaultTemplate = (jenisSurat: JenisSurat) => `<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: 'Times New Roman', serif; margin: 50px 60px; line-height: 1.6; font-size: 12pt; color: #000; }
    .header { text-align: center; border-bottom: 3px double #000; padding-bottom: 15px; margin-bottom: 25px; }
    .header-line { font-size: 11pt; margin: 2px 0; }
    .header-title { font-size: 16pt; font-weight: bold; text-transform: uppercase; margin-top: 10px; }
    .title { text-align: center; font-weight: bold; text-decoration: underline; text-transform: uppercase; font-size: 14pt; margin: 20px 0 10px; }
    .nomor { text-align: left; margin-bottom: 20px; font-size: 12pt; }
    .content { margin-bottom: 15px; text-align: justify; line-height: 1.8; }
    .data-table { margin: 15px 0 20px; border-collapse: collapse; width: 100%; }
    .data-table td { padding: 4px 10px; vertical-align: top; }
    .data-table td:first-child { width: 180px; font-weight: normal; }
    .numbered-list { margin: 15px 0; padding-left: 30px; }
    .numbered-list li { margin-bottom: 10px; text-align: justify; line-height: 1.6; }
    .business-table { margin: 10px 0 10px 30px; border-collapse: collapse; }
    .business-table td { padding: 3px 8px; vertical-align: top; }
    .business-table td:first-child { width: 180px; }
    .batas-table { margin: 5px 0 5px 30px; border-collapse: collapse; }
    .batas-table td { padding: 2px 5px; }
    .footer { float: right; text-align: center; margin-top: 40px; width: 300px; }
    .footer-name { margin-top: 60px; text-decoration: underline; font-weight: bold; }
    .footer-nip { font-size: 10pt; }
    p { margin: 8px 0; text-align: justify; }
  </style>
</head>
<body>
  <!-- HEADER -->
  <div class="header">
    <p class="header-line" style="font-weight: bold; font-size: 14pt;">HUKUM TUA {{.Warga.Kelurahan}}</p>
    <p class="header-line">KECAMATAN {{.Warga.Kecamatan}}</p>
    <p class="header-line">KABUPATEN {{.Warga.Kabupaten}}</p>
  </div>

  <!-- TITLE & NOMOR -->
  <div class="title">SURAT KETERANGAN USAHA</div>
  <div class="nomor">Nomor : {{.NomorSurat}}</div>

  <!-- OPENING -->
  <p class="content">
    Yang bertanda tangan di bawah ini Hukum Tua Desa {{.Warga.Kelurahan}} Kecamatan {{.Warga.Kecamatan}} Kabupaten {{.Warga.Kabupaten}}, bahwa :
  </p>

  <!-- DATA TABLE -->
  <table class="data-table">
    <tr><td>Nama Lengkap</td><td>: <strong>{{.Warga.Nama}}</strong></td></tr>
    <tr><td>NIK</td><td>: {{.Warga.NIK}}</td></tr>
    <tr><td>Tempat/Tanggal Lahir</td><td>: {{.Warga.TempatLahir}}, {{.Warga.TanggalLahir}}</td></tr>
    <tr><td>Jenis Kelamin</td><td>: {{if eq .Warga.JenisKelamin "L"}}Laki-laki{{else}}Perempuan{{end}}</td></tr>
    <tr><td>Agama</td><td>: {{.Warga.Agama}}</td></tr>
    <tr><td>Status</td><td>: {{.Warga.StatusKawin}}</td></tr>
    <tr><td>Alamat</td><td>: {{.Warga.Alamat}}</td></tr>
  </table>

  <!-- NUMBERED STATEMENTS -->
  <ol class="numbered-list">
    <li>Bahwa yang bersangkutan adalah masyarakat Desa {{.Warga.Kelurahan}} Kecamatan {{.Warga.Kecamatan}},</li>
    <li>Bahwa yang bersangkutan tidak pernah tersangkut suatu urusan perkara di desa,</li>
    <li>Bahwa yang bersangkutan telah melunasi semua kewajibannya sebagai warga masyarakat Desa {{.Warga.Kelurahan}} seperti : PBB, uang sampah, dan lain sebagainya untuk tahun {{index .DataSurat "tahun_kewajiban"}},</li>
    <li>Bahwa usaha yang bersangkutan dimulai sejak Tahun {{index .DataSurat "tahun_mulai_usaha"}},</li>
    <li>Bahwa yang bersangkutan <strong>"mempunyai usaha"</strong> :</li>
  </ol>

  <!-- BUSINESS DETAILS -->
  <table class="business-table">
    <tr><td>1. Jenis Usaha</td><td>: {{index .DataSurat "jenis_usaha"}}</td></tr>
    <tr><td>2. Merk/Nama Usaha</td><td>: "{{index .DataSurat "merk_usaha"}}"</td></tr>
    <tr><td>3. Alamat tempat usaha</td><td>: {{index .DataSurat "alamat_usaha"}}</td></tr>
    <tr>
      <td>4. Batas-batas Tempat Usaha</td>
      <td>:
        <table class="batas-table">
          <tr><td>Utara</td><td>: {{index .DataSurat "batas_utara"}}</td></tr>
          <tr><td>Selatan</td><td>: {{index .DataSurat "batas_selatan"}}</td></tr>
          <tr><td>Timur</td><td>: {{index .DataSurat "batas_timur"}}</td></tr>
          <tr><td>Barat</td><td>: {{index .DataSurat "batas_barat"}}</td></tr>
        </table>
      </td>
    </tr>
    <tr><td>5. Sifat Tempat Usaha</td><td>: {{index .DataSurat "sifat_tempat_usaha"}}</td></tr>
  </table>

  <!-- CLOSING -->
  <p class="content" style="margin-top: 20px;">
    Demikian Surat Keterangan Mempunyai Usaha ini dibuat dengan benar dan dipergunakan sebagaimana mestinya.
  </p>

  <!-- FOOTER / TTD -->
  <div class="footer">
    {{.Warga.Kelurahan}}, {{.DateToday}}
    <br>Pj. HUKUM TUA,
    <p class="footer-name">{{.DesaKepalaDesa}}</p>
    <p class="footer-nip">{{.DesaNIP}}</p>
  </div>
</body>
</html>`;

export default function TemplatesList() {
  const [jenisSurat, setJenisSurat] = useState<JenisSurat[]>([]);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [desas, setDesas] = useState<Desa[]>([]);
  const [selectedDesaId, setSelectedDesaId] = useState("");
  const [editingTemplate, setEditingTemplate] = useState<Template | null>(null);
  const [editingJenisSurat, setEditingJenisSurat] = useState<JenisSurat | null>(null);
  const [templateHTML, setTemplateHTML] = useState("");
  const [formatKertas, setFormatKertas] = useState("A4");
  const [isGeneral, setIsGeneral] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saveLoading, setSaveLoading] = useState(false);
  const [error, setError] = useState("");
  const [activeTab, setActiveTab] = useState<"editor" | "preview" | "variables" | "history">("editor");
  const [showImportWizard, setShowImportWizard] = useState(false);
  const [showVersionHistory, setShowVersionHistory] = useState(false);
  const [docxWizardJS, setDocxWizardJS] = useState<JenisSurat | null>(null);

  // Preview Print modal state
  const [previewJenisSurat, setPreviewJenisSurat] = useState<JenisSurat | null>(null);
  const [previewTemplateData, setPreviewTemplateData] = useState<Template | null>(null);
  const [docxBuffer, setDocxBuffer] = useState<ArrayBuffer | null>(null);
  const [previewPdfUrl, setPreviewPdfUrl] = useState<string | null>(null);
  const [docxPreviewLoading, setDocxPreviewLoading] = useState(false);
  const [docxPreviewError, setDocxPreviewError] = useState("");
  const printRef = useRef<HTMLDivElement>(null);
  const docxContainerRef = useRef<HTMLDivElement>(null);

  // Quill editor ref
  const editorRef = useRef<HTMLDivElement>(null);
  const quillRef = useRef<any>(null);

  const user = getUser();

  useEffect(() => {
    async function loadData() {
      try {
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

        // Load jenis surat dari API. JANGAN pakai fallback ID dummy ("1","2",...) —
        // ID harus UUID asli agar aksi seperti upload DOCX tidak mengirim ID palsu.
        let jsData: JenisSurat[] = [];
        try {
          const apiData = await request("/api/jenis-surat");
          if (apiData && apiData.length > 0) {
            jsData = apiData;
          }
        } catch (e) {
          console.error("Gagal memuat jenis surat dari API", e);
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

  // Initialize Quill editor
  useEffect(() => {
    if (editingJenisSurat && editorRef.current && !quillRef.current) {
      // Dynamic import Quill CSS
      const link = document.createElement("link");
      link.rel = "stylesheet";
      link.href = "https://cdn.quilljs.com/1.3.7/quill.snow.css";
      document.head.appendChild(link);

      setTimeout(() => {
        if (Quill && editorRef.current && !quillRef.current) {
          quillRef.current = new Quill(editorRef.current, {
            theme: "snow",
            modules: {
              toolbar: [
                [{ header: [1, 2, 3, false] }],
                ["bold", "italic", "underline", "strike"],
                [{ list: "ordered" }, { list: "bullet" }],
                [{ align: [] }],
                ["link", "image"],
                ["clean"],
              ],
            },
          });
          quillRef.current.on("text-change", () => {
            setTemplateHTML(quillRef.current.root.innerHTML);
          });
        }
      }, 100);
    }
  }, [editingJenisSurat]);

  const handleEditClick = (js: JenisSurat) => {
    const existing = templates.find((t) => t.jenis_surat_id === js.id);
    setEditingJenisSurat(js);

    if (existing) {
      setEditingTemplate(existing);
      setTemplateHTML(existing.template_html);
      setFormatKertas(existing.format_kertas || "A4");
      setIsGeneral(existing.is_general || false);

      // Update Quill if initialized
      if (quillRef.current) {
        quillRef.current.root.innerHTML = existing.template_html;
      }
    } else {
      setEditingTemplate(null);
      setTemplateHTML(getDefaultTemplate(js));
      setFormatKertas("A4");
      setIsGeneral(false);

      // Update Quill if initialized
      if (quillRef.current) {
        quillRef.current.root.innerHTML = getDefaultTemplate(js);
      }
    }

    setActiveTab("editor");
    setShowImportWizard(false);
  };

  const handleSaveTemplate = async () => {
    if (!editingJenisSurat) return;
    setSaveLoading(true);
    setError("");

    try {
      const payload: any = {
        jenis_surat_id: editingJenisSurat.id,
        template_html: templateHTML,
        format_kertas: formatKertas,
        is_general: isGeneral,
      };

      // Only include desa_id for per-desa templates
      if (!isGeneral) {
        payload.desa_id = selectedDesaId;
      }

      const saved = await request("/api/templates", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      // update local template list
      const updated = templates.filter((t) => t.jenis_surat_id !== editingJenisSurat.id);
      updated.push(saved);
      setTemplates(updated);

      // Update editing template
      setEditingTemplate(saved);

      alert("Template berhasil disimpan!");
    } catch (err: any) {
      setError(err.message || "Gagal menyimpan template.");
    } finally {
      setSaveLoading(false);
    }
  };

  // Handle rollback from version history
  const handleRollback = (html: string, format: string, version: number) => {
    setTemplateHTML(html);
    setFormatKertas(format);

    // Update Quill if initialized
    if (quillRef.current) {
      quillRef.current.root.innerHTML = html;
    }

    setActiveTab("editor");
    setShowVersionHistory(false);

    alert(`Template dikembalikan ke versi ${version}. Klik "Simpan Template" untuk menyimpan perubahan.`);
  };

  // Handle DOCX import
  const handleDOCXImport = (html: string, placeholders: string[]) => {
    setTemplateHTML(html);

    // Update Quill if initialized
    if (quillRef.current) {
      quillRef.current.root.innerHTML = html;
    }

    setShowImportWizard(false);
    setActiveTab("preview");

    alert(`${placeholders.length} placeholder berhasil dipetakan ke variabel template!`);
  };

  // Sample data for preview — used by both editor preview and card preview
  const SAMPLE_DATA: Record<string, string> = {
    "{{.Warga.Nama}}": "BUDI SANTOSO",
    "{{.Warga.NIK}}": "3201234567890001",
    "{{.Warga.TempatLahir}}": "Manado",
    "{{.Warga.TanggalLahir}}": "17 Juni 1990",
    "{{.Warga.JenisKelamin}}": "L",
    "{{.Warga.Agama}}": "Kristen",
    "{{.Warga.StatusKawin}}": "Kawin",
    "{{.Warga.Alamat}}": "Jl. Sudirman No. 10",
    "{{.Warga.Kelurahan}}": "Kalawat",
    "{{.Warga.Kecamatan}}": "Kalawat",
    "{{.Warga.Kabupaten}}": "Minahasa Utara",
    "{{.DesaNama}}": "Kalawat",
    "{{.DesaKepalaDesa}}": "ALFRIDA, A.Md.Kes",
    "{{.DesaNIP}}": "196902061993032004",
    "{{.NomorSurat}}": "001/SKU/08.10/VI/2026",
    "{{.DateToday}}": "17 Juni 2026",
    // Data Surat fields for SKU
    "{{index .DataSurat \"tahun_kewajiban\"}}": "2026",
    "{{index .DataSurat \"tahun_mulai_usaha\"}}": "2020",
    "{{index .DataSurat \"jenis_usaha\"}}": "Toko Kelontong",
    "{{index .DataSurat \"merk_usaha\"}}": "Toko Budi",
    "{{index .DataSurat \"alamat_usaha\"}}": "Jl. Desa Kalawat No. 5",
    "{{index .DataSurat \"batas_utara\"}}": "Toko Sari",
    "{{index .DataSurat \"batas_selatan\"}}": "Jalan Desa",
    "{{index .DataSurat \"batas_timur\"}}": "Sungai kecil",
    "{{index .DataSurat \"batas_barat\"}}": "Tanah kosong",
    "{{index .DataSurat \"sifat_tempat_usaha\"}}": "Permanen",
  };

  // Renders preview HTML from a given template string with dummy data
  const renderPreviewHTML = (html: string): string => {
    let preview = html;
    // Replace conditionals with sample values (male branch)
    preview = preview.replace(/\{\{if eq \.Warga\.JenisKelamin "L"\}\}([\s\S]*?)\{\{else\}\}([\s\S]*?)\{\{end\}\}/g, '$1');
    // Replace all variables
    Object.entries(SAMPLE_DATA).forEach(([key, value]) => {
      preview = preview.split(key).join(value);
    });
    return preview;
  };

  // Preview HTML with sample data (for editor preview tab)
  const getPreviewHTML = () => renderPreviewHTML(templateHTML);

  // Render DOCX into container whenever buffer arrives
  useEffect(() => {
    if (!docxBuffer || !docxContainerRef.current || !renderDocxAsync) return;
    const el = docxContainerRef.current;
    el.innerHTML = "";
    renderDocxAsync(docxBuffer, el).catch((err: Error) => {
      setDocxPreviewError("docx-preview: " + err.message);
    });
  }, [docxBuffer]);

  // Open preview print modal for a specific card
  const handlePreviewPrint = async (js: JenisSurat) => {
    const perDesa = templates.find((t) => t.jenis_surat_id === js.id && !t.is_general);
    const general = templates.find((t) => t.jenis_surat_id === js.id && t.is_general);
    const tpl = perDesa || general || null;
    setPreviewJenisSurat(js);
    setPreviewTemplateData(tpl);
    setDocxBuffer(null);
    if (previewPdfUrl) { URL.revokeObjectURL(previewPdfUrl); }
    setPreviewPdfUrl(null);
    setDocxPreviewError("");

    // If this is a DOCX-only template (has id but no template_html), fetch filled preview from server.
    // Server returns PDF (LibreOffice) when available, otherwise DOCX (docx-preview fallback).
    if (tpl && !tpl.template_html && tpl.id) {
      setDocxPreviewLoading(true);
      try {
        const token = localStorage.getItem("token");
        const res = await fetch(`${API_BASE}/api/templates/${tpl.id}/preview`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
        });
        if (!res.ok) {
          const errData = await res.json().catch(() => ({ error: "Gagal memuat preview" }));
          throw new Error(errData.error || "Gagal memuat preview");
        }
        const contentType = res.headers.get("content-type") || "";
        if (contentType.includes("application/pdf")) {
          const blob = await res.blob();
          setPreviewPdfUrl(URL.createObjectURL(blob));
        } else {
          const buf = await res.arrayBuffer();
          setDocxBuffer(buf);
        }
      } catch (err: any) {
        setDocxPreviewError(err.message || "Gagal memuat preview");
      } finally {
        setDocxPreviewLoading(false);
      }
    }
  };

  // Print the preview content
  const handlePrint = () => {
    window.print();
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
            Edit template surat dengan WYSIWYG editor, preview real-time, dan pilih format kertas.
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
          const perDesaTemplate = templates.find((t) => t.jenis_surat_id === js.id && !t.is_general);
          const generalTemplate = templates.find((t) => t.jenis_surat_id === js.id && t.is_general);
          const hasTemplate = perDesaTemplate || generalTemplate;

          return (
            <div key={js.id} className="glass-card" style={{ display: "flex", flexDirection: "column", justifyContent: "between", gap: "16px" }}>
              <div>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start", marginBottom: "12px" }}>
                  <span className="badge badge-primary">{js.kode}</span>
                  {hasTemplate ? (
                    <span className="badge badge-success">
                      {perDesaTemplate ? "Per-Desa" : "Umum"}
                    </span>
                  ) : (
                    <span className="badge badge-warning" style={{ opacity: 0.7 }}>Default</span>
                  )}
                </div>
                <h3 style={{ fontSize: "16px", fontWeight: "700" }}>{js.nama}</h3>
                <p style={{ color: "var(--text-muted)", fontSize: "13px", marginTop: "6px" }}>
                  {hasTemplate ? (
                    <span>Format: {hasTemplate.format_kertas || "A4"} | v{hasTemplate.version}</span>
                  ) : (
                    <span>Gunakan template default sistem</span>
                  )}
                </p>
              </div>

              <div style={{ marginTop: "auto", display: "flex", flexDirection: "column", gap: "8px" }}>
                <div style={{ display: "flex", gap: "8px" }}>
                  <button
                    className="btn btn-secondary"
                    style={{ flex: 1, fontSize: "13px", padding: "8px" }}
                    onClick={() => handleEditClick(js)}
                  >
                    📝 HTML
                  </button>
                  <button
                    className="btn btn-primary"
                    style={{ flex: 1, fontSize: "13px", padding: "8px" }}
                    onClick={() => setDocxWizardJS(js)}
                  >
                    📄 DOCX
                  </button>
                </div>
                <button
                  className="btn"
                  style={{
                    width: "100%",
                    fontSize: "13px",
                    padding: "8px",
                    background: "linear-gradient(135deg, hsla(270, 80%, 55%, 0.9), hsla(210, 100%, 55%, 0.9))",
                    color: "#fff",
                    border: "none",
                    boxShadow: "0 4px 12px hsla(270, 80%, 55%, 0.3)",
                  }}
                  onClick={() => handlePreviewPrint(js)}
                >
                  🖨️ Preview Print
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {/* Template Editor Modal */}
      {editingJenisSurat && (
        <div style={{
          position: "fixed",
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: "rgba(0,0,0,0.8)",
          backdropFilter: "blur(4px)",
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          zIndex: 1000,
          padding: "20px"
        }}>
          <div className="glass-card" style={{
            maxWidth: "1400px",
            width: "100%",
            height: "95vh",
            display: "flex",
            flexDirection: "column",
            gap: "16px",
            padding: "24px"
          }}>
            {/* Header */}
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start" }}>
              <div>
                <h3 style={{ fontSize: "20px", fontWeight: "700" }}>
                  Edit Template: {editingJenisSurat.nama}
                  {editingTemplate && (
                    <span style={{ fontSize: "14px", fontWeight: 400, color: "var(--text-muted)", marginLeft: "12px" }}>
                      v{editingTemplate.version}
                    </span>
                  )}
                </h3>
                <p style={{ color: "var(--text-muted)", fontSize: "12px", marginTop: "4px" }}>
                  Gunakan editor WYSIWYG atau edit langsung HTML. Klik "Insert Variable" untuk menambahkan placeholder.
                </p>
              </div>
              <button
                className="btn btn-secondary"
                onClick={() => {
                  setEditingJenisSurat(null);
                  setEditingTemplate(null);
                  if (quillRef.current) {
                    quillRef.current = null;
                  }
                }}
              >
                ✕ Tutup
              </button>
            </div>

            {/* Options Bar */}
            <div style={{ display: "flex", gap: "16px", alignItems: "center", padding: "12px", background: "rgba(0,0,0,0.2)", borderRadius: "8px" }}>
              {/* Format Kertas */}
              <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
                <label style={{ fontSize: "13px", color: "var(--text-muted)" }}>Format:</label>
                <select
                  className="form-control"
                  style={{ width: "120px", padding: "6px 12px" }}
                  value={formatKertas}
                  onChange={(e) => setFormatKertas(e.target.value)}
                >
                  <option value="A4">A4 (210×297mm)</option>
                  <option value="F4">F4 (215×330mm)</option>
                </select>
              </div>

              {/* General Toggle (Superadmin only) */}
              {user?.role === "superadmin" && (
                <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
                  <label style={{ fontSize: "13px", color: "var(--text-muted)" }}>Tipe:</label>
                  <select
                    className="form-control"
                    style={{ width: "140px", padding: "6px 12px" }}
                    value={isGeneral ? "general" : "desa"}
                    onChange={(e) => setIsGeneral(e.target.value === "general")}
                  >
                    <option value="desa">Per-Desa</option>
                    <option value="general">Umum (Semua Desa)</option>
                  </select>
                </div>
              )}

              {/* History Button */}
              {editingTemplate && (
                <button
                  className="btn btn-secondary"
                  onClick={() => setShowVersionHistory(true)}
                >
                  📜 Riwayat
                </button>
              )}

              {/* Import DOCX Button */}
              <button
                className="btn btn-secondary"
                style={{ marginLeft: "auto" }}
                onClick={() => setShowImportWizard(true)}
              >
                📄 Import DOCX
              </button>
            </div>

            {/* Tab Navigation */}
            <div style={{ display: "flex", gap: "8px", borderBottom: "1px solid var(--border-color)", paddingBottom: "8px" }}>
              <button
                className={`btn ${activeTab === "editor" ? "btn-primary" : "btn-secondary"}`}
                onClick={() => setActiveTab("editor")}
              >
                ✏️ Editor WYSIWYG
              </button>
              <button
                className={`btn ${activeTab === "preview" ? "btn-primary" : "btn-secondary"}`}
                onClick={() => setActiveTab("preview")}
              >
                👁️ Preview
              </button>
              <button
                className={`btn ${activeTab === "variables" ? "btn-primary" : "btn-secondary"}`}
                onClick={() => setActiveTab("variables")}
              >
                📋 Variabel
              </button>
              {editingTemplate && (
                <button
                  className={`btn ${activeTab === "history" ? "btn-primary" : "btn-secondary"}`}
                  onClick={() => setActiveTab("history")}
                >
                  📜 Riwayat
                </button>
              )}
            </div>

            {/* Content Area */}
            <div style={{ flex: 1, minHeight: 0, display: "flex", gap: "16px", overflow: "hidden" }}>
              {/* Editor / Preview Panel */}
              <div style={{ flex: 2, display: "flex", flexDirection: "column", overflow: "hidden" }}>
                {activeTab === "editor" && (
                  <>
                    {/* WYSIWYG Editor */}
                    <div ref={editorRef} style={{ flex: 1, overflow: "auto" }} />

                    {/* HTML Source Toggle */}
                    <details style={{ marginTop: "12px" }}>
                      <summary style={{ cursor: "pointer", color: "var(--text-muted)", fontSize: "13px", marginBottom: "8px" }}>
                        🔧 Edit HTML Source
                      </summary>
                      <textarea
                        className="form-control"
                        style={{
                          fontFamily: "monospace",
                          fontSize: "12px",
                          height: "200px",
                          resize: "vertical",
                          background: "rgba(0,0,0,0.3)"
                        }}
                        value={templateHTML}
                        onChange={(e) => {
                          setTemplateHTML(e.target.value);
                          if (quillRef.current) {
                            quillRef.current.root.innerHTML = e.target.value;
                          }
                        }}
                      />
                    </details>
                  </>
                )}

                {activeTab === "preview" && (
                  <div
                    style={{
                      flex: 1,
                      overflow: "auto",
                      background: "#fff",
                      borderRadius: "8px",
                      padding: "20px",
                      display: "flex",
                      justifyContent: "center"
                    }}
                  >
                    <div
                      style={{
                        width: formatKertas === "F4" ? "700px" : "650px",
                        background: "#fff",
                        color: "#000",
                        fontSize: "12px"
                      }}
                      dangerouslySetInnerHTML={{ __html: getPreviewHTML() }}
                    />
                  </div>
                )}

                {activeTab === "variables" && (
                  <div style={{ flex: 1, overflow: "auto", padding: "16px" }}>
                    <h4 style={{ marginBottom: "16px" }}>Variabel Template Tersedia</h4>
                    <p style={{ color: "var(--text-muted)", fontSize: "13px", marginBottom: "16px" }}>
                      Klik variabel untuk menambahkan ke editor:
                    </p>
                    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "8px" }}>
                      {TEMPLATE_VARIABLES.map((v) => (
                        <div
                          key={v.key}
                          style={{
                            padding: "8px 12px",
                            background: "rgba(0,0,0,0.2)",
                            borderRadius: "6px",
                            cursor: "pointer",
                            display: "flex",
                            justifyContent: "space-between",
                            alignItems: "center"
                          }}
                          onClick={() => {
                            if (quillRef.current) {
                              const range = quillRef.current.getSelection();
                              quillRef.current.insertText(range?.index || 0, v.key);
                            }
                            setActiveTab("editor");
                          }}
                        >
                          <code style={{ fontSize: "11px", color: "var(--primary)" }}>{v.key}</code>
                          <span style={{ fontSize: "11px", color: "var(--text-muted)" }}>{v.desc}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {activeTab === "history" && editingTemplate && (
                  <div style={{ flex: 1, overflow: "auto", padding: "16px" }}>
                    <VersionHistory
                      templateId={editingTemplate.id}
                      currentHTML={templateHTML}
                      currentFormatKertas={formatKertas}
                      onRollback={handleRollback}
                      onClose={() => setActiveTab("editor")}
                    />
                  </div>
                )}
              </div>
            </div>

            {/* Footer Actions */}
            <div style={{ display: "flex", justifyContent: "space-between", paddingTop: "16px", borderTop: "1px solid var(--border-color)" }}>
              <div style={{ display: "flex", gap: "12px" }}>
                <button
                  className="btn btn-secondary"
                  onClick={() => {
                    setEditingJenisSurat(null);
                    setEditingTemplate(null);
                  }}
                  disabled={saveLoading}
                >
                  Batal
                </button>
              </div>
              <button
                className="btn btn-primary"
                onClick={handleSaveTemplate}
                disabled={saveLoading}
              >
                {saveLoading ? "Menyimpan..." : "💾 Simpan Template"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* DOCX Import Wizard (HTML, lama) */}
      {showImportWizard && (
        <DOCXImportWizard
          onImport={handleDOCXImport}
          onClose={() => setShowImportWizard(false)}
        />
      )}

      {/* DOCX Template Wizard (Strategi B: docx master + pemetaan token) */}
      {docxWizardJS && (
        <DocxTemplateWizard
          jenisSurat={docxWizardJS}
          desaId={selectedDesaId}
          onClose={() => setDocxWizardJS(null)}
          onSaved={async () => {
            if (selectedDesaId) {
              try {
                const templs = await request(`/api/templates?desa_id=${selectedDesaId}`);
                setTemplates(templs);
              } catch {
                /* ignore reload error */
              }
            }
          }}
        />
      )}

      {/* Version History Modal (Standalone) */}
      {showVersionHistory && editingTemplate && !activeTab.includes("history") && (
        <VersionHistory
          templateId={editingTemplate.id}
          currentHTML={templateHTML}
          currentFormatKertas={formatKertas}
          onRollback={handleRollback}
          onClose={() => setShowVersionHistory(false)}
        />
      )}

      {/* ====== PREVIEW PRINT MODAL ====== */}
      {previewJenisSurat && (
        <>
          {/* Print-only stylesheet: when printing, hide everything except #print-area */}
          <style>{`
            @media print {
              body > *:not(#print-preview-portal) { display: none !important; }
              #print-preview-portal { position: static !important; }
              #print-preview-portal * { visibility: visible !important; }
              .preview-modal-chrome { display: none !important; }
              #print-area {
                box-shadow: none !important;
                margin: 0 !important;
                border-radius: 0 !important;
                background: #fff !important;
              }
              @page {
                size: ${(previewTemplateData?.format_kertas || "A4") === "F4" ? "215mm 330mm" : "A4"};
                margin: 0;
              }
            }
          `}</style>

          <div
            id="print-preview-portal"
            style={{
              position: "fixed",
              inset: 0,
              background: "rgba(0,0,0,0.88)",
              backdropFilter: "blur(6px)",
              display: "flex",
              flexDirection: "column",
              zIndex: 1300,
              animation: "fadeInPreview 0.25s ease",
            }}
          >
            {/* Inline animation keyframe */}
            <style>{`
              @keyframes fadeInPreview {
                from { opacity: 0; transform: scale(0.97); }
                to   { opacity: 1; transform: scale(1); }
              }
              @keyframes slideUpPreview {
                from { opacity: 0; transform: translateY(12px); }
                to   { opacity: 1; transform: translateY(0); }
              }
            `}</style>

            {/* Top Chrome Bar */}
            <div
              className="preview-modal-chrome"
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                padding: "14px 24px",
                borderBottom: "1px solid hsla(220,30%,25%,0.6)",
                background: "hsla(223,47%,11%,0.95)",
                backdropFilter: "blur(12px)",
                flexShrink: 0,
              }}
            >
              <div style={{ display: "flex", alignItems: "center", gap: "14px" }}>
                <div style={{
                  width: "36px",
                  height: "36px",
                  borderRadius: "8px",
                  background: "linear-gradient(135deg, hsla(270,80%,55%,0.9), hsla(210,100%,55%,0.9))",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: "18px",
                  boxShadow: "0 0 12px hsla(270,80%,55%,0.4)",
                }}>
                  🖨️
                </div>
                <div>
                  <h3 style={{ fontSize: "17px", fontWeight: 700 }}>
                    Preview Cetak: {previewJenisSurat.nama}
                  </h3>
                  <p style={{ fontSize: "12px", color: "var(--text-muted)", marginTop: "2px" }}>
                    Data contoh (dummy) • Format: {previewTemplateData?.format_kertas || "A4"}
                    {previewTemplateData && ` • v${previewTemplateData.version}`}
                  </p>
                </div>
              </div>

              <div style={{ display: "flex", gap: "10px", alignItems: "center" }}>
                <span
                  className="badge badge-warning"
                  style={{ fontSize: "11px", padding: "5px 10px" }}
                >
                  CONTOH
                </span>
                <button
                  className="btn"
                  style={{
                    background: "linear-gradient(135deg, hsla(145,80%,45%,0.9), hsla(160,80%,40%,0.9))",
                    color: "#fff",
                    fontSize: "14px",
                    padding: "9px 20px",
                    border: "none",
                    boxShadow: "0 4px 12px hsla(145,80%,45%,0.3)",
                  }}
                  onClick={handlePrint}
                >
                  🖨️ Cetak
                </button>
                <button
                  className="btn btn-secondary"
                  style={{ fontSize: "14px", padding: "9px 16px" }}
                  onClick={() => {
                    setPreviewJenisSurat(null);
                    setPreviewTemplateData(null);
                    setDocxBuffer(null);
                    if (previewPdfUrl) { URL.revokeObjectURL(previewPdfUrl); }
                    setPreviewPdfUrl(null);
                    setDocxPreviewError("");
                  }}
                >
                  ✕ Tutup
                </button>
              </div>
            </div>

            {/* Scrollable Paper Area */}
            <div
              style={{
                flex: 1,
                overflow: "auto",
                display: "flex",
                justifyContent: "center",
                padding: "32px 20px 60px",
                background: "linear-gradient(180deg, hsla(222,47%,12%,1) 0%, hsla(222,47%,8%,1) 100%)",
              }}
            >
              {/* Determine content */}
              {(() => {
                const tpl = previewTemplateData;
                const html = tpl?.template_html || getDefaultTemplate(previewJenisSurat);
                const hasDocx = tpl && !tpl.template_html && tpl.id; // DOCX-only template
                const paperFormat = tpl?.format_kertas || "A4";
                const paperWidth = paperFormat === "F4" ? "215mm" : "210mm";
                const paperMinHeight = paperFormat === "F4" ? "330mm" : "297mm";

                if (hasDocx) {
                  // DOCX template — show server-rendered preview
                  if (docxPreviewLoading) {
                    return (
                      <div
                        style={{
                          display: "flex",
                          flexDirection: "column",
                          alignItems: "center",
                          justifyContent: "center",
                          padding: "80px 40px",
                          animation: "slideUpPreview 0.35s ease",
                          gap: "20px",
                        }}
                      >
                        <div className="spinner" style={{ width: "40px", height: "40px" }} />
                        <p style={{ color: "var(--text-muted)", fontSize: "14px" }}>
                          Merender preview DOCX...
                        </p>
                      </div>
                    );
                  }

                  if (docxPreviewError) {
                    return (
                      <div
                        style={{
                          maxWidth: "560px",
                          textAlign: "center",
                          padding: "60px 40px",
                          animation: "slideUpPreview 0.35s ease",
                        }}
                      >
                        <div style={{
                          width: "80px",
                          height: "80px",
                          borderRadius: "50%",
                          background: "hsla(355,85%,55%,0.1)",
                          border: "2px solid hsla(355,85%,55%,0.3)",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          fontSize: "36px",
                          margin: "0 auto 24px",
                        }}>
                          ⚠️
                        </div>
                        <h3 style={{ fontSize: "20px", fontWeight: 700, marginBottom: "12px", color: "var(--danger)" }}>
                          Gagal Render Preview
                        </h3>
                        <p style={{ color: "var(--text-muted)", lineHeight: 1.7, fontSize: "14px" }}>
                          {docxPreviewError}
                        </p>
                      </div>
                    );
                  }

                  // PDF from LibreOffice — show in browser's native PDF viewer
                  if (previewPdfUrl) {
                    return (
                      <iframe
                        src={previewPdfUrl}
                        style={{
                          width: "100%",
                          height: "80vh",
                          border: "none",
                          animation: "slideUpPreview 0.35s ease",
                        }}
                        title="Preview Surat"
                      />
                    );
                  }

                  // docx-preview fallback (no LibreOffice on server)
                  return (
                    <div
                      ref={docxContainerRef}
                      style={{
                        width: "100%",
                        animation: "slideUpPreview 0.35s ease",
                      }}
                    />
                  );
                }

                // HTML template — render with dummy data
                const renderedHTML = renderPreviewHTML(html);

                return (
                  <div
                    ref={printRef}
                    id="print-area"
                    style={{
                      width: paperWidth,
                      minHeight: paperMinHeight,
                      background: "#fff",
                      color: "#000",
                      boxShadow: "0 8px 40px rgba(0,0,0,0.5), 0 0 0 1px rgba(255,255,255,0.05)",
                      borderRadius: "4px",
                      overflow: "hidden",
                      animation: "slideUpPreview 0.35s ease",
                      flexShrink: 0,
                    }}
                    dangerouslySetInnerHTML={{ __html: renderedHTML }}
                  />
                );
              })()}
            </div>

            {/* Bottom info bar */}
            <div
              className="preview-modal-chrome"
              style={{
                padding: "10px 24px",
                borderTop: "1px solid hsla(220,30%,25%,0.6)",
                background: "hsla(223,47%,11%,0.95)",
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                gap: "8px",
                flexShrink: 0,
              }}
            >
              <span style={{ fontSize: "12px", color: "var(--text-muted)" }}>
                ⚠️ Ini adalah preview dengan <strong>data contoh</strong>. Isi sebenarnya akan berbeda saat surat dicetak dari kiosk.
              </span>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
