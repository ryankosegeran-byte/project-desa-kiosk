-- 003_nomor_surat_batch.sql
-- Tabel untuk menyimpan jatah nomor surat per jenis (offline batch)

CREATE TABLE IF NOT EXISTS nomor_surat_batch (
    jenis_surat_id  TEXT PRIMARY KEY REFERENCES jenis_surat(id),
    nomor_terakhir  INTEGER NOT NULL DEFAULT 0,
    batas_atas      INTEGER NOT NULL DEFAULT 100,
    format_nomor    TEXT DEFAULT '{nomor}/{kode_surat}/{kode_desa}/{bulan_romawi}/{tahun}',
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);
