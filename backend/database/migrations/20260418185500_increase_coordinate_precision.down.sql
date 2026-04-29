-- Down migration: Restore DECIMAL(9,6)
ALTER TABLE work_centers 
    ALTER COLUMN latitude TYPE DECIMAL(9, 6),
    ALTER COLUMN longitude TYPE DECIMAL(9, 6);

ALTER TABLE attendances
    ALTER COLUMN check_in_latitude TYPE DECIMAL(9, 6),
    ALTER COLUMN check_in_longitude TYPE DECIMAL(9, 6),
    ALTER COLUMN check_out_latitude TYPE DECIMAL(9, 6),
    ALTER COLUMN check_out_longitude TYPE DECIMAL(9, 6);
