CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    position_id BIGINT NOT NULL,
    work_center_id BIGINT NOT NULL,
    work_shift_id BIGINT NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    CONSTRAINT employees_user_id_foreign FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT employees_position_id_foreign FOREIGN KEY (position_id) REFERENCES positions (id) ON DELETE CASCADE,
    CONSTRAINT employees_work_center_id_foreign FOREIGN KEY (work_center_id) REFERENCES work_centers (id) ON DELETE CASCADE,
    CONSTRAINT employees_work_shift_id_foreign FOREIGN KEY (work_shift_id) REFERENCES work_shifts (id) ON DELETE CASCADE
);
