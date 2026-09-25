-- Urutan column pada index HARUS sama persis dengan ORDER BY pada query
-- keyset pagination di student_repository.go, termasuk arah DESC-nya.
-- Bila tidak sama, Postgres tidak bisa langsung melompat ke posisi cursor
-- dan seluruh keuntungan keyset pagination hilang tanpa pesan kesalahan.
CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
    ON students (created_at DESC, id DESC);
