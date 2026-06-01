ALTER TABLE students ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true NOT NULL;

UPDATE students SET is_active = true WHERE is_active IS NULL;

CREATE INDEX IF NOT EXISTS idx_students_is_active ON students(is_active);
