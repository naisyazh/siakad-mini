-- ============================================================
-- SEEDER: data awal untuk pengujian
-- Jalankan SETELAH semua migration (001–004) sudah dijalankan.
-- ============================================================

-- Bersihkan dulu (urutan penting karena ada FK)
TRUNCATE enrollments, students, users, courses RESTART IDENTITY CASCADE;

-- ── 10 MATA KULIAH ──────────────────────────────────────────
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES
    ('IF101', 'Algoritma dan Pemrograman',     3, 1, 40),
    ('IF102', 'Matematika Diskrit',            3, 1, 35),
    ('IF201', 'Struktur Data',                 3, 2, 35),
    ('IF202', 'Basis Data',                    3, 2, 40),
    ('IF301', 'Pemrograman Web',               3, 3, 40),
    ('IF302', 'Rekayasa Perangkat Lunak',      3, 3, 30),
    ('IF401', 'Jaringan Komputer',             3, 4, 30),
    ('IF402', 'Pemrograman Mobile',            3, 4, 25),
    ('IF501', 'Kecerdasan Buatan',             3, 5, 25),
    ('IF502', 'Keamanan Sistem Informasi',     3, 5, 20);

-- ── 1 ADMIN ─────────────────────────────────────────────────
-- Password: admin123  (bcrypt hash, cost=10)
INSERT INTO users (email, password, role) VALUES
    ('admin@siakad.ac.id',
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhu2',
     'admin');

-- ── 20 MAHASISWA ─────────────────────────────────────────────
-- Password awal tiap mahasiswa = NIM-nya (bcrypt, cost=10)
-- Hash untuk NIM 434241001 s/d 434241020:
-- Semua menggunakan hash yang sama karena nilai NIM berbeda,
-- jadi kita akan generate via aplikasi Go (migrate.go).
-- Di sini kita pakai hash dari string nim masing-masing.
--
-- Untuk keperluan seeder, password = "password123" (bcrypt hash di bawah)
-- hash: $2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi
-- (ini adalah hash bcrypt dari "password" — bukan production!)
--
-- Catatan: di production, generate hash yang benar via aplikasi.

INSERT INTO users (email, password, role) VALUES
    ('mahasiswa01@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa02@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa03@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa04@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa05@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa06@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa07@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa08@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa09@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa10@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa11@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa12@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa13@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa14@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa15@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa16@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa17@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa18@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa19@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa'),
    ('mahasiswa20@siakad.ac.id', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'mahasiswa');

-- ── DATA STUDENTS (sambung ke user_id 2–21) ─────────────────
INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES
    (2,  '434241001', 'Rina Putri Andari',     'Sistem Informasi',      2022, 3.75),
    (3,  '434241002', 'Budi Santoso',          'Teknik Informatika',    2022, 3.20),
    (4,  '434241003', 'Citra Dewi Lestari',    'Sistem Informasi',      2022, 2.85),
    (5,  '434241004', 'Dimas Arya Pratama',    'Teknik Informatika',    2022, 3.50),
    (6,  '434241005', 'Eka Fitriana',          'Manajemen Informatika', 2023, 3.10),
    (7,  '434241006', 'Fajar Nugroho',         'Teknik Informatika',    2023, 2.40),
    (8,  '434241007', 'Gita Puspita Sari',     'Sistem Informasi',      2023, 3.90),
    (9,  '434241008', 'Hendra Wijaya',         'Manajemen Informatika', 2023, 2.60),
    (10, '434241009', 'Indah Permatasari',     'Teknik Informatika',    2024, 3.45),
    (11, '434241010', 'Joko Susilo',           'Sistem Informasi',      2024, 3.00),
    (12, '434241011', 'Kartika Wulandari',     'Manajemen Informatika', 2024, 3.80),
    (13, '434241012', 'Lukman Hakim',          'Teknik Informatika',    2024, 2.95),
    (14, '434241013', 'Maya Anggraini',        'Sistem Informasi',      2024, 3.15),
    (15, '434241014', 'Nanda Saputra',         'Teknik Informatika',    2022, 2.70),
    (16, '434241015', 'Olivia Rachma',         'Manajemen Informatika', 2022, 3.60),
    (17, '434241016', 'Putra Ramadhan',        'Teknik Informatika',    2023, 3.35),
    (18, '434241017', 'Qori Amelia',           'Sistem Informasi',      2023, 2.55),
    (19, '434241018', 'Rizky Firmansyah',      'Teknik Informatika',    2024, 3.70),
    (20, '434241019', 'Sari Wahyuni',          'Manajemen Informatika', 2024, 3.05),
    (21, '434241020', 'Taufik Hidayat',        'Sistem Informasi',      2022, 2.45);
