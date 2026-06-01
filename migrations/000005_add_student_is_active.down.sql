DROP INDEX IF EXISTS idx_students_is_active;
ALTER TABLE students DROP COLUMN IF EXISTS is_active;
