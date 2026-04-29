ALTER TABLE attendances 
DROP CONSTRAINT IF EXISTS attendances_work_center_id_foreign,
DROP COLUMN IF EXISTS work_center_id;

ALTER TABLE work_centers 
DROP CONSTRAINT IF EXISTS work_centers_manager_id_foreign,
DROP COLUMN IF EXISTS manager_id;
