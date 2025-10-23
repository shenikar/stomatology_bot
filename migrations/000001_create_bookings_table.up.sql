CREATE TABLE
    IF NOT EXISTS bookings (
        id SERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL,
        name VARCHAR(255),
        contact VARCHAR(255),
        datetime TIMESTAMPTZ
    );