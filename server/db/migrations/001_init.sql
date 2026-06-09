-- 001_init.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==========================================
-- Desa
-- ==========================================
CREATE TABLE IF NOT EXISTS desa (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nama            VARCHAR(255) NOT NULL,
    kode_desa       VARCHAR(20) UNIQUE NOT NULL,
    kecamatan       VARCHAR(255),
    kabupaten       VARCHAR(255),
    provinsi        VARCHAR(255),
    kepala_desa     VARCHAR(255),
    nip_kepala_desa VARCHAR(50),
    alamat_kantor   TEXT,
    logo_path       VARCHAR(500),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ==========================================
-- Users
-- ==========================================
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username        VARCHAR(100) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    nama            VARCHAR(255) NOT NULL,
    role            VARCHAR(20) NOT NULL CHECK(role IN ('superadmin', 'pic_desa', 'kiosk')),
    jabatan         VARCHAR(100),              -- custom title: 'Sekretaris Desa', 'Operator Kiosk', etc.
    desa_id         UUID REFERENCES desa(id),  -- NULL for superadmin
    active          BOOLEAN DEFAULT true,
    last_login_at   TIMESTAMPTZ,               -- track last login
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ==========================================
-- Activity log (untuk monitoring PIC oleh superadmin)
-- ==========================================
CREATE TABLE IF NOT EXISTS user_activity_log (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id),
    desa_id         UUID REFERENCES desa(id),
    action          VARCHAR(100) NOT NULL,     -- 'LOGIN', 'REGISTER_WARGA', 'OCR_KTP', etc.
    entity_type     VARCHAR(50),               -- 'warga', 'surat', etc.
    entity_id       UUID,
    detail          JSONB,                     -- extra context
    ip_address      VARCHAR(45),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_activity_user ON user_activity_log(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_desa ON user_activity_log(desa_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_date ON user_activity_log(created_at);

-- ==========================================
-- Jenis Surat (master data)
-- ==========================================
CREATE TABLE IF NOT EXISTS jenis_surat (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    kode            VARCHAR(50) UNIQUE NOT NULL,
    nama            VARCHAR(255) NOT NULL,
    deskripsi       TEXT,
    fields_schema   JSONB NOT NULL,            -- definisi field form
    aktif_global    BOOLEAN DEFAULT true,
    urutan          INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Aktivasi jenis surat per desa
CREATE TABLE IF NOT EXISTS desa_jenis_surat (
    desa_id         UUID NOT NULL REFERENCES desa(id) ON DELETE CASCADE,
    jenis_surat_id  UUID NOT NULL REFERENCES jenis_surat(id) ON DELETE CASCADE,
    aktif           BOOLEAN DEFAULT true,
    urutan          INTEGER DEFAULT 0,
    PRIMARY KEY (desa_id, jenis_surat_id)
);

-- ==========================================
-- Template surat per desa per jenis
-- ==========================================
CREATE TABLE IF NOT EXISTS surat_template (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    jenis_surat_id  UUID NOT NULL REFERENCES jenis_surat(id),
    desa_id         UUID NOT NULL REFERENCES desa(id),
    template_html   TEXT NOT NULL,
    version         INTEGER DEFAULT 1,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(jenis_surat_id, desa_id)
);

-- ==========================================
-- OCR Provider Config
-- ==========================================
CREATE TABLE IF NOT EXISTS ocr_providers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider        VARCHAR(50) NOT NULL,       -- 'mistral', 'gemini', 'groq'
    display_name    VARCHAR(255) NOT NULL,
    api_key_enc     TEXT NOT NULL,               -- encrypted
    model_name      VARCHAR(255),
    priority        INTEGER DEFAULT 0,
    aktif           BOOLEAN DEFAULT true,
    strategy        VARCHAR(20) DEFAULT 'failover',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ==========================================
-- Kiosk registry
-- ==========================================
CREATE TABLE IF NOT EXISTS kiosks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    desa_id         UUID NOT NULL REFERENCES desa(id),
    nama            VARCHAR(255),
    api_key         VARCHAR(255) UNIQUE NOT NULL,
    last_seen_at    TIMESTAMPTZ,
    last_sync_at    TIMESTAMPTZ,
    status          VARCHAR(20) DEFAULT 'active',
    ip_address      VARCHAR(45),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ==========================================
-- Warga
-- ==========================================
CREATE TABLE IF NOT EXISTS warga (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nik             VARCHAR(16) UNIQUE NOT NULL,
    rfid_uid        VARCHAR(100),               -- UID chip RFID KTP-el (hex)
    nama            VARCHAR(255) NOT NULL,
    tempat_lahir    VARCHAR(255),
    tanggal_lahir   DATE,
    jenis_kelamin   CHAR(1) CHECK(jenis_kelamin IN ('L', 'P')),
    alamat          TEXT,
    rt              VARCHAR(5),
    rw              VARCHAR(5),
    kelurahan       VARCHAR(255),
    kecamatan       VARCHAR(255),
    kabupaten       VARCHAR(255),
    provinsi        VARCHAR(255),
    agama           VARCHAR(50),
    status_kawin    VARCHAR(50),
    pekerjaan       VARCHAR(255),
    kewarganegaraan VARCHAR(10) DEFAULT 'WNI',
    foto_ktp_path   VARCHAR(500),               -- path ke foto KTP (disimpan di server)
    desa_id         UUID NOT NULL REFERENCES desa(id),
    registered_by   UUID REFERENCES users(id),  -- siapa yang registrasi
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_warga_nik ON warga(nik);
CREATE UNIQUE INDEX IF NOT EXISTS idx_warga_rfid ON warga(rfid_uid) WHERE rfid_uid IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_warga_desa ON warga(desa_id);
CREATE INDEX IF NOT EXISTS idx_warga_nama ON warga(nama);

-- ==========================================
-- Surat (synced dari kiosk)
-- ==========================================
CREATE TABLE IF NOT EXISTS surat (
    id              UUID PRIMARY KEY,
    nomor_surat     VARCHAR(100),
    jenis_surat_id  UUID NOT NULL REFERENCES jenis_surat(id),
    warga_id        UUID REFERENCES warga(id),
    nik_pemohon     VARCHAR(16) NOT NULL,
    nama_pemohon    VARCHAR(255) NOT NULL,
    data_surat      JSONB NOT NULL,
    status          VARCHAR(20) DEFAULT 'PRINTED',
    desa_id         UUID NOT NULL REFERENCES desa(id),
    kiosk_id        UUID REFERENCES kiosks(id),
    kiosk_created_at TIMESTAMPTZ NOT NULL,
    synced_at       TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_surat_desa ON surat(desa_id);
CREATE INDEX IF NOT EXISTS idx_surat_jenis ON surat(jenis_surat_id);
CREATE INDEX IF NOT EXISTS idx_surat_tanggal ON surat(kiosk_created_at);

-- ==========================================
-- Sync log
-- ==========================================
CREATE TABLE IF NOT EXISTS sync_log (
    id              BIGSERIAL PRIMARY KEY,
    kiosk_id        UUID REFERENCES kiosks(id),
    direction       VARCHAR(10) NOT NULL,
    entity_type     VARCHAR(50),
    entity_count    INTEGER,
    status          VARCHAR(20),
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
