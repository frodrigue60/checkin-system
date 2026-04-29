-- 20260426130000_add_service_notes_to_attendances.down.sql

ALTER TABLE attendances DROP COLUMN IF EXISTS check_out_note;
ALTER TABLE attendances DROP COLUMN IF EXISTS check_out_address;
