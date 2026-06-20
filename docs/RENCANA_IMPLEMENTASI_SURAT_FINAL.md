# Rencana Implementasi Final — Sistem Template Surat

> **Menggantikan** `IMPLEMENTATION_PLAN_SURAT.md` v1/v2/v3 (silakan diarsipkan/dihapus).
> **Status arsitektur: TERKUNCI & TERBUKTI** — spike di `template_surat/kalawat/surat_keterangan_usaha/sku.docx` (2026-06-18): kop Garuda utuh + semua nilai terisi, render via Word ~1–2 detik.

---

## 1. Keputusan arsitektur (terkunci)

- **docx tiap desa = master.** Fidelity 100%. **TIDAK** pernah dikonversi ke HTML.
- **Token `{{snake_case}}`** ditandai admin langsung di Word (mis. `{{nama}}`, `{{jenis_usaha}}`), semua format dibiarkan.
- **Render di KIOSK (offline)** via **Microsoft Word** (sudah terpasang di PC kantor desa). Output PDF → cetak via **SumatraPDF** (pipeline yang sudah ada).
- **Server FreeBSD** = admin panel online + sumber **sync/polling**. Internet desa sering putus → kiosk wajib jalan offline; sync hanya saat online.
- Preview di admin panel (server) pakai **LibreOffice** (opsional; FreeBSD tidak ada Word).

## 2. Prinsip yang menyelesaikan masalah lama

Penyebab "render jelek" lama: pipeline `docx → HTML (mammoth) → auto-detect regex/highlight` di `DOCXImportWizard.tsx` meng-inject `{{.DataSurat.x}}` ke data base64 logo (kop hancur) + global-replace "X" → 39 field sampah.

Pendekatan baru:
1. **Jangan ubah docx ke HTML.** Render docx aslinya.
2. **Deteksi token** dengan menggabung isi semua `<w:t>` per paragraf lalu regex `\{\{([a-z0-9_]+)\}\}` — **bukan** global-replace.
3. **Fill di kiosk pakai find-replace Word**, yang otomatis mengatasi masalah "teks terpecah/split run" (terbukti di spike). → **tidak perlu langkah normalisasi docx.**

## 3. Model 3-sumber nilai (sudah cocok dengan kode)

Setiap token punya satu sumber (mirror `tmplData` di `kiosk/print/pdf.go`):

| Sumber | Contoh | Asal nilai | Muncul di form kiosk? |
|---|---|---|---|
| `warga` | `{{nama}}`, `{{nik}}` | data KTP (`wargaRepo.FindByNIK`) | tidak (auto) |
| `manual` | `{{jenis_usaha}}` | form kiosk | **ya** |
| `sistem` | `{{nomor_surat}}`, `{{tanggal}}`, `{{kepala_desa}}` | nomor/tanggal/config desa | tidak (auto) |

## 4. Perubahan data

**Tabel `surat_template`** (server Postgres + kiosk SQLite):
- `+ template_docx` (BYTEA / BLOB) — docx master yang sudah ditandai token.
- `+ placeholders` (JSONB / TEXT) — pemetaan per-template.
- `template_html` & `format_kertas` → opsional/usang untuk template docx (docx membawa ukuran kertasnya sendiri).

**Pindahkan pemetaan field dari `JenisSurat.FieldsSchema` → per-template `placeholders`** (tiap desa beda token). Bentuk `PlaceholderDef`:
```
key, label, source(warga|manual|sistem),
warga_field?, sistem_field?, type?(text|select|date|number), options?, required?, urutan?
```

Migrasi baru: `server/db/migrations/005_template_docx.sql`, `kiosk/db/migrations/003_template_docx.sql`.

## 5. Komponen & file

**Server (FreeBSD) — `server/`**
- `api/template_handler.go`: `POST .../upload-docx` (simpan docx, deteksi token, balikan daftar token), simpan mapping, preview (LibreOffice → PDF).
- `db/template_repo.go`: simpan/baca docx + placeholders.
- `api/sync_handler.go`: ikutkan `template_docx` (base64) + `placeholders` di payload sync.
- util Go baru: deteksi token dari docx (zip + gabung `<w:t>` + regex), cakup `document.xml`, `header*.xml`, `footer*.xml`.

**Kiosk (offline, Windows) — `kiosk/`**
- `print/docx.go` (baru): render docx → PDF via **Word COM** (`github.com/go-ole/go-ole`). Jaga 1 instance Word **"warm"** untuk preview cepat. `Visible=false`, macro disabled. Logika fill = seperti spike.
- `api/surat.go` `handlePrintSurat`: ganti `pdfGen.GeneratePDF(HTML…)` → render docx. **Reuse** nomor surat (`GetNextNumber`/`FormatNomorSurat`), `FindByNIK`, config desa, `printer.PrintPDF`.
- endpoint **live-preview**: render dengan data form (boleh sebagian) → balikan PDF untuk panel preview.

**Admin UI — `web/dashboard/`**
- Wizard baru (ganti `DOCXImportWizard` yang regex): upload docx → tampil token terdeteksi → tiap token pilih sumber + tipe/label → preview PDF → simpan **per desa**.

**Kiosk UI — `web/kiosk-ui/`**
- Form dibangun dari placeholder ber-`source=manual` saja.
- Panel preview menampilkan PDF hasil render (update **debounce** saat isi field).

## 6. Alur lengkap

**Authoring (admin, online):** tandai `{{token}}` di Word → upload → petakan sumber tiap token → preview → simpan per desa di FreeBSD.

**Sync:** polling kiosk → tarik docx + mapping → simpan di SQLite kiosk.

**Kiosk (offline):** warga buka app → scan e-KTP → pilih surat → preview muncul (Word render diam-diam) → isi field manual (preview re-render, debounce) → **Print** → PDF yang sama dikirim ke printer. Nomor surat di-assign offline (sistem yang ada).

> Word/LibreOffice = mesin render **tak terlihat** di latar (sama peran seperti headless Chrome/`chromedp` sekarang). Warga tidak pernah membuka Word.

## 7. Fase pengerjaan

1. **Data model**: migrasi server+kiosk, model `PlaceholderDef`, update repo, util deteksi token (Go).
2. **Mesin render kiosk** (inti): `print/docx.go` (Word via go-ole, warm instance) + swap `handlePrintSurat` + endpoint live-preview. *(reuse logika spike)*
3. **UI admin**: wizard upload + mapping + preview + endpoint.
4. **UI kiosk**: form dari placeholder manual + panel preview.
5. **Sync + bersih-bersih**: sync docx+mapping; hapus `DOCXImportWizard` (regex), `extract_docx.js/.mjs`, `docx_extractor/`; arsipkan docs v1/v2/v3.

## 8. Catatan teknis

- **go-ole** untuk Word COM; jaga 1 instance Word hidup → preview cepat; tangani restart bila Word crash.
- **Hapus highlight kuning** pada token saat fill (atau minta admin un-highlight di template) — nilai mewarisi highlight blank asli.
- **Tidak perlu normalisasi docx**: deteksi via gabungan teks, fill via Word (atasi split run).
- Token: huruf kecil + `_`. **Jangan** "XXX" (penyebab korupsi base64 dulu).
- Konvensi token sistem standar: `{{nomor_surat}}`, `{{tanggal}}`, `{{kepala_desa}}`, `{{nip_kepala_desa}}`.
