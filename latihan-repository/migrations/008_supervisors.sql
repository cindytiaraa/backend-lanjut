-- Migration 008 (Tugas D.2): pembimbing tugas akhir untuk entity students
-- psql -U postgres -d praktikum_backend -f migrations/008_supervisors.sql
CREATE TABLE IF NOT EXISTS supervisors (
    id               SERIAL       PRIMARY KEY,
    name             VARCHAR(120) NOT NULL UNIQUE,
    quota            INTEGER      NOT NULL CHECK (quota > 0),
    supervised_count INTEGER      NOT NULL DEFAULT 0,
    version          INTEGER      NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- Jaring pengaman: tidak pernah terpicu bila kode aplikasi benar.
    CONSTRAINT supervisors_supervised_within_quota
        CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

CREATE TABLE IF NOT EXISTS supervisions (
    id            BIGSERIAL   PRIMARY KEY,
    -- UNIQUE pada student_id: SATU mahasiswa hanya boleh punya SATU
    -- pembimbing. Dijaga database, bukan hanya kode aplikasi.
    student_id    INTEGER     NOT NULL UNIQUE REFERENCES students(id)    ON DELETE CASCADE,
    supervisor_id INTEGER     NOT NULL        REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS supervisions_supervisor_idx ON supervisions (supervisor_id);

-- supervised_count adalah turunan dari baris supervisions. Bila baris
-- supervisions hilang karena ON DELETE CASCADE (mahasiswa dihapus),
-- penghitung harus ikut turun. Pada alur normal (assign / switch) trigger
-- ini tidak berperan: switch memakai UPDATE, bukan DELETE.
CREATE OR REPLACE FUNCTION supervisions_after_delete() RETURNS trigger AS $$
BEGIN
    UPDATE supervisors
       SET supervised_count = supervised_count - 1
     WHERE id = OLD.supervisor_id AND supervised_count > 0;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS supervisions_after_delete_trg ON supervisions;
CREATE TRIGGER supervisions_after_delete_trg
    AFTER DELETE ON supervisions
    FOR EACH ROW EXECUTE FUNCTION supervisions_after_delete();

INSERT INTO supervisors (name, quota) VALUES
    ('Pembimbing Rebutan', 3),
    ('Pembimbing Alfa',    10),
    ('Pembimbing Beta',    10)
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, description) VALUES
    ('supervision:manage', 'Menetapkan, mengganti pembimbing, dan mengubah kuota pembimbing')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'supervision:manage')
ON CONFLICT DO NOTHING;
