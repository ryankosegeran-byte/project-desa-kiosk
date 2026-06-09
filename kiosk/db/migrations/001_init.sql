-- 001_init.sql

-- ==========================================
-- Metadata kiosk
-- ==========================================
CREATE TABLE IF NOT EXISTS kiosk_config (
    key         TEXT PRIMARY KEY,
    value       TEXT NOT NULL
);

-- ==========================================
-- Jenis surat (synced dari server, hanya yang aktif untuk desa ini)
-- ==========================================
CREATE TABLE IF NOT EXISTS jenis_surat (
    id              TEXT PRIMARY KEY,          -- UUID
    kode            TEXT UNIQUE NOT NULL,      -- 'SK_DOMISILI', 'SKTM', etc.
    nama            TEXT NOT NULL,             -- 'Surat Keterangan Domisili'
    deskripsi       TEXT,
    fields_schema   TEXT NOT NULL,             -- JSON: definisi field form
    aktif           BOOLEAN DEFAULT 1,         -- aktif untuk desa ini
    urutan          INTEGER DEFAULT 0,         -- urutan tampil di kiosk
    updated_at      DATETIME
);

-- ==========================================
-- Template surat per jenis (synced dari server, future: editable)
-- ==========================================
CREATE TABLE IF NOT EXISTS surat_template (
    id              TEXT PRIMARY KEY,          -- UUID
    jenis_surat_id  TEXT NOT NULL REFERENCES jenis_surat(id),
    desa_id         TEXT NOT NULL,
    template_html   TEXT NOT NULL,             -- HTML template dengan placeholder
    version         INTEGER DEFAULT 1,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_template_jenis_desa ON surat_template(jenis_surat_id, desa_id);

-- ==========================================
-- Data warga (synced dari server)
-- ==========================================
CREATE TABLE IF NOT EXISTS warga (
    id              TEXT PRIMARY KEY,          -- UUID
    nik             TEXT UNIQUE NOT NULL,
    rfid_uid        TEXT,                      -- UID dari chip RFID KTP-el (hex string)
    nama            TEXT NOT NULL,
    tempat_lahir    TEXT,
    tanggal_lahir   DATE,
    jenis_kelamin   TEXT CHECK(jenis_kelamin IN ('L', 'P')),
    alamat          TEXT,
    rt              TEXT,
    rw              TEXT,
    kelurahan       TEXT,
    kecamatan       TEXT,
    kabupaten       TEXT,
    provinsi        TEXT,
    agama           TEXT,
    status_kawin    TEXT,
    pekerjaan       TEXT,
    kewarganegaraan TEXT DEFAULT 'WNI',
    desa_id         TEXT NOT NULL,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    synced_at       DATETIME
);

CREATE INDEX IF NOT EXISTS idx_warga_nik ON warga(nik);
CREATE INDEX IF NOT EXISTS idx_warga_rfid ON warga(rfid_uid);          -- lookup by RFID scan
CREATE INDEX IF NOT EXISTS idx_warga_desa ON warga(desa_id);
CREATE INDEX IF NOT EXISTS idx_warga_nama ON warga(nama);

-- ==========================================
-- Surat yang dibuat di kiosk
-- ==========================================
CREATE TABLE IF NOT EXISTS surat (
    id              TEXT PRIMARY KEY,          -- UUID (dibuat di kiosk)
    nomor_surat     TEXT,                      -- auto-generated
    jenis_surat_id  TEXT NOT NULL REFERENCES jenis_surat(id),
    jenis_surat_kode TEXT NOT NULL,            -- denormalized
    jenis_surat_nama TEXT NOT NULL,            -- denormalized
    warga_id        TEXT REFERENCES warga(id),
    nik_pemohon     TEXT NOT NULL,
    nama_pemohon    TEXT NOT NULL,
    data_surat      TEXT NOT NULL,             -- JSON: field-field spesifik per jenis surat
    status          TEXT DEFAULT 'DRAFT' CHECK(status IN ('DRAFT','PRINTED','SYNCED','FAILED')),
    pdf_path        TEXT,
    desa_id         TEXT NOT NULL,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    printed_at      DATETIME,
    synced          BOOLEAN DEFAULT 0,
    synced_at       DATETIME
);

CREATE INDEX IF NOT EXISTS idx_surat_desa ON surat(desa_id);
CREATE INDEX IF NOT EXISTS idx_surat_synced ON surat(synced);
CREATE INDEX IF NOT EXISTS idx_surat_nik ON surat(nik_pemohon);
CREATE INDEX IF NOT EXISTS idx_surat_jenis ON surat(jenis_surat_id);
CREATE INDEX IF NOT EXISTS idx_surat_tanggal ON surat(created_at);

-- ==========================================
-- Sync queue
-- ==========================================
CREATE TABLE IF NOT EXISTS sync_queue (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type     TEXT NOT NULL,             -- 'surat'
    entity_id       TEXT NOT NULL,
    operation       TEXT NOT NULL,             -- 'CREATE', 'UPDATE'
    payload         TEXT NOT NULL,             -- JSON
    attempts        INTEGER DEFAULT 0,
    max_attempts    INTEGER DEFAULT 5,
    last_error      TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    processed_at    DATETIME
);

CREATE INDEX IF NOT EXISTS idx_sync_pending ON sync_queue(processed_at) WHERE processed_at IS NULL;

-- ==========================================
-- Activity log
-- ==========================================
CREATE TABLE IF NOT EXISTS activity_log (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    action          TEXT NOT NULL,
    detail          TEXT,                      -- JSON
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);
