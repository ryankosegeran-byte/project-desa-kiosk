-- 010_desa_theme.sql
-- Adds kiosk theme selection per village (central, synced to kiosk).
ALTER TABLE desa ADD COLUMN IF NOT EXISTS theme VARCHAR(50) NOT NULL DEFAULT 'merah-putih';