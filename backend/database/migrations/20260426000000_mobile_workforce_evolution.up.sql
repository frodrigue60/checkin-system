-- 20260426000000_mobile_workforce_evolution.up.sql

-- 1. Add evidence_url to attendances for future photo support
ALTER TABLE attendances ADD COLUMN IF NOT EXISTS evidence_url TEXT;

-- 2. Add shift_type to work_shifts to support flexible/field work
-- Types: 'fixed' (default), 'flexible' (no late penalties), 'field' (no late or geofence penalties)
ALTER TABLE work_shifts ADD COLUMN IF NOT EXISTS shift_type VARCHAR(20) DEFAULT 'fixed';

-- 3. Relax idempotency: The previous unique constraint might be too strict if we want multi-check per day.
-- However, we currently don't have a unique constraint on (employee_id, date), 
-- it was enforced in CODE. So no SQL change needed for that specifically, 
-- but we might want to ensure we don't have a stale one.
