CREATE TABLE IF NOT EXISTS roles (
    name        VARCHAR(20) PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
('admin', 'Akses penuh terhadap seluruh data dan pengaturan'),
('staff', 'Boleh melihat data seluruh user, tetapi tidak boleh mengubah'),
('user', 'Hanya boleh mengelola datanya sendiri')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

INSERT INTO permissions (name, description) VALUES
('student:list', 'Melihat daftar seluruh student'),
('student:read:any', 'Melihat data student mana pun'),
('student:create', 'Membuat data student'),
('student:update:any', 'Mengubah data student mana pun'),
('student:delete', 'Menghapus data student')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO role_permissions (role_name, permission_name) VALUES
('admin', 'student:list'),
('admin', 'student:read:any'),
('admin', 'student:create'),
('admin', 'student:update:any'),
('admin', 'student:delete'),
('staff', 'student:list'),
('staff', 'student:read:any'),
('staff', 'student:create')
ON CONFLICT DO NOTHING;

-- Kunci role user agar hanya role yang dikenal yang bisa masuk.
UPDATE users SET role = 'user' WHERE role NOT IN (SELECT name FROM roles);
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey
    FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;
CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);

-- owner_id untuk ownership student. Baris lama diberi owner user pertama.
ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id INTEGER;

UPDATE students
SET owner_id = (SELECT id FROM users ORDER BY id LIMIT 1)
WHERE owner_id IS NULL;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM students WHERE owner_id IS NULL) THEN
        RAISE EXCEPTION 'students.owner_id masih NULL; buat minimal satu user atau isi owner_id data lama terlebih dahulu';
    END IF;
END $$;

ALTER TABLE students DROP CONSTRAINT IF EXISTS students_owner_id_fkey;
ALTER TABLE students
    ADD CONSTRAINT students_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES users(id);
ALTER TABLE students ALTER COLUMN owner_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students (owner_id);
