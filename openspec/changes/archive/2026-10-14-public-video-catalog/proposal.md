# Proposal

## Why

A página inicial declara que o catálogo não existe, e a API só responde liveness. O visitante ainda não consegue ver vídeos públicos persistidos, com página estável e falha distinta de lista vazia.

## What Changes

- `GET /v1/videos` lista só vídeos com visibilidade `PUBLIC` e publicação `PUBLISHED`, sem credencial.
- A resposta é um DTO público: identificador, título, duração, thumbnail opcional, nome e handle públicos do canal, mais `next_cursor`.
- A página usa cursor opaco e `limit`. A ordem é `published_at` descendente e, no empate, `id` descendente.
- O PostgreSQL passa a guardar canal e vídeo. Um seed local, identificado como dado de desenvolvimento, inclui um vídeo público e outros que não podem aparecer.
- A home do Next substitui o aviso de catálogo ausente por grade, vazio, carregamento ou erro. A leitura inicial ocorre no servidor, só com dados públicos.
- `GET /healthz` permanece liveness e continua sem consultar o banco.

Fora deste escopo: busca (`q`), filtro por categoria ou canal, autenticação, upload, processamento, reprodução, recomendação e página de assistir. A seção 27 do curso prevê esses filtros; eles ficam para uma mudança posterior. O card não vira link para rota inexistente.

## Capabilities

### New Capabilities

- `public-video-list`: listagem pública paginada, persistida e restrita a vídeos publicados e públicos.
- `catalog-home`: página inicial que mostra essa lista e distingue vazio, carregamento e falha da API.

### Modified Capabilities

- Nenhuma. `openspec/specs/` ainda não tem capacidades consolidadas. O comportamento de `/healthz` não muda.

## Impact

- Go: módulo de vídeo, pool PostgreSQL e migration. A API ganha dependência de banco só na listagem.
- Next.js: a home em `web/src/components/catalog/home-page.tsx` deixa de ser o estado “ainda não implementado”. `GO_API_INTERNAL_URL` passa a ser usada no servidor.
- Dados: tabelas mínimas de canal e vídeo, seed de desenvolvimento e volume já existente do Compose.
- Contrato: `api/openapi.yaml` ganha `GET /v1/videos` e os erros dessa rota. Sem worker, fila, storage ou Kubernetes.
