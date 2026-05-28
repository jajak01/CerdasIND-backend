-- ====================================================================
-- 1. MEMBUAT TIPE DATA CUSTOM (ENUM)
-- ====================================================================
CREATE TYPE user_role AS ENUM ('admin', 'peserta');
CREATE TYPE jenis_soal AS ENUM ('pilihan_ganda', 'isian_singkat');
CREATE TYPE status_ujian AS ENUM ('berlangsung', 'menunggu_koreksi', 'selesai');

-- ====================================================================
-- 2. MEMBUAT TABEL UTAMA & RELASI
-- ====================================================================

-- Tabel Users (Menampung Admin & Peserta dalam satu pintu login)
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role DEFAULT 'peserta' NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel Jenjang (Kategori Utama: SD, SMP, SMA)
CREATE TABLE jenjang (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(10) UNIQUE NOT NULL
);

-- Tabel Mata Pelajaran (Berelasi dengan Jenjang, e.g., Matematika di SMA)
CREATE TABLE mapel (
    id SERIAL PRIMARY KEY,
    jenjang_id INT NOT NULL REFERENCES jenjang(id) ON DELETE CASCADE,
    nama VARCHAR(50) NOT NULL,
    UNIQUE(jenjang_id, nama) -- Mencegah duplikasi nama mapel di jenjang yang sama
);

-- Tabel Bundles (Sampul/Paket Ujian yang di-upload via Excel)
CREATE TABLE bundles (
    id BIGSERIAL PRIMARY KEY,
    mapel_id INT NOT NULL REFERENCES mapel(id) ON DELETE CASCADE,
    nama_bundle VARCHAR(150) NOT NULL,
    deskripsi TEXT,
    waktu_menit INT DEFAULT 60 NOT NULL,
    is_active BOOLEAN DEFAULT false NOT NULL,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel Soal (Menyimpan ribuan isi pertanyaan dari Excel)
CREATE TABLE soal (
    id BIGSERIAL PRIMARY KEY,
    bundle_id BIGINT NOT NULL REFERENCES bundles(id) ON DELETE CASCADE,
    tipe_soal jenis_soal NOT NULL DEFAULT 'pilihan_ganda',
    teks_soal TEXT NOT NULL,
    pilihan_jawaban JSONB, -- Berisi array pilihan A, B, C, D jika PG. NULL jika Isian.
    kunci_jawaban TEXT NOT NULL,
    pembahasan TEXT,
    bobot_nilai INT DEFAULT 1 NOT NULL, -- Poin maksimal jika soal ini benar
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel History Ujian (Lembar kerja dan rekam jejak nilai siswa)
CREATE TABLE history_ujian (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bundle_id BIGINT NOT NULL REFERENCES bundles(id) ON DELETE CASCADE,
    waktu_mulai TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    waktu_selesai TIMESTAMP,
    skor_akhir NUMERIC(5,2) DEFAULT 0.00, -- Nilai total (bisa desimal, e.g., 87.50)
    detail_jawaban JSONB, -- Menyimpan histori jawaban siswa per nomor soal
    status status_ujian DEFAULT 'berlangsung' NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ====================================================================
-- 3. MEMBUAT DATABASE INDEX (Untuk Optimasi Performa Kecepatan API)
-- ====================================================================
CREATE INDEX idx_mapel_jenjang ON mapel(jenjang_id);
CREATE INDEX idx_bundles_mapel ON bundles(mapel_id);
CREATE INDEX idx_soal_bundle ON soal(bundle_id);
CREATE INDEX idx_history_user ON history_ujian(user_id);
CREATE INDEX idx_history_bundle ON history_ujian(bundle_id);
CREATE INDEX idx_history_status ON history_ujian(status);

-- DATA AWAL ESENSI
-- Insert Kategori Jenjang
INSERT INTO jenjang (nama) VALUES ('SD'), ('SMP'), ('SMA');

-- Insert Mata Pelajaran SD (Terikat ke jenjang_id = 1)
INSERT INTO mapel (jenjang_id, nama) VALUES 
(1, 'Bahasa Indonesia'), 
(1, 'Bahasa Inggris'), 
(1, 'Matematika'), 
(1, 'IPA');

-- Insert Mata Pelajaran SMP (Terikat ke jenjang_id = 2)
INSERT INTO mapel (jenjang_id, nama) VALUES 
(2, 'Bahasa Indonesia'), 
(2, 'Bahasa Inggris'), 
(2, 'Matematika'), 
(2, 'IPA');

-- Insert Mata Pelajaran SMA (Terikat ke jenjang_id = 3)
INSERT INTO mapel (jenjang_id, nama) VALUES 
(3, 'Bahasa Indonesia'), 
(3, 'Bahasa Inggris'), 
(3, 'Matematika'), 
(3, 'Biologi'), 
(3, 'Fisika'), 
(3, 'Kimia');
