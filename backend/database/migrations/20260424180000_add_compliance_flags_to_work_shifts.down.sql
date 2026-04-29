-- Down
ALTER TABLE work_shifts 
DROP COLUMN enforce_lateness,
DROP COLUMN enforce_lunch_limit,
DROP COLUMN enforce_geofence;
