-- 004_template_general.sql
-- Menambahkan fitur template umum (general) dan format kertas

-- 1. Tambahkan kolom is_general dan format_kertas ke surat_template
ALTER TABLE surat_template ADD COLUMN IF NOT EXISTS is_general BOOLEAN DEFAULT false;
ALTER TABLE surat_template ADD COLUMN IF NOT EXISTS format_kertas VARCHAR(10) DEFAULT 'A4';

-- 2. Buat index untuk query template hierarchy
CREATE INDEX IF NOT EXISTS idx_template_jenis_general ON surat_template(jenis_surat_id, is_general) WHERE is_general = true;
CREATE INDEX IF NOT EXISTS idx_template_desa ON surat_template(desa_id);

-- 3. Buat unique constraint untuk is_general templates menggunakan partial unique index
-- PostgreSQL tidak support ADD CONSTRAINT UNIQUE WHERE, jadi gunakan CREATE INDEX
CREATE UNIQUE INDEX IF NOT EXISTS unique_template_general
ON surat_template(jenis_surat_id) WHERE is_general = true;

-- 4. Tabel untuk history versi template (audit trail)
CREATE TABLE IF NOT EXISTS template_version_history (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id     UUID NOT NULL REFERENCES surat_template(id) ON DELETE CASCADE,
    version         INTEGER NOT NULL,
    template_html   TEXT NOT NULL,
    format_kertas  VARCHAR(10) DEFAULT 'A4',
    change_note     TEXT,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_template_version_template ON template_version_history(template_id);
CREATE INDEX IF NOT EXISTS idx_template_version_date ON template_version_history(created_at DESC);

-- 5. Constraint untuk format_kertas (skip jika sudah ada)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_format_kertas'
    ) THEN
        ALTER TABLE surat_template ADD CONSTRAINT check_format_kertas CHECK (format_kertas IN ('A4', 'F4'));
    END IF;
END $$;
