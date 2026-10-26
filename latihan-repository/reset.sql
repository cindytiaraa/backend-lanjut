-- Reset keadaan untuk uji beban. Jalankan SEBELUM setiap uji:
--   psql -U postgres -d praktikum_backend -f reset.sql
DELETE FROM enrollments
 WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%');

UPDATE courses c
   SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id);

-- Bagian pembimbing (Tugas D.2/D.3): hanya mahasiswa buatan alat uji beban.
DELETE FROM supervisions
 WHERE student_id IN (SELECT id FROM students WHERE nim LIKE '2099%');

UPDATE supervisors s
   SET supervised_count = (SELECT COUNT(*) FROM supervisions v WHERE v.supervisor_id = s.id);
