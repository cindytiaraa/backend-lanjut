-- Migration 009: jaga enrolled_count untuk semua penghapusan enrollment.
-- Jalankan setelah migration 006; aman dijalankan ulang.
UPDATE courses c
    SET enrolled_count = (
         SELECT COUNT(*)
            FROM enrollments e
          WHERE e.course_id = c.id
    );

CREATE OR REPLACE FUNCTION enrollments_after_delete_update_course_count() RETURNS trigger AS $$
BEGIN
    UPDATE courses
       SET enrolled_count = enrolled_count - 1
     WHERE id = OLD.course_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS enrollments_after_delete_update_course_count_trg ON enrollments;
CREATE TRIGGER enrollments_after_delete_update_course_count_trg
    AFTER DELETE ON enrollments
    FOR EACH ROW EXECUTE FUNCTION enrollments_after_delete_update_course_count();