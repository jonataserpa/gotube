ALTER TABLE users
    ADD COLUMN display_name TEXT;

ALTER TABLE channels
    ADD COLUMN description TEXT NOT NULL DEFAULT '',
    ADD COLUMN owner_id UUID REFERENCES users (id);

CREATE UNIQUE INDEX channels_one_owner
    ON channels (owner_id)
    WHERE owner_id IS NOT NULL;

CREATE TABLE categories (
    id UUID PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    active BOOLEAN NOT NULL
);

CREATE TABLE channel_categories (
    channel_id UUID NOT NULL REFERENCES channels (id),
    category_id UUID NOT NULL REFERENCES categories (id),
    PRIMARY KEY (channel_id, category_id)
);
