-- 009_warga_nik_unique_ignore_deleted.sql
-- The unique NIK index must ignore soft-deleted rows, so a NIK freed by a
-- deleted warga can be re-registered.

DROP INDEX IF EXISTS idx_warga_nik_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_warga_nik_unique
    ON warga(nik)
    WHERE nik IS NOT NULL AND status = 'complete' AND deleted_at IS NULL;