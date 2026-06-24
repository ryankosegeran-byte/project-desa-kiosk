-- 008_warga_rfid_unique_ignore_deleted.sql
-- The unique RFID index must ignore soft-deleted rows, so a card freed by a
-- deleted warga can be re-linked to a new warga.

DROP INDEX IF EXISTS idx_warga_rfid;

CREATE UNIQUE INDEX IF NOT EXISTS idx_warga_rfid
    ON warga(rfid_uid)
    WHERE rfid_uid IS NOT NULL AND deleted_at IS NULL;