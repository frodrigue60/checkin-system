-- Migration: Remove check_in_note from attendances
ALTER TABLE attendances DROP COLUMN IF EXISTS check_in_note;
