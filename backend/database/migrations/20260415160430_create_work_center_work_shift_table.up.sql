CREATE TABLE work_center_work_shift (
    id BIGSERIAL PRIMARY KEY,
    work_center_id BIGINT NOT NULL,
    work_shift_id BIGINT NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    CONSTRAINT work_center_work_shift_work_center_id_foreign FOREIGN KEY (work_center_id) REFERENCES work_centers (id) ON DELETE CASCADE,
    CONSTRAINT work_center_work_shift_work_shift_id_foreign FOREIGN KEY (work_shift_id) REFERENCES work_shifts (id) ON DELETE CASCADE
);
