# Proposal

## Why

O repositório ainda é um starter documental: não existe módulo Go, aplicação Next.js nem contrato HTTP executável. Sem uma base que sobe, responde e pode ser testada, as próximas capacidades (catálogo, sessão, mídia) não têm onde viver nem como ser verificadas.

## What Changes

- Módulo Go provisório `example.com/gotube` com `cmd/api`, configuração validada no boot, `GET /healthz` e encerramento gracioso.
- Aplicação Next.js App Router em `web/`, com shell acessível e tema claro/escuro. O aviso de catálogo ausente não faz parte do comportamento consolidado; a listagem é catalog-home.
- PostgreSQL local via Docker Compose, volume persistente e credenciais só de desenvolvimento.
- OpenAPI limitado ao endpoint implementado, testes, scripts reais, CI e README de execução.

Fora de escopo: login, catálogo persistido, worker, outbox, fila, object storage, Redis, MediaMTX, Kubernetes e qualquer rota que ainda não exista.

## Capabilities

### New Capabilities

- `api-liveness`: o processo HTTP confirma que está aceitando requisições, recusa configuração inválida antes de escutar e encerra com prazo.
- `public-shell`: a interface pública mostra navegação e tema, sem simular vídeos, busca ou sessão. A listagem persistida é catalog-home.
- `local-postgres`: o ambiente de desenvolvimento sobe um PostgreSQL reproduzível, ainda sem a API consultar o banco.

### Modified Capabilities

- Nenhuma capacidade já consolidada. `catalog-home` e `public-video-list` já estão em `openspec/specs/`. Esta mudança não as altera.

## Impact

- Go: `go.mod`, `cmd/api` e `internal/platform`. A API não abre conexão com o banco nesta etapa.
- Next.js: nova aplicação em `web/`, preservando `web/AGENTS.md`.
- Dados: Compose e volume local; nenhuma migration de domínio.
- Operação: `.env.example`, `api/openapi.yaml`, scripts de verificação, workflow de CI e README. Sem push, merge ou deploy.
