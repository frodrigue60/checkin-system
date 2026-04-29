CREATE TABLE work_shifts (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    expected_check_in TIME(0) WITHOUT TIME ZONE NOT NULL,
    expected_check_out TIME(0) WITHOUT TIME ZONE NOT NULL,
    allowed_lunch_time TIME(0) WITHOUT TIME ZONE NOT NULL,
    tolerance_time TIME(0) WITHOUT TIME ZONE NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NULL
);
