-- Up migration: Increase coordinate precision
ALTER TABLE work_centers 
    ALTER COLUMN latitude TYPE DOUBLE PRECISION,
    ALTER COLUMN longitude TYPE DOUBLE PRECISION;

ALTER TABLE attendances
    ALTER COLUMN check_in_latitude TYPE DOUBLE PRECISION,
    ALTER COLUMN check_in_longitude TYPE DOUBLE PRECISION,
    ALTER COLUMN check_out_latitude TYPE DOUBLE PRECISION,
    ALTER COLUMN check_out_longitude TYPE DOUBLE PRECISION;
