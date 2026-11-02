# Progresso de aprendizado

## bootstrap-project — 14/10/2026

- Mudança: `openspec/changes/bootstrap-project` (OpenSpec 1.14.1).
- Curso consultado: seções 1–3, M08–M10, 21–23 e F01–F06. Não implementa o catálogo nem a sessão.
- Conceitos: liveness versus readiness, prazo de leitura de cabeçalho, `Shutdown` com dono da goroutine, Server Component com uma ilha de tema, script antes da hidratação.
- Falha observada: `HTTP_SHUTDOWN_TIMEOUT=0s` faz `go run ./cmd/api` terminar com erro de configuração e sem escutar. `API_ADDR=http://127.0.0.1:18080` termina com `API_ADDR deve ser host:port, não uma URL` e código 1.

### Comandos e resultados

- `go test -race ./...`, `go vet ./...` e `gofmt -l .` vazio: sucesso.
- `GET http://127.0.0.1:18080/healthz`: `200` e `{"status":"ok"}`, com a API no ar e sem usar o banco. `POST` no mesmo caminho: `405`.
- `web`: `npm test` 4 testes em 3 arquivos; `npm run typecheck`, `npm run lint` e `npm run build` sucesso. Build gerou `/` estático.
- Navegador em `http://127.0.0.1:3000`: shell em pt-BR, estado vazio, tema persistido em `theme` e sem diálogo de erro de hidratação. Em 320px, `scrollWidth` igual a `clientWidth`.
- `docker compose config`: imagem `postgres:18.1-alpine` e volume `gotube_gotube-postgres`. Container ficou `healthy`. `select version()` retornou PostgreSQL 18.1. Recriar o container manteve o mesmo volume.
- Redocly lint de `api/openapi.yaml`: descrição válida. Restou o aviso de `info.license`, porque esta etapa não escolhe uma licença.

### Limitações

- `127.0.0.1:8080` já estava ocupada; a verificação da API usou `127.0.0.1:18080`.
- O workflow `.github/workflows/ci.yml` não rodou no GitHub.
- `npm audit` do `web/` reporta 5 vulnerabilidades altas na árvore gerada. Não rodei `npm audit fix --force`.
- ESLint 9.39.5 instalado pelo `create-next-app` aparece como deprecated. O lint passou mesmo assim.
- npm bloqueou o postinstall de `unrs-resolver`. O lint executado não dependeu desse script.
- OpenSpec tooling também reportou 4 vulnerabilidades altas, isoladas em `.tooling/openspec`.
- Não há `go.sum`: o módulo ainda não tem dependência externa.

## Shell visual — 14/10/2026

- ADR `docs/adr/0002-obsidian-shell.md`: shell escuro no estilo da referência, marca GoTube.
- O HTML de exemplo não foi copiado com vídeos, canais ou métricas fictícias. Busca, novo vídeo e destinos além de Início ficam desabilitados.
- Tema padrão passa a ser escuro. `localStorage` `light` continua trocando o tema.

## public-video-catalog — 14/10/2026

- Mudança: `openspec/changes/public-video-catalog`.
- Conceitos: paginação por cursor (keyset) em `(published_at, id)` descendente, liveness separada da consulta ao banco, Server Component lendo só `GO_API_INTERNAL_URL`.
- O teste de listagem falhou primeiro com `query ainda não existe`. Depois da query, o mesmo teste passou.
- Os testes da home falharam enquanto a página ainda dizia que o catálogo não existia. Depois da grade, passaram.

### Comandos e resultados

- `go test ./cmd/... ./internal/...` com `DATABASE_URL` no Compose local: sucesso. `gofmt -l .` vazio. `go vet ./...` sucesso.
- `go run ./cmd/migrate` e `go run ./cmd/seed` no banco `gotube` do Compose: sucesso, sem saída de erro.
- API em `127.0.0.1:18082` porque `8080` e `18080` já estavam ocupadas.
- `GET /healthz`: `200` e `{"status":"ok"}`.
- `GET /v1/videos`: `200`, 12 itens, `next_cursor` presente. A página não trouxe “Rascunho de exemplo” nem “Vídeo privado de exemplo”. Havia item com `thumbnail_url` nulo.
- `GET /v1/videos?cursor=<next_cursor>`: `200`, 2 itens (`Exemplo público 20` e `Catálogo público de exemplo`), sem repetir id da primeira página, `next_cursor` nulo.
- `cursor=nao-e-cursor`: `400` `INVALID_CURSOR`. `limit=51`: `400` `INVALID_LIMIT`.
- Com o container Postgres parado: `GET /v1/videos` `503` `DATABASE_UNAVAILABLE` e o corpo não tinha `items`. `GET /healthz` continuou `200`. O container foi iniciado de novo e a listagem voltou `200`.
- `web`: `npm test` 12 testes; `npm run lint`, `npm run typecheck` e `npm run build` sucesso. A rota `/` ficou como partial prerender.
- Redocly lint de `api/openapi.yaml`: válido. Continua o aviso de `info.license`.
- Navegador em `http://127.0.0.1:3012`: grade com título, canal, duração e “Sem imagem”; “Próxima página” abriu `/?cursor=` com os dois itens seguintes. Em 320px, `scrollWidth` igual a `clientWidth`. O HTML da primeira resposta já continha “Exemplo público 32” e não continha o rascunho nem o privado.
- O Next que estava em `3000` foi encerrado para subir este servidor com `GO_API_INTERNAL_URL`. A prévia desta verificação ficou em `3012`.

### Limitações

- O workflow `.github/workflows/ci.yml` ganhou o serviço PostgreSQL 18.1, mas não rodou no GitHub.
- `npm audit` do `web/` continua com 5 vulnerabilidades altas. Não rodei `npm audit fix --force`.

## identity-session — 14/10/2026

- Mudança: `openspec/changes/identity-session`.
- Conceitos: Argon2id, JWT HS256 de 15 minutos só em memória no browser, refresh opaco em cookie `HttpOnly` com rotação e família, recuperação com resposta genérica.
- Cadastro não abre sessão. Login inativo devolve o mesmo `401` de senha errada. Reset revoga os refresh da conta.

### Comandos e resultados

- `go test ./cmd/... ./internal/...` com `DATABASE_URL` do Compose: sucesso. `gofmt -l` vazio e `go vet ./cmd/... ./internal/...` sucesso.
- `go run ./cmd/migrate` aplicou `002_identity.sql` no banco `gotube`.
- API em `127.0.0.1:18083` com `AUTH_EXPOSE_RESET_TOKEN=true` e `AUTH_COOKIE_SECURE=false`. `8080`, `18080` e `18082` já estavam ocupadas.
- Next de produção em `127.0.0.1:3014` com `GO_API_INTERNAL_URL=http://127.0.0.1:18083`.
- `POST /api/auth/register` com `Origin` de outro host: `403` `FORBIDDEN_ORIGIN`. Com o mesmo host do `Host`: `201` e `{id,email}`, sem cookie.
- `POST /api/auth/logout`: `204` sem corpo e `Set-Cookie` `gotube_refresh` com `Max-Age=0`, `HttpOnly`, `SameSite=Lax`, `Path=/api/auth`.
- `web`: build de produção passou depois do proxy. As rotas `/login`, `/register`, `/forgot-password`, `/reset-password` e `/api/auth/[...path]` existem.
- Navegador em `http://127.0.0.1:3014`: cadastro de `navegador-20261014@example.com` mostrou “Conta criada.” e o shell continuou em “Entrar”. Login levou a `/` com o e-mail no shell, “Sair”, “Exemplo público 32” e “Nenhuma inscrição ainda.”. Sair voltou o link “Entrar” e a grade permaneceu. Recuperação mostrou a mensagem genérica e o link “Redefinir senha neste laboratório”. A redefinição mostrou “Senha atualizada.”. Entrar com a senha nova voltou o e-mail e o catálogo. Recarregar depois de sair manteve “Entrar”.

### Correções vistas na verificação

- `request.nextUrl.origin` não coincidia com `http://127.0.0.1:3014`. A checagem passou a comparar o host do `Origin` com o cabeçalho `Host`.
- `logout` e `reset` respondem `204`. O proxy do Next quebrava ao montar `Response` com corpo nesses status. Resposta vazia com `null` devolve o `204` e o `Set-Cookie`.

### Limitações

- O workflow de CI não rodou no GitHub.
- `npm audit` do `web/` não foi forçado.
- Redocly continua com o aviso de `info.license` e o aviso de que `POST /v1/auth/logout` não documenta 4xx. Logout não tem 4xx de propósito.
- `GET/PATCH /me`, canal e categorias ficam para a próxima etapa.

## channel-management — 14/10/2026

- Mudança: `openspec/changes/channel-management`.
- Conceitos: um canal por conta, categorias ativas associadas ao canal, leitura pública por handle sem e-mail, filtro `channel_id` que não relaxa `PUBLIC` + `PUBLISHED`.
- `GET /v1/categories` falhou primeiro com 404. O filtro `channel_id` falhou devolvendo os dois canais. Os testes da home e do shell falharam antes do link e de `/perfil`. A página pública falhou na importação do componente.

### Comandos e resultados

- `go test ./cmd/... ./internal/...` com `DATABASE_URL` do Compose: sucesso.
- `go run ./cmd/migrate` e `go run ./cmd/seed` no banco `gotube`: sucesso. O canal `laboratorio` continua sem dono.
- `web`: `npm test` 17 testes; `npm run typecheck`, `npm run lint` e `npm run build` sucesso. Rotas `/perfil`, `/channels/[handle]` e `/api/account/[...path]`.
- Redocly lint de `api/openapi.yaml`: válido. Avisos que permanecem: `info.license`, `POST /v1/auth/logout` sem 4xx e `GET /v1/categories` sem 4xx. A lista não tem 4xx de propósito; sem banco ela responde 503.
- API desta verificação em `127.0.0.1:18084`. `8080` e `18080` já estavam ocupadas. O `.env` local não define `AUTH_JWT_SECRET`; o processo recebeu o segredo de laboratório do exemplo, sem gravar o arquivo.
- Next de produção em `127.0.0.1:3015` com `GO_API_INTERNAL_URL=http://127.0.0.1:18084`. O `web/.env` não foi alterado. A porta `3000` já tinha outro Next.
- `GET /healthz`: `200` `{"status":"ok"}`.
- `GET /v1/categories`: `200` com Go e Programação, sem a categoria inativa.
- `GET /v1/channels/laboratorio`: `200` com nome e descrição vazia, sem e-mail.
- Navegador em `http://127.0.0.1:3015`: cadastro e login de `canal-20261014@example.com`. O shell mostrou o e-mail, “Perfil” e “Sair”. “Novo vídeo” continuou desabilitado. Em `/perfil`, salvar o nome mostrou “Nome atualizado.”. Criar o canal `ada-canal` com a categoria Go mostrou “Canal criado.”. A página pública mostrou `/ada-canal`, “Canal Ada” e “Notas de laboratório”, sem o e-mail da conta. O HTML dessa rota, pedido sem cookie, também não continha o e-mail. O card “Laboratório GoTube” da home abriu `/channels/laboratorio` com o nome do canal e sem o e-mail. `localStorage` não tinha o access token.

### Limitações

- O workflow de CI não rodou no GitHub.
- `npm audit` do `web/` não foi forçado.
- A API em `18084` e o Next em `3015` ficaram no ar para esta verificação.

## video-upload — 14/10/2026

- Mudança: `openspec/changes/video-upload`.
- Conceitos: rascunho `DRAFT`/`PRIVATE`/`WAITING_UPLOAD` no canal da conta, PUT assinado direto ao storage, conclusão que confere o objeto e grava um único `video.uploaded.v1`. Sem multipart, relay, worker ou playback.
- O teste de schema falhou primeiro com `column "processing_status" does not exist`. `POST /v1/videos` falhou com 405. A sessão falhou com 404 antes da rota. O shell falhou enquanto Novo vídeo seguia desabilitado.

### Comandos e resultados

- `go test ./cmd/... ./internal/...` com `DATABASE_URL` do Compose: sucesso. Inclui rascunho, sessão, objeto ausente ou grande demais, conclusão 202 e segunda conclusão sem outro evento. `/healthz` segue 200 quando banco ou storage não atendem essas rotas.
- `go run ./cmd/migrate` no banco `gotube`: a migration `004_video_upload.sql` entrou. O boot da API não aplica migration.
- `web`: `npm test` 20 testes em 8 arquivos; `npm run typecheck`, `npm run lint` e `npm run build` sucesso. O access token do fluxo não usa `localStorage` nem `NEXT_PUBLIC_`.
- Redocly lint de `api/openapi.yaml`: válido. Avisos que permanecem: `info.license`, `POST /v1/auth/logout` sem 4xx e `GET /v1/categories` sem 4xx.
- ADR `docs/adr/0003-direct-upload.md`: PUT assinado; `/healthz` não consulta o storage; sem multipart e sem relay nesta etapa.
- Compose ganhou MinIO. A imagem `minio/minio` não existe mais no Docker Hub; o serviço usa `cgr.dev/chainguard/minio:latest`. `PutBucketCors` respondeu 501, então o CORS do browser fica em `MINIO_API_CORS_ALLOW_ORIGIN`.
- API desta verificação em `127.0.0.1:18085`, com o segredo de laboratório do exemplo só no processo. `8080` e `18084` já estavam ocupadas. Next de produção em `127.0.0.1:3016` com `GO_API_INTERNAL_URL=http://127.0.0.1:18085`. O `web/.env` não foi alterado.
- `GET /healthz`: `200` `{"status":"ok"}`. `POST /healthz`: `405`.
- Navegador em `http://127.0.0.1:3016`: a sessão restaurada habilitou Novo vídeo e a busca continuou desabilitada. O diálogo criou “Aula direta de laboratório”. O envio de 4 MiB mostrou “Enviando 0%”, o XHR aceitou `4194304/4194304`, a interface mostrou “Arquivo recebido.” e só então “Envio confirmado.”. Os dois rascunhos ficaram `QUEUED`, `DRAFT` e `PRIVATE`. A segunda conclusão da mesma sessão respondeu `202` e o evento `video.uploaded.v1` continuou único. A home e `GET /v1/videos` não mostraram esses títulos. `localStorage` ficou vazio.

### Limitações

- O workflow de CI não rodou no GitHub.
- `npm audit` do `web/` não foi forçado.
- A API em `18085`, o Next em `3016` e o MinIO ficaram no ar para esta verificação. `18084` e `3015` continuam com o processo anterior.

## video-processing

O relay publica a outbox com confirmação e o worker gera uma thumbnail e uma rendition HLS. `READY` continua rascunho privado. O dono acompanha o estado e só pede de novo quando houve falha.

### Conceito

A outbox grava o evento na mesma transação do estado. Outro processo publica na fila e só marca a linha depois da confirmação. Se cair no meio, a mesma mensagem pode sair de novo; o worker reconhece a geração já concluída e não cria outro conjunto de arquivos. O FFmpeg roda fora da transação, com argumentos separados e prazo. A conclusão usa um fencing token para uma lease antiga não sobrescrever a geração vigente.

### Comandos e resultados

- `go test ./cmd/... ./internal/...` com `DATABASE_URL`: sucesso. Cobre relay, cerca, falha limitada, rotas do dono e `/healthz` sem a fila.
- `web`: `npm test` 22 testes em 8 arquivos; `npm run typecheck`, `npm run lint` e `npm run build` sucesso.
- Redocly lint de `api/openapi.yaml`: válido, com os mesmos três avisos de antes.
- Compose ganhou RabbitMQ `4.3.6-alpine`. `go run ./cmd/migrate` aplicou a migration `005` no banco `gotube`. O boot da API não migra.
- API em `127.0.0.1:18085`, Next em `127.0.0.1:3016`, relay e worker no ar com a fila. `18084` e `3015` não foram alterados.
- Navegador: o envio de um arquivo curto real mostrou “Envio confirmado.”, “Na fila.” e “Processamento concluído.”. O texto não anunciou publicação nem que dá para assistir. O vídeo ficou `READY`, `DRAFT`, `PRIVATE`, duração 1 segundo, com playlist, segmento e thumbnail. `video.uploaded.v1` e `video.ready.v1` foram publicados. A listagem pública não trouxe o título.

### Limitações

- Há uma rendition só. Nada consome `video.ready.v1`.
- O workflow de CI não rodou no GitHub.

## Próxima tarefa

O próximo item do roadmap é `video-playback`: autorização, player, privacidade e edição, publicação e exclusão. O comando de planejamento é `/opsx-propose video-playback`.
