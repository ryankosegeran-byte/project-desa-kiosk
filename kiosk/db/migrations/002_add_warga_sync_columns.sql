-- 002_add_warga_sync_columns.sql
-- Adds columns synced from server for draft registration support

ALTER TABLE warga ADD COLUMN foto_ktp_path TEXT;
ALTER TABLE warga ADD COLUMN status TEXT DEFAULT 'complete';
ALTER TABLE warga ADD COLUMN draft_token TEXT;
