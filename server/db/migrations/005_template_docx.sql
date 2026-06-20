-- 005_template_docx.sql
-- Strategi B: DOCX sebagai master + pemetaan placeholder per-template.
-- Lihat docs/RENCANA_IMPLEMENTASI_SURAT_FINAL.md

-- DOCX master yang sudah ditandai token {{...}} (NULL untuk template HTML lama)
ALTER TABLE surat_template ADD COLUMN IF NOT EXISTS template_docx BYTEA;

-- Pemetaan token -> sumber nilai (warga/manual/sistem), per-template
ALTER TABLE surat_template ADD COLUMN IF NOT EXISTS placeholders JSONB;
