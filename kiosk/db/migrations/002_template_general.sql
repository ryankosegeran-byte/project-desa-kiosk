-- 002_template_general.sql
-- Menambahkan fitur template umum (general) dan format kertas untuk Kiosk SQLite

-- 1. Tambahkan kolom is_general dan format_kertas
-- Catatan: ALTER TABLE ADD COLUMN di SQLite tidak mendukung IF NOT EXISTS
-- sehingga migrasi ini akan error jika kolom sudah ada (cek manual atau hapus migrasi ini)
ALTER TABLE surat_template ADD COLUMN is_general INTEGER DEFAULT 0;
ALTER TABLE surat_template ADD COLUMN format_kertas TEXT DEFAULT 'A4';

-- 2. Create index untuk query template hierarchy (SQLite syntax)
CREATE INDEX IF NOT EXISTS idx_template_jenis_general ON surat_template(jenis_surat_id, is_general);

-- 3. Tabel untuk history versi template (audit trail)
CREATE TABLE IF NOT EXISTS template_version_history (
    id              TEXT PRIMARY KEY,
    template_id     TEXT NOT NULL,
    version         INTEGER NOT NULL,
    template_html   TEXT NOT NULL,
    format_kertas  TEXT DEFAULT 'A4',
    change_note     TEXT,
    created_by      TEXT,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tpl_version_template ON template_version_history(template_id);
CREATE INDEX IF NOT EXISTS idx_tpl_version_date ON template_version_history(created_at);
