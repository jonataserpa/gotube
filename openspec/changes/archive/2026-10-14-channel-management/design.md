# Design

## Context

Ver `proposal.md`. A API já emite access token HS256 e o browser guarda esse token só em memória. `channels` existe sem dono, descrição nem categoria. `GET /v1/videos` filtra público publicado com cursor, sem `channel_id`. ADR-0001 mantém `GET /healthz` independente do banco. ADR-0002 continua a reger o shell; esta etapa não cria outro tema.

O curso, na seção 6, também lista avatar e CRUD administrativo de categorias. Esta etapa diverge de propósito: não há object storage, e o CRUD administrativo está na etapa 12. A divergência fica neste design para não parecer esquecimento.

## Goals / Non-Goals

**Goals:**

- Autorizar leitura privada e mutação com o access token já existente.
- Garantir no máximo um canal por conta e handle único normalizado.
- Trocar o conjunto de categorias do canal numa única transação.
- Manter a listagem pública restrita a publicado e público quando houver `channel_id`.

**Non-Goals:**

- Novo mecanismo de sessão, upload de avatar, papel administrativo e filtro `q` ou `category_id`.

## Decisions

1. **Bearer no access token.** As rotas privadas exigem `Authorization: Bearer`. O parser já rejeita outro algoritmo, emissor, audiência e expiração. O refresh continua só no cookie e não autoriza estas rotas. Alternativa descartada: promover o access token a cookie, o que mudaria o contrato da sessão já arquivada.

2. **Nome de exibição nulo até a edição.** `users.display_name` nasce nulo. `GET /v1/me` devolve `display_name: null` enquanto o dono não gravar um nome. A validação conta runas, não bytes: o nome tem de 1 a 80 runas depois de tirar espaços das pontas. Campo desconhecido ou tipo inválido segue o JSON estrito já usado na autenticação.

3. **Dono opcional e único.** `channels.owner_id` referencia `users` e é nulo no canal de laboratório. Um índice único parcial impede dois canais com o mesmo dono. A aplicação também recusa o segundo `POST` com 409. Violação única de handle ou de dono vira 409, não 500. O handle gravado é minúsculo, sem espaços nas pontas, com 3 a 30 caracteres entre letra, dígito e hífen.

4. **Categorias substituídas por completo.** `categories` tem `id`, `slug`, `name` e `active`. `channel_categories` liga canal e categoria. `PATCH` ou `POST` com `category_ids` apaga o conjunto anterior e grava o novo na mesma transação, só se todos os ids estiverem ativos. Lista vazia deixa o canal sem categoria. A leitura pública omite categoria que deixou de estar ativa, sem apagar o vínculo.

5. **Filtro na listagem existente.** `channel_id` ausente preserva a página atual. UUID válido sem vídeo público publicado devolve página vazia. Valor que não é UUID devolve 400 `INVALID_CHANNEL`. A query continua exigindo `PUBLIC` e `PUBLISHED`.

6. **Páginas.** `/channels/[handle]` é Server Component e chama a API interna sem credencial. `/perfil` é formulário cliente: lê e grava perfil, canal e categorias com o token em memória, pelo proxy same-origin que já confere `Origin` com `Host`. O proxy encaminha `Authorization` e não grava o token. Busca e Novo vídeo não ganham destino.

7. **Códigos.** 401 `UNAUTHENTICATED`, 400 `VALIDATION_ERROR` ou `INVALID_CHANNEL`, 403 `CHANNEL_FORBIDDEN`, 404 `CHANNEL_NOT_FOUND`, 409 `CHANNEL_ALREADY_EXISTS` ou `HANDLE_ALREADY_REGISTERED`, 503 `DATABASE_UNAVAILABLE`. O envelope continua `error.code` e `error.message` em português, sem `request_id`.

## Risks / Trade-offs

- [Corrida de handle ou de dono] → a restrição única do banco decide; a aplicação traduz `23505` para 409.
- [Canal de laboratório editável por engano] → `owner_id` nulo não casa com nenhuma conta; a rota responde 403.
- [`channel_id` alargar a listagem] → o filtro é adicional à condição já publicada; o teste com rascunho do mesmo canal precisa falhar antes da query.
- [Categoria inativa ainda vinculada] → a leitura esconde; a reativação futura volta a mostrá-la. Não há CRUD de reativação nesta etapa.

## Migration Plan

Nova migration, aplicada só por `go run ./cmd/migrate`, não no boot da API. Colunas novas são nulas ou têm default que não reescreve o canal de laboratório. O seed de desenvolvimento insere categorias ativas e uma inativa, marcado como fixture, sem dono no canal existente. Rollback é restaurar o banco anterior: não há migration descendente automática. A API antiga, se ainda selecionar colunas nomeadas de `channels`, continua lendo o catálogo.

## Open Questions

Nenhuma que mude o contrato. O formato do avatar espera a etapa com object storage.
