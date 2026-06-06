-- Up
ALTER TABLE work_shifts 
ADD COLUMN enforce_lateness BOOLEAN NOT NULL DEFAULT true,
ADD COLUMN enforce_lunch_limit BOOLEAN NOT NULL DEFAULT true,
ADD COLUMN enforce_geofence BOOLEAN NOT NULL DEFAULT true;

