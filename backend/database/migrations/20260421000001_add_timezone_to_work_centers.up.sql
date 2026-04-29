-- Add timezone column to work_centers
ALTER TABLE work_centers ADD COLUMN timezone VARCHAR(50) DEFAULT 'UTC';

-- Optional: Update existing centers to a specific default if known
UPDATE work_centers SET timezone = 'America/Mexico_City' WHERE timezone = 'UTC';
