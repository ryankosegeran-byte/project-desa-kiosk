-- 006_template_pdf.sql
-- PDF pendamping (opsional) sebagai "tampilan" template di dashboard.
-- serv00 (FreeBSD) tidak punya Word/LibreOffice untuk render docx->PDF, jadi admin
-- meng-upload PDF hasil export Word secara manual hanya untuk preview visual.
-- NULL = belum ada tampilan PDF. Tidak ikut ke kiosk (kiosk render sendiri via Word).
ALTER TABLE surat_template ADD COLUMN IF NOT EXISTS template_pdf BYTEA;
