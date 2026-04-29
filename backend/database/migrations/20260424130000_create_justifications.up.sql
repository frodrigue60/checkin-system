-- Up
CREATE TABLE justifications (
    id SERIAL PRIMARY KEY,
    incident_id INTEGER NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    employee_id INTEGER NOT NULL REFERENCES employees(id),
    message TEXT NOT NULL,
    evidence_url TEXT,
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected
    resolved_by INTEGER REFERENCES users(id),
    resolution_note TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_incident_justification UNIQUE(incident_id)
);

-- Indices
CREATE INDEX idx_justifications_incident ON justifications(incident_id);
CREATE INDEX idx_justifications_employee ON justifications(employee_id);
CREATE INDEX idx_justifications_status ON justifications(status);
