CREATE TABLE reports (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    total_hours_worked DECIMAL(8, 2) NOT NULL,
    total_earnings DECIMAL(10, 2) NOT NULL,
    days_worked INTEGER NOT NULL,
    daily_breakdown JSONB NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    CONSTRAINT reports_employee_id_foreign FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE
);
