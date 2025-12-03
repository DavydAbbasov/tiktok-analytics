CREATE TYPE video_tracking_status AS ENUM ('active', 'stopped', 'error');

ALTER TABLE videos
    ADD COLUMN tracking_status video_tracking_status NOT NULL DEFAULT 'active',
    ADD COLUMN last_error      TEXT,
    ADD COLUMN last_error_at   TIMESTAMPTZ;


