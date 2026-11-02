CREATE TABLE channels (
    id UUID PRIMARY KEY,
    handle TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

CREATE TABLE videos (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL REFERENCES channels (id),
    title TEXT NOT NULL,
    duration_seconds INTEGER NOT NULL CHECK (duration_seconds >= 0),
    thumbnail_url TEXT,
    visibility TEXT NOT NULL CHECK (visibility IN ('PUBLIC', 'UNLISTED', 'PRIVATE')),
    publication_status TEXT NOT NULL CHECK (
        publication_status IN ('DRAFT', 'PUBLISHED', 'DELETING', 'DELETED', 'BLOCKED')
    ),
    published_at TIMESTAMPTZ,
    CONSTRAINT published_has_timestamp CHECK (
        (
            publication_status = 'PUBLISHED'
            AND published_at IS NOT NULL
        )
        OR (
            publication_status <> 'PUBLISHED'
            AND published_at IS NULL
        )
    )
);

CREATE INDEX videos_public_published_keyset
    ON videos (published_at DESC, id DESC)
    WHERE visibility = 'PUBLIC' AND publication_status = 'PUBLISHED';
