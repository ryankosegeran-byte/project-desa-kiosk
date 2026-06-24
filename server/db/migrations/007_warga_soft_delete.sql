-- 007_warga_soft_delete.sql
-- Add soft-delete support to the warga table.

ALTER TABLE warga ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_warga_deleted ON warga(deleted_at) WHERE deleted_at IS NOT NULL;