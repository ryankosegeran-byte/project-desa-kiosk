-- 002_warga_draft.sql
-- Menambahkan dukungan draft registrasi warga (simpan parsial + link)

-- Tambah kolom status (draft / complete)
ALTER TABLE warga ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'complete';

-- Tambah kolom draft_token untuk link lookup
ALTER TABLE warga ADD COLUMN IF NOT EXISTS draft_token VARCHAR(36);

-- Tambah kolom foto_ktp_path (sudah ada di schema 001, tapi pastikan ada)
ALTER TABLE warga ADD COLUMN IF NOT EXISTS foto_ktp_path VARCHAR(500);

-- Longgarkan constraint NOT NULL pada nik dan nama agar draft bisa kosong
DO $$
BEGIN
    -- Drop NOT NULL pada nik jika masih ada
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'warga' AND column_name = 'nik' AND is_nullable = 'NO'
    ) THEN
        ALTER TABLE warga ALTER COLUMN nik DROP NOT NULL;
    END IF;

    -- Drop NOT NULL pada nama jika masih ada
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'warga' AND column_name = 'nama' AND is_nullable = 'NO'
    ) THEN
        ALTER TABLE warga ALTER COLUMN nama DROP NOT NULL;
    END IF;
END $$;

-- Drop inline UNIQUE constraint pada nik (warga_nik_key) jika ada
ALTER TABLE warga DROP CONSTRAINT IF EXISTS warga_nik_key;

-- Drop index unik lama pada nik jika ada
DROP INDEX IF EXISTS idx_warga_nik;

-- Buat conditional unique index: hanya untuk nik yang bukan NULL dan status complete
CREATE UNIQUE INDEX IF NOT EXISTS idx_warga_nik_unique
    ON warga(nik)
    WHERE nik IS NOT NULL AND status = 'complete';

-- Index untuk pencarian draft by token
CREATE INDEX IF NOT EXISTS idx_warga_draft_token ON warga(draft_token) WHERE draft_token IS NOT NULL;

-- Index untuk filter by status
CREATE INDEX IF NOT EXISTS idx_warga_status ON warga(status);
