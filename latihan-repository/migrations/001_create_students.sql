CREATE TABLE IF NOT EXISTS students (
    id         SERIAL       PRIMARY KEY,
    nim        VARCHAR(20)  NOT NULL,
    name       VARCHAR(100) NOT NULL,
    grade      DOUBLE PRECISION NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Keunikan NIM tanpa membedakan huruf besar/kecil.
-- Menggantikan pemeriksaan manual di pertemuan 2 yang rawan race condition.
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_lower_key
    ON students (LOWER(nim));

-- Indeks tambahan untuk pencarian nama (ILIKE) agar tidak full scan saat data banyak.
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
