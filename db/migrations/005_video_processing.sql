ALTER TABLE videos
    ADD COLUMN generation INTEGER NOT NULL DEFAULT 1 CHECK (generation >= 1);

ALTER TABLE outbox
    ADD COLUMN claim_until TIMESTAMPTZ;

CREATE INDEX outbox_unpublished
    ON outbox (created_at)
    WHERE published_at IS NULL;

CREATE TABLE video_assets (
    id UUID PRIMARY KEY,
    video_id UUID NOT NULL REFERENCES videos (id),
    generation INTEGER NOT NULL CHECK (generation >= 1),
    kind TEXT NOT NULL CHECK (kind IN ('HLS_PLAYLIST', 'HLS_SEGMENT', 'THUMBNAIL')),
    object_key TEXT NOT NULL,
    checksum TEXT NOT NULL,
    CONSTRAINT video_assets_generation_key UNIQUE (video_id, generation, object_key)
);

CREATE TABLE processing_jobs (
    id UUID PRIMARY KEY,
    video_id UUID NOT NULL REFERENCES videos (id),
    generation INTEGER NOT NULL CHECK (generation >= 1),
    status TEXT NOT NULL CHECK (status IN ('QUEUED', 'PROCESSING', 'READY', 'FAILED')),
    lease_until TIMESTAMPTZ,
    fencing_token BIGINT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    error_code TEXT,
    error_message TEXT,
    CONSTRAINT processing_jobs_generation_key UNIQUE (video_id, generation)
);
