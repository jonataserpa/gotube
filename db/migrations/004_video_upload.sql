ALTER TABLE videos
    ADD COLUMN processing_status TEXT NOT NULL DEFAULT 'WAITING_UPLOAD'
        CHECK (processing_status IN ('WAITING_UPLOAD', 'QUEUED', 'PROCESSING', 'READY', 'FAILED'));

UPDATE videos
SET processing_status = 'READY'
WHERE publication_status = 'PUBLISHED';

CREATE TABLE upload_sessions (
    id UUID PRIMARY KEY,
    video_id UUID NOT NULL REFERENCES videos (id),
    owner_id UUID NOT NULL REFERENCES users (id),
    object_key TEXT NOT NULL UNIQUE,
    declared_size BIGINT NOT NULL CHECK (declared_size > 0),
    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX upload_sessions_one_open
    ON upload_sessions (video_id)
    WHERE completed_at IS NULL;

CREATE TABLE outbox (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    type TEXT NOT NULL,
    version INTEGER NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0
);
