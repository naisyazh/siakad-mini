-- Tabel students: data profil mahasiswa.
-- user_id adalah FK ke users — satu user hanya boleh punya satu student (1-1).
-- deleted_at dipakai untuk soft delete: baris tidak dihapus, hanya ditandai waktu hapus.
CREATE TABLE IF NOT EXISTS students (
    id           SERIAL        PRIMARY KEY,
    user_id      INT           NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nim          VARCHAR(12)   NOT NULL UNIQUE,
    nama         VARCHAR(100)  NOT NULL,
    prodi        VARCHAR(100)  NOT NULL,
    angkatan     SMALLINT      NOT NULL CHECK (angkatan >= 2000),
    ipk_terakhir NUMERIC(3,2)  NOT NULL DEFAULT 0.00
                               CHECK (ipk_terakhir >= 0.00 AND ipk_terakhir <= 4.00),
    deleted_at   TIMESTAMPTZ   NULL      -- NULL = aktif, terisi = sudah dihapus (soft delete)
);

-- Index agar filter by prodi dan angkatan cepat
CREATE INDEX IF NOT EXISTS idx_students_prodi    ON students (prodi);
CREATE INDEX IF NOT EXISTS idx_students_angkatan ON students (angkatan);
-- Index agar soft-delete query efisien (sering filter WHERE deleted_at IS NULL)
CREATE INDEX IF NOT EXISTS idx_students_deleted  ON students (deleted_at) WHERE deleted_at IS NULL;
