# Design

## Context

A API em `cmd/api` só publica `GET /healthz`. A home em `web/src/components/catalog/home-page.tsx` diz que o catálogo não existe. O Compose já sobe PostgreSQL 18.1, e a API não abre conexão. `docs/adr/0001-liveness-before-readiness.md` separa liveness de readiness. `openspec/specs/` está vazio; a mudança `bootstrap-project` ainda não foi arquivada e a spec `public-shell` dela exige o texto de catálogo ausente. Ver `proposal.md` para o motivo.

## Goals / Non-Goals

**Goals:**

- Uma listagem keyset que não perde nem repete empate de `published_at`.
- A home distingue itens, vazio e erro, com a primeira leitura no servidor.
- O seed local prova o filtro e se identifica como desenvolvimento.

**Non-Goals:**

- Busca, categoria, canal, player, login e readiness nova.
- sqlc, fila, storage e cache HTTP público. Cache fica para um exercício posterior, como no laboratório da seção 28.

## Decisions

### Rota e estados

O prefixo `/v1` segue a seção 6 do curso. `/healthz` fica fora dele e sem banco, como no ADR 0001. Vídeo visível na lista tem `visibility = PUBLIC` e `publication_status = PUBLISHED`. `UNLISTED` não entra: o curso reserva esse estado a quem tem o link. A query também não devolve `processing_status`. Publicar sem estar pronto continua sendo regra futura de publicação; esta lista não promove rascunho.

Alternativa rejeitada: filtrar só por `PUBLIC` e deixar rascunho público aparecer. O prompt pede os dois estados.

### Cursor

A chave é `(published_at, id)`, os dois descendentes. O cursor opaco é a codificação desses dois campos, não um offset. Cursor malformado responde 400. Cursor bem formado cujo item já sumiu continua a partir dessa chave. `limit` padrão 12, máximo 50, alinhado ao laboratório da seção 28. A home sempre pede 12; o limite não vira campo editável.

O envelope de erro desta rota é `{"error":{"code","message"}}`. O `request_id` do exemplo da seção 6 espera middleware de correlação, que esta mudança não introduz. Divergência consciente: os códigos `INVALID_LIMIT`, `INVALID_CURSOR` e `DATABASE_UNAVAILABLE` já bastam para o cliente separar 400 e 503 de uma página vazia.

### Persistência

Tabelas mínimas `channels` e `videos`, com `published_at` nulo fora do publicado. IDs são UUID gerados no seed e, no futuro, na criação. A migration vive em `db/migrations` e roda por um comando explícito, não no boot da API. O seed fica em arquivo separado, com cabeçalho de dado de desenvolvimento, e também não roda sozinho.

A API lê `DATABASE_URL`. As variáveis `POSTGRES_*` continuam sendo do Compose. Pool pgx na plataforma; a regra de quais linhas saem fica na query, coberta por teste contra PostgreSQL. Sem repositório genérico.

Alternativa rejeitada: sqlc nesta única consulta. Entra quando houver várias queries geradas. Alternativa rejeitada: a API criar o schema ao subir. Migration e processo ficam separados.

### Página

O chrome segue `docs/adr/0002-obsidian-shell.md`: barra superior, menu lateral e tema escuro. Esta mudança preenche a área principal; não redesenha o shell nem liga busca, chips ou “Novo vídeo”.

Server Component lê `GO_API_INTERNAL_URL`, variável só de servidor. A URL do cursor vai em `/?cursor=`. `loading.tsx` cobre a troca de página. Payload inválido e HTTP de erro caem na mesma tela de erro, com ação de tentar de novo. Thumbnail nula usa área reservada e texto, não uma imagem quebrada. Duração aparece em minutos e segundos; o contrato continua em segundos. Não há link de assistir nem de canal.

Zod valida a fronteira. TanStack Query não entra: a página não guarda uma segunda cópia cliente da lista.

### Teste de banco no CI

O teste de listagem sobe o schema num banco de teste. O workflow de CI ganha serviço PostgreSQL para esse teste. Teste que não sobe o banco não conta como prova do filtro.

## Risks / Trade-offs

- [A spec `public-shell` de `bootstrap-project` ainda pede o texto de catálogo ausente] → esta mudança substitui essa página. Arquivar o bootstrap antes de ajustar essa frase gravaria um requisito vencido em `openspec/specs/`. O arquivamento precisa retirar ou atualizar essa exigência.
- [Cursor não assinado pode ser forjado se o formato vazar] → forjar a chave só escolhe o ponto de continuação da ordem pública; não revela linha excluída. Assinatura fica para quando o cursor carregar filtro de usuário.
- [503 da listagem pode ser confundido com liveness] → `/healthz` segue 200. O corpo 503 não é `{"status":"ok"}` nem `items: []`.
- [Seed aplicado em produção mostraria vídeos fictícios] → o comando de seed é manual, nomeado como desenvolvimento e documentado como inadequado para produção.

## Migration Plan

Não há deploy. Em desenvolvimento: subir o Compose, aplicar a migration, aplicar o seed e apontar `DATABASE_URL` e `GO_API_INTERNAL_URL`. Rollback é parar a API e manter o volume; dropar as tabelas novas é comando explícito, separado de `compose down`.

## Open Questions

Nenhuma em aberto para o desenho das tarefas. A exclusão de `q`, `category_id` e `channel_id` está decidida na proposta, apesar de a seção 27 listar esses filtros no contrato futuro. A revisão antes de implementar pode rejeitar essa decisão.
