-- Seed 250 pengajuan untuk pengujian dashboard dan counter.
-- Jalankan setelah schema.sql pada database monitoring_email.
BEGIN;

INSERT INTO email_requests (
    jenis_usulan,
    nama,
    nip_nim,
    prodi_unit,
    formulir_path,
    status,
    detail,
    created_at,
    updated_at
)
SELECT
    CASE (n - 1) % 3
        WHEN 0 THEN 'Dosen / Pegawai ITH'
        WHEN 1 THEN 'Mahasiswa ITH'
        ELSE 'Lembaga ITH'
    END,
    'Pengguna Uji ' || LPAD(n::text, 3, '0'),
    'SEED250' || LPAD(n::text, 3, '0'),
    CASE (n - 1) % 4
        WHEN 0 THEN 'Informatika'
        WHEN 1 THEN 'Sistem Informasi'
        WHEN 2 THEN 'Administrasi Akademik'
        ELSE 'Unit Pengujian'
    END,
    NULL,
    CASE (n - 1) % 4
        WHEN 0 THEN 'Diajukan'
        WHEN 1 THEN 'Diproses'
        WHEN 2 THEN 'Selesai'
        ELSE 'Ditolak'
    END,
    'SEED-TEST-500 - data pengujian nomor ' || n,
    NOW() - ((500 - n) || ' minutes')::interval,
    NOW() - ((500 - n) || ' minutes')::interval
FROM generate_series(1, 500) AS series(n);

COMMIT;

-- Verifikasi jumlah seed:
-- SELECT COUNT(*) FROM email_requests WHERE detail LIKE 'SEED-TEST-500%';

-- Hapus data seed setelah pengujian:
-- DELETE FROM email_requests WHERE detail LIKE 'SEED-TEST-500%';
