-- Add lunch_overstay to incident_type enum
ALTER TYPE incident_type ADD VALUE IF NOT EXISTS 'lunch_overstay';

-- Ensure all existing incidents have a status
UPDATE incidents SET status = 'pending' WHERE status IS NULL;
