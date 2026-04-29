ALTER TABLE positions 
DROP COLUMN IF EXISTS late_penalty_fee, 
DROP COLUMN IF EXISTS out_of_range_fee;
