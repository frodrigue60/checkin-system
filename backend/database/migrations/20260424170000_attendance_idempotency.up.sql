CREATE UNIQUE INDEX idx_unique_active_attendance ON attendances (employee_id) WHERE (check_out IS NULL);
