CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    user_id BIGINT,
    name VARCHAR(255),
    contact VARCHAR(255),
    datetime TIMESTAMPTZ
);