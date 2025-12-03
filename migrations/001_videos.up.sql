-- Create tables
CREATE TABLE IF NOT EXISTS videos (
    id SERIAL PRIMARY KEY,
    tiktok_id TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    current_views BIGINT NOT NULL DEFAULT 0,
    current_earnings NUMERIC(12, 4) NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
