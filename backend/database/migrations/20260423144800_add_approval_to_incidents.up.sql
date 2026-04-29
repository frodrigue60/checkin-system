ALTER TABLE incidents ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending';
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS resolved_by BIGINT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS resolution_note TEXT;

-- Update existing incidents to 'pending' if they don't have it (though they will have it by default)
UPDATE incidents SET status = 'pending' WHERE status IS NULL;
