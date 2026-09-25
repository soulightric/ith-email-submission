-- Jalankan sekali saat setup: psql -U postgres -d monitoring_email -f schema.sql

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          VARCHAR(20) NOT NULL DEFAULT 'admin'
);

CREATE TABLE IF NOT EXISTS email_requests (
    id            SERIAL PRIMARY KEY,
    jenis_usulan  VARCHAR(50)  NOT NULL,           -- Dosen / Pegawai ITH, Mahasiswa ITH, Lembaga ITH
    nama          VARCHAR(150) NOT NULL,
    nip_nim       VARCHAR(50)  NOT NULL,
    prodi_unit    VARCHAR(150) NOT NULL,
    formulir_path TEXT,
    status        VARCHAR(30)  NOT NULL DEFAULT 'Diajukan', -- Diajukan, Diproses, Selesai, Ditolak
    detail        TEXT,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE email_requests ADD COLUMN IF NOT EXISTS formulir_path TEXT;

CREATE INDEX IF NOT EXISTS idx_email_requests_nama    ON email_requests (nama);
CREATE INDEX IF NOT EXISTS idx_email_requests_nip_nim ON email_requests (nip_nim);
CREATE INDEX IF NOT EXISTS idx_email_requests_status  ON email_requests (status);
CREATE INDEX IF NOT EXISTS idx_email_requests_jenis   ON email_requests (jenis_usulan);
CREATE INDEX IF NOT EXISTS idx_email_requests_created ON email_requests (created_at);

-- Contoh membuat akun admin (ganti hash dengan hasil bcrypt yang sebenarnya):
-- INSERT INTO users (username, password_hash, role) VALUES ('admin', '$2a$10$...', 'admin');
