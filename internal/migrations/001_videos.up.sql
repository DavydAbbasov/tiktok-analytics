-- Create tables
CREATE TABLE IF NOT EXISTS videos (
    id SERIAL PRIMARY KEY,
    tiktok_video_id TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    current_views BIGINT NOT NULL DEFAULT 0,
    current_earnings NUMERIC(12, 4) NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at trigger to entities table
CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON videos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

