-- Migration 010: permission moderasi ulasan
-- psql -U postgres -d praktikum_backend -f migrations/010_review_permissions.sql
--
-- Ulasannya sendiri disimpan di MongoDB, tetapi hak aksesnya tetap dikelola
-- di PostgreSQL bersama permission lain. Satu sistem hak akses untuk seluruh
-- aplikasi, apa pun database tempat datanya disimpan.
-- (Di modul bernomor 007; di proyek ini 007-009 sudah dipakai, maka 010.)
INSERT INTO permissions (name, description) VALUES
    ('review:moderate', 'Menghapus ulasan milik siapa pun')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'review:moderate')
ON CONFLICT DO NOTHING;
