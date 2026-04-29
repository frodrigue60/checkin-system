ALTER TABLE reports 
ADD COLUMN work_center_id BIGINT NULL,
ADD CONSTRAINT reports_work_center_id_foreign FOREIGN KEY (work_center_id) REFERENCES work_centers (id) ON DELETE SET NULL;
