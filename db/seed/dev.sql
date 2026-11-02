-- Fixture de desenvolvimento. Este arquivo é dado de desenvolvimento.
-- DEVELOPMENT FIXTURE. Do not apply in production.

INSERT INTO channels (id, handle, name)
VALUES (
    '00000000-0000-4000-8000-000000000001',
    'laboratorio',
    'Laboratório GoTube'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO categories (id, slug, name, active)
VALUES
    (
        '00000000-0000-4000-8000-000000000101',
        'programacao',
        'Programação',
        TRUE
    ),
    (
        '00000000-0000-4000-8000-000000000102',
        'go',
        'Go',
        TRUE
    ),
    (
        '00000000-0000-4000-8000-000000000103',
        'arquivada',
        'Arquivada',
        FALSE
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO videos (
    id,
    channel_id,
    title,
    duration_seconds,
    thumbnail_url,
    visibility,
    publication_status,
    published_at
)
VALUES
    (
        '00000000-0000-4000-8000-000000000010',
        '00000000-0000-4000-8000-000000000001',
        'Catálogo público de exemplo',
        125,
        'https://example.com/gotube/dev-thumb.jpg',
        'PUBLIC',
        'PUBLISHED',
        '2026-01-02T15:04:05Z'
    ),
    (
        '00000000-0000-4000-8000-000000000011',
        '00000000-0000-4000-8000-000000000001',
        'Rascunho de exemplo',
        10,
        NULL,
        'PUBLIC',
        'DRAFT',
        NULL
    ),
    (
        '00000000-0000-4000-8000-000000000012',
        '00000000-0000-4000-8000-000000000001',
        'Vídeo privado de exemplo',
        30,
        NULL,
        'PRIVATE',
        'PUBLISHED',
        '2026-01-01T00:00:00Z'
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO videos (
    id,
    channel_id,
    title,
    duration_seconds,
    thumbnail_url,
    visibility,
    publication_status,
    published_at
)
SELECT
    ('00000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
    '00000000-0000-4000-8000-000000000001',
    'Exemplo público ' || n,
    60 + n,
    CASE
        WHEN n % 2 = 0 THEN NULL
        ELSE 'https://example.com/gotube/dev-thumb.jpg'
    END,
    'PUBLIC',
    'PUBLISHED',
    timestamptz '2026-02-01T00:00:00Z' + make_interval(hours => n)
FROM generate_series(20, 32) AS n
ON CONFLICT (id) DO NOTHING;

UPDATE videos
SET processing_status = 'READY'
WHERE publication_status = 'PUBLISHED';
