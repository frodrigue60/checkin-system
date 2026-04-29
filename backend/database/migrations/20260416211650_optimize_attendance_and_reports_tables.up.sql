CREATE INDEX IF NOT EXISTS attendances_work_center_id_index ON attendances (work_center_id);

ALTER TABLE reports 
ADD CONSTRAINT reports_unique_employee_period UNIQUE (employee_id, start_date, end_date);
