DO $$ BEGIN
    CREATE TYPE holiday_type AS ENUM ('mandatory', 'optional');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE holidays (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    date DATE NOT NULL UNIQUE,
    description TEXT NULL,
    type holiday_type NOT NULL DEFAULT 'mandatory',
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NULL,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NULL
);
