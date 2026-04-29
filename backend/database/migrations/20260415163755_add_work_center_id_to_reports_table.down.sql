ALTER TABLE reports 
DROP CONSTRAINT IF EXISTS reports_work_center_id_foreign,
DROP COLUMN IF EXISTS work_center_id;
