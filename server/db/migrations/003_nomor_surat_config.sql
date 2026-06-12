-- 003_nomor_surat_config.sql
CREATE TABLE IF NOT EXISTS nomor_surat_config (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    desa_id         UUID NOT NULL REFERENCES desa(id),
    jenis_surat_id  UUID NOT NULL REFERENCES jenis_surat(id),
    nomor_mulai     INTEGER NOT NULL DEFAULT 1,
    batas_atas      INTEGER NOT NULL DEFAULT 100,
    nomor_terakhir  INTEGER NOT NULL DEFAULT 0,
    format_nomor    TEXT DEFAULT '{nomor}/{kode_surat}/{kode_desa}/{bulan_romawi}/{tahun}',
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(desa_id, jenis_surat_id)
);
