ALTER TABLE reports ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'valid';
UPDATE reports SET status = 'valid' WHERE status IS NULL;
