DO $$ BEGIN
    CREATE TYPE incident_type AS ENUM ('late', 'early', 'absent', 'overtime', 'out_of_range');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE incidents (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL,
    work_center_id BIGINT NOT NULL,
    attendance_id BIGINT NOT NULL,
    type incident_type NOT NULL DEFAULT 'late',
    description TEXT NULL,
    is_late BOOLEAN NOT NULL DEFAULT false,
    delay_minutes INTEGER NOT NULL DEFAULT 0,
    is_out_of_range BOOLEAN NOT NULL DEFAULT false,
    distance INTEGER NOT NULL DEFAULT 0,
    check_in_time TIME(0) WITHOUT TIME ZONE NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    CONSTRAINT incidents_employee_id_foreign FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE,
    CONSTRAINT incidents_work_center_id_foreign FOREIGN KEY (work_center_id) REFERENCES work_centers (id) ON DELETE CASCADE,
    CONSTRAINT incidents_attendance_id_foreign FOREIGN KEY (attendance_id) REFERENCES attendances (id) ON DELETE CASCADE
);
