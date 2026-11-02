# Tasks

## 1. Dados

- [x] 1.1 Ajustar o teste de schema para exigir `users.display_name`, `channels.owner_id` nulo no canal de laboratório, `channels.description`, `categories` e `channel_categories`; verificar que falha antes da migration nova
- [x] 1.2 Criar a migration e incluir no seed de desenvolvimento categorias ativas e uma inativa, sem dono no canal de laboratório; verificar que o teste de 1.1 passa e que o boot da API não executa migration

## 2. Perfil

- [x] 2.1 Escrever o teste HTTP de `GET/PATCH /v1/me`: 200 com identificador, e-mail e nome, sem hash; 401 sem token; 400 para nome em branco ou campo desconhecido, sem gravar; 503 sem banco com `GET /healthz` ainda 200; verificar que falha antes do perfil existir
- [x] 2.2 Implementar `GET/PATCH /v1/me` com Bearer e nome de 1 a 80 runas; verificar que o teste de 2.1 passa

## 3. Canal

- [x] 3.1 Escrever os testes de ausência explícita, criação 201, segundo canal 409, handle repetido 409, handle inválido 400, edição do dono, 403 para outra conta e para o canal sem dono, 404 para id desconhecido, leitura pública sem e-mail e 503 sem criar canal; verificar que falham antes do canal autenticado existir
- [x] 3.2 Implementar `GET /v1/me/channel`, `POST /v1/channels`, `PATCH /v1/channels/{id}` e `GET /v1/channels/{handle}`; verificar que os testes de 3.1 passam

## 4. Categorias

- [x] 4.1 Escrever os testes da lista só com ativas, da associação completa, do 400 que não altera o conjunto quando há id inativo ou desconhecido, da recusa por outra conta e da leitura pública que omite categoria inativa; incluir 503 da lista sem corpo de sucesso; verificar que falham antes da associação existir
- [x] 4.2 Implementar `GET /v1/categories` e a troca transacional de `category_ids` na criação e na edição do próprio canal; verificar que os testes de 4.1 passam

## 5. Listagem por canal

- [x] 5.1 Escrever o teste de `GET /v1/videos?channel_id=`: só públicos publicados daquele canal, rascunho do mesmo canal omitido, id mal formado 400 e id sem vídeo 200 com página vazia; verificar que falha antes do filtro existir
- [x] 5.2 Implementar o filtro sem relaxar visibilidade nem publicação; verificar que o teste de 5.1 passa e que a lista sem `channel_id` permanece a atual

## 6. Contrato

- [x] 6.1 Acrescentar em `api/openapi.yaml` somente as rotas e códigos desta mudança; verificar que o teste de contrato permite esse acréscimo, continua proibindo `/watch`, `/playback`, `/upload` e `/admin`, e que o lint do OpenAPI passa

## 7. Next.js

- [x] 7.1 Escrever os testes da home, do shell e das páginas: o nome do canal aponta para `/channels/{handle}` e o card não aponta para reprodução; com sessão há destino de perfil; busca e Novo vídeo seguem indisponíveis; a página pública do canal mostra nome e descrição sem e-mail; verificar que falham antes dessas rotas existirem
- [x] 7.2 Implementar o proxy same-origin que encaminha `Authorization`, as páginas `/perfil` e `/channels/[handle]`, o link do card e o destino do shell; verificar com os testes do `web/`, typecheck, lint e build, e que o access token não usa `localStorage` nem `NEXT_PUBLIC_`

## 8. Integração

- [x] 8.1 Rodar `go test ./cmd/... ./internal/...` com `DATABASE_URL` do Compose; verificar perfil, canal, categorias, filtro e que `/healthz` segue 200 quando o banco não atende essas rotas
- [x] 8.2 No navegador, editar o nome, criar o canal com categoria ativa, abrir a página pública sem e-mail e seguir o card da home até o canal de laboratório; registrar os resultados em `docs/learning/progress.md`

## Workflow follow-up

- Revisar esta proposta antes de implementar. A aplicação começa só com `/opsx-apply` ou pedido explícito.
- Arquivar depois da revisão. Os Purpose consolidados de `public-video-list`, `catalog-home` e `public-shell` permanecem; o delta não os altera.
