-- Migration: Add field work support to attendances
ALTER TABLE attendances ADD COLUMN check_in_address TEXT;
ALTER TABLE attendances ADD COLUMN is_field_work BOOLEAN DEFAULT FALSE;
