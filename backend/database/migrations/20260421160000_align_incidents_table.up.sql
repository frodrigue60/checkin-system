-- Align incidents table with the backend code (models and handlers)
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS metadata_json JSONB;

-- Optionally remove old columns if they are not used, but let's keep them for now to avoid data loss if any exists.
-- ALTER TABLE incidents DROP COLUMN IF EXISTS description;
-- ALTER TABLE incidents DROP COLUMN IF EXISTS is_late;
-- ALTER TABLE incidents DROP COLUMN IF EXISTS delay_minutes;
-- ALTER TABLE incidents DROP COLUMN IF EXISTS is_out_of_range;
-- ALTER TABLE incidents DROP COLUMN IF EXISTS distance;
-- ALTER TABLE incidents DROP COLUMN IF EXISTS check_in_time;
