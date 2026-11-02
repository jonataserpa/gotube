# Próxima mudança após bootstrap — colar após /opsx-propose public-video-catalog

Especifique o catálogo público do GoTube, conectado de ponta a ponta.
Backend: listar somente vídeos PUBLIC/PUBLISHED, DTO público, limite de página,
cursor estável, PostgreSQL e endpoint documentado em api/openapi.yaml.
Frontend: página inicial Next com título, thumbnail, duração, loading/vazio/erro.
SSR só consulta dados públicos. Não implementar autenticação, upload ou recomendação.
Fixtures/seed são dados de desenvolvimento explicitamente identificados.
Inclua cenários para rascunhos e vídeos privados não aparecerem, API indisponível,
paginação consistente e tela vazia. Revise a proposta comigo antes de implementar.
