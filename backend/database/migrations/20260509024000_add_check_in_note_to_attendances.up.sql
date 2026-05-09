-- Migration: Add check_in_note to attendances
ALTER TABLE attendances ADD COLUMN check_in_note TEXT;
