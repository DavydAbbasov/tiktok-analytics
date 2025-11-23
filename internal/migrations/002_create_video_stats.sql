CREATE TABLE IF NOT EXISTS video_stats (
    id SERIAL PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    views BIGINT NOT NULL,
    earnings NUMERIC(12, 4) NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_video_stats_video_id_captured_at
    ON video_stats(video_id, captured_at);