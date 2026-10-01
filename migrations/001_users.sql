-- Tabel users: menyimpan akun login.
-- Role hanya boleh 'admin' atau 'mahasiswa' — dijaga oleh CHECK constraint.
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL      PRIMARY KEY,
    email      VARCHAR(120) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,           -- selalu disimpan dalam bentuk bcrypt hash
    role       VARCHAR(20)  NOT NULL DEFAULT 'mahasiswa'
                            CHECK (role IN ('admin', 'mahasiswa')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index agar pencarian by email (saat login) cepat
CREATE INDEX IF NOT EXISTS idx_users_email ON users (LOWER(email));
