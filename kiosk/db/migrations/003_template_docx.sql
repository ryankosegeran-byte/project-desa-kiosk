-- 003_template_docx.sql
-- Strategi B: DOCX sebagai master + pemetaan placeholder per-template (Kiosk SQLite).
-- Catatan: SQLite tidak mendukung IF NOT EXISTS pada ADD COLUMN; runner migrasi
-- (sqlite.go) sudah melewati error "duplicate column name" agar aman dijalankan ulang.

ALTER TABLE surat_template ADD COLUMN template_docx BLOB;
ALTER TABLE surat_template ADD COLUMN placeholders TEXT;
