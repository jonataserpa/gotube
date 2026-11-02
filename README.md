# GoTube

Monorepo de estudo: API Go na raiz e Next.js em `web/`. O bootstrap e o catálogo público estão implementados. Sessão, upload e reprodução ainda não existem.

O pacote original permanece em `template/`. A aplicação executável está nesta raiz.

## Versões desta etapa

| Ferramenta | Versão |
| --- | --- |
| Go | 1.27.1 local; `go 1.27` no `go.mod` |
| Node | 26.8.2 local; CI usa 24, dentro de `>=20.9.0` |
| npm | 12.0.2, lockfileVersion 3 |
| Next.js | 16.4.0 |
| React | 19.3.0 |
| Tailwind CSS | 4, configurado pelo `create-next-app` |
| PostgreSQL | `postgres:18.1-alpine` |
| OpenSpec | 1.14.1 em `.tooling/openspec` |

O módulo Go é `example.com/gotube`. Ele é provisório. Quando houver um caminho real de repositório, troque a linha `module` em `go.mod`, atualize os imports e rode `go mod tidy`.

## Preparar o ambiente

```bash
cp .env.example .env
```

`.env` não entra no Git. A senha de exemplo serve só para o PostgreSQL local.

Nesta máquina, `127.0.0.1:8080` já estava ocupada. O exemplo continua `:8080`. Se a porta estiver livre, use o valor do exemplo. Se não estiver, exporte outra antes de subir a API:

```bash
API_ADDR=127.0.0.1:18080
```

## PostgreSQL

```bash
docker compose -f deploy/compose/compose.yaml --env-file .env up -d
docker compose -f deploy/compose/compose.yaml --env-file .env ps
```

O volume `gotube_gotube-postgres` permanece depois de recriar o container. A primeira inicialização pode levar cerca de 30 segundos. O healthcheck usa `start_period: 40s` para esse init.

Encerrar sem apagar os dados:

```bash
docker compose -f deploy/compose/compose.yaml --env-file .env down
```

`down -v` remove o volume e os dados. Dropar as tabelas do catálogo é um comando separado; `compose down` não faz isso.

A API não cria o schema ao subir. No diretório da raiz, com `DATABASE_URL` no ambiente:

```bash
go run ./cmd/migrate
go run ./cmd/seed
```

`cmd/migrate` aplica `db/migrations`. `cmd/seed` aplica `db/seed/dev.sql`, um fixture de desenvolvimento. O seed não deve ser usado em produção. Os dois comandos são separados: a migration não carrega o seed e o seed não migra.

## API

```bash
set -a
source .env
set +a
go run ./cmd/api
```

Em outro terminal, use o host e a porta em que a API escutou. Com o exemplo `:8080`:

```bash
curl -sS http://127.0.0.1:8080/healthz
```

A resposta implementada é `{"status":"ok"}` com HTTP 200.

A listagem pública, com o schema e o seed aplicados:

```bash
curl -sS http://127.0.0.1:8080/v1/videos
```

Ela não pede credencial. Sem `limit`, a página traz até 12 itens. A página seguinte usa o `next_cursor` opaco. `limit` acima de 50 ou cursor irreconhecível respondem 400. Se o PostgreSQL não atender, a rota responde 503 e o corpo não é uma lista vazia. `/healthz` continua 200 nesse caso.

`GET /healthz` é liveness: o processo está aceitando requisições. Não é readiness. Não consulta o PostgreSQL e não deve ser usado, no futuro, como prova de que o banco está pronto. Readiness terá outro caminho quando existir uma rota que dependa do banco. A decisão está em `docs/adr/0001-liveness-before-readiness.md`.

Um prazo explícito inválido ou não positivo impede o boot, antes de abrir o listener. Ausente, os prazos ficam em 2s, 5s, 10s, 60s e 10s para cabeçalho, leitura, escrita, idle e shutdown. SIGINT e SIGTERM pedem `Shutdown` dentro do prazo de shutdown.

Encerre a API com Ctrl+C no terminal que executa `go run`.

## Web

```bash
cd web
npm ci
npm run dev
```

Abra `http://127.0.0.1:3000`. A home lê `GO_API_INTERNAL_URL` no servidor e mostra a grade, o vazio ou o erro do catálogo. O card não abre uma página de reprodução. O botão Tema alterna claro e escuro e guarda a escolha em `localStorage` sob a chave `theme`. Busca, novo vídeo, inscrições e biblioteca continuam indisponíveis: essas rotas ainda não existem.

Encerre o Next com Ctrl+C.

Scripts reais:

```bash
npm run typecheck
npm run lint
npm run test
npm run build
```

`npm run test` executa Vitest uma vez e termina. `web/.env.example` define `GO_API_INTERNAL_URL` para o servidor Next. A variável não usa o prefixo `NEXT_PUBLIC_`.

## Verificação local

```bash
gofmt -l .
go vet ./...
go test ./...
docker compose -f deploy/compose/compose.yaml --env-file .env config --quiet
```

OpenSpec:

```bash
bash scripts/openspec.sh validate bootstrap-project --type change --strict --no-interactive
```

O workflow de CI está em `.github/workflows/ci.yml`. Ele não foi executado no GitHub nesta sessão.

## Próximo passo

Revisar `public-video-catalog` e arquivar `bootstrap-project` só depois de atualizar a exigência de `public-shell` que ainda pede o texto de catálogo ausente. Busca, categoria e canal ficam para outra mudança.
