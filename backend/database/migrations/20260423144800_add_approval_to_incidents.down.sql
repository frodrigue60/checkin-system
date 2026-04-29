ALTER TABLE incidents DROP COLUMN IF EXISTS status;
ALTER TABLE incidents DROP COLUMN IF EXISTS resolved_by;
ALTER TABLE incidents DROP COLUMN IF EXISTS resolution_note;
