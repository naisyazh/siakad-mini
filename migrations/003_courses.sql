-- Tabel courses: katalog mata kuliah yang tersedia.
-- kuota = batas maksimal mahasiswa yang bisa mengambil matkul ini.
CREATE TABLE IF NOT EXISTS courses (
    id        SERIAL       PRIMARY KEY,
    kode_mk   VARCHAR(20)  NOT NULL UNIQUE,
    nama_mk   VARCHAR(100) NOT NULL,
    sks       SMALLINT     NOT NULL CHECK (sks >= 1 AND sks <= 6),
    semester  SMALLINT     NOT NULL CHECK (semester >= 1 AND semester <= 8),
    kuota     SMALLINT     NOT NULL DEFAULT 30 CHECK (kuota >= 0)
);

-- Index untuk pencarian by semester dan kode_mk
CREATE INDEX IF NOT EXISTS idx_courses_semester ON courses (semester);
