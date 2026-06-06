ALTER TABLE work_shifts 
DROP COLUMN IF EXISTS enforce_lateness,
DROP COLUMN IF EXISTS enforce_lunch_limit,
DROP COLUMN IF EXISTS enforce_geofence;
