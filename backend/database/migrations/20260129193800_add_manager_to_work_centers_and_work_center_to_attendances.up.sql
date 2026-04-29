ALTER TABLE work_centers 
ADD COLUMN manager_id BIGINT NULL,
ADD CONSTRAINT work_centers_manager_id_foreign FOREIGN KEY (manager_id) REFERENCES employees (id) ON DELETE SET NULL;

ALTER TABLE attendances 
ADD COLUMN work_center_id BIGINT NULL,
ADD CONSTRAINT attendances_work_center_id_foreign FOREIGN KEY (work_center_id) REFERENCES work_centers (id) ON DELETE SET NULL;
