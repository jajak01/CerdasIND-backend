DROP INDEX IF EXISTS idx_sessions_public_id;
DROP INDEX IF EXISTS idx_students_public_id;
DROP INDEX IF EXISTS idx_bundles_public_id;

ALTER TABLE sessions DROP COLUMN IF EXISTS public_id;
ALTER TABLE students DROP COLUMN IF EXISTS public_id;
ALTER TABLE bundles DROP COLUMN IF EXISTS public_id;
