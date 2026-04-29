CREATE TABLE attendances (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL,
    work_shift_id BIGINT NULL,
    date DATE NOT NULL,
    check_in TIME(0) WITHOUT TIME ZONE NULL,
    lunch_start TIME(0) WITHOUT TIME ZONE NULL,
    lunch_end TIME(0) WITHOUT TIME ZONE NULL,
    check_out TIME(0) WITHOUT TIME ZONE NULL,
    check_in_latitude DECIMAL(9, 6) NULL,
    check_in_longitude DECIMAL(9, 6) NULL,
    check_out_latitude DECIMAL(9, 6) NULL,
    check_out_longitude DECIMAL(9, 6) NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    CONSTRAINT attendances_employee_id_foreign FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE,
    CONSTRAINT attendances_work_shift_id_foreign FOREIGN KEY (work_shift_id) REFERENCES work_shifts (id) ON DELETE SET NULL
);
