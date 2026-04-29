-- 20260426130000_add_service_notes_to_attendances.up.sql

ALTER TABLE attendances ADD COLUMN IF NOT EXISTS check_out_note TEXT;
ALTER TABLE attendances ADD COLUMN IF NOT EXISTS check_out_address TEXT;
