-- Migration: Remove field work support from attendances
ALTER TABLE attendances DROP COLUMN IF EXISTS check_in_address;
ALTER TABLE attendances DROP COLUMN IF EXISTS is_field_work;
