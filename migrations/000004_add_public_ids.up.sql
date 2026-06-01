CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE bundles ADD COLUMN IF NOT EXISTS public_id UUID;
ALTER TABLE students ADD COLUMN IF NOT EXISTS public_id UUID;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS public_id UUID;

UPDATE bundles SET public_id = gen_random_uuid() WHERE public_id IS NULL;
UPDATE students SET public_id = gen_random_uuid() WHERE public_id IS NULL;
UPDATE sessions SET public_id = gen_random_uuid() WHERE public_id IS NULL;

ALTER TABLE bundles ALTER COLUMN public_id SET DEFAULT gen_random_uuid();
ALTER TABLE students ALTER COLUMN public_id SET DEFAULT gen_random_uuid();
ALTER TABLE sessions ALTER COLUMN public_id SET DEFAULT gen_random_uuid();

ALTER TABLE bundles ALTER COLUMN public_id SET NOT NULL;
ALTER TABLE students ALTER COLUMN public_id SET NOT NULL;
ALTER TABLE sessions ALTER COLUMN public_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_bundles_public_id ON bundles(public_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_students_public_id ON students(public_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_public_id ON sessions(public_id);
