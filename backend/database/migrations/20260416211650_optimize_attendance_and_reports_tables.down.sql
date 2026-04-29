ALTER TABLE reports 
DROP CONSTRAINT IF EXISTS reports_unique_employee_period;

DROP INDEX IF EXISTS attendances_work_center_id_index;
