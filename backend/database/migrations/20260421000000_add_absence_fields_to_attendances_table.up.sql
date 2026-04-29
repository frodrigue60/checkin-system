ALTER TABLE attendances ADD COLUMN is_absence BOOLEAN DEFAULT FALSE;
ALTER TABLE attendances ADD COLUMN absence_reason TEXT;
