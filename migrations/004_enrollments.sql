-- Tabel enrollments: KRS — mencatat mahasiswa mengambil mata kuliah.
-- Constraint UNIQUE (student_id, course_id, tahun_akademik) mencegah mahasiswa
-- mengambil matkul yang sama dua kali pada tahun akademik yang sama.
CREATE TABLE IF NOT EXISTS enrollments (
    id               SERIAL       PRIMARY KEY,
    student_id       INT          NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id        INT          NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
    tahun_akademik   VARCHAR(20)  NOT NULL, -- contoh: "2026/2027-Ganjil"
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- Ini yang mencegah duplikasi KRS (business rule #2)
    CONSTRAINT uq_enrollment UNIQUE (student_id, course_id, tahun_akademik)
);

-- Index untuk query "ambil semua KRS milik mahasiswa X"
CREATE INDEX IF NOT EXISTS idx_enrollments_student ON enrollments (student_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_course  ON enrollments (course_id);
