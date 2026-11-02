# Tasks

## 1. API Go

- [x] 1.1 Criar o módulo `example.com/gotube` e o teste que recusa endereço ausente e prazo inválido; verificar com `go test ./internal/platform/config` — passou em `go test -race ./...`
- [x] 1.2 Implementar a carga de configuração e o teste de que `cmd/api` não escuta quando a configuração é inválida; verificar com `go test ./cmd/api` — passou; execução real com URL em `API_ADDR` saiu com código 1
- [x] 1.3 Implementar `GET /healthz` e o teste de 200 JSON sem credencial, além de método não contratado; verificar com `go test ./internal/platform/httpapi` — passou; `curl` em `:18080` devolveu 200 e POST devolveu 405
- [x] 1.4 Implementar prazos HTTP e shutdown e o teste de parada do listener; verificar com `go test -race ./internal/platform/httpapi` — passou
- [x] 1.5 Registrar a decisão de liveness versus readiness em `docs/adr/0001-liveness-before-readiness.md` e descrever `/healthz` no README; o ADR separa readiness futura

## 2. Shell Next.js

- [x] 2.1 Gerar Next.js estável em `web/` sem sobrescrever `web/AGENTS.md` e registrar versões no lockfile; `web/package.json` tem Next 16.4.0 e o lockfile existe
- [x] 2.2 Adicionar teste do shell e da página vazia e implementar navegação, tema e estado honesto; `npm test` passou com 4 testes
- [x] 2.3 Completar scripts de typecheck, lint e build e executá-los; typecheck, lint e build terminaram com sucesso

## 3. Contrato e PostgreSQL

- [x] 3.1 Escrever `api/openapi.yaml` só com `GET /healthz` e o teste que impede endpoint de catálogo; teste Go passou e Redocly aceitou o documento, com aviso de licença
- [x] 3.2 Adicionar Compose com imagem fixada, volume e healthcheck, mais `.env.example`; `docker compose config` ok, container `healthy` e volume mantido na recriação
- [x] 3.3 Adicionar CI para Go, Next e validação do Compose com as versões compatíveis; o workflow referencia `go.mod`, `web/package-lock.json` e os scripts reais. Não foi executado no GitHub

## 4. Integração

- [x] 4.1 Subir a API em porta livre e conferir `/healthz` contra o contrato; `API_ADDR=127.0.0.1:18080 go run ./cmd/api` e `curl` retornaram 200 `{"status":"ok"}`
- [x] 4.2 Abrir o shell no navegador, acionar o tema e conferir ausência de erro de hidratação no console; em `http://127.0.0.1:3000` o tema persistiu e não houve diálogo de hidratação. Em 320px não houve overflow horizontal
- [x] 4.3 Executar `gofmt`, `go vet` e `go test ./...` e atualizar `docs/learning/progress.md` com resultados reais — os três gates Go passaram e o progresso registra os comandos

## Workflow follow-up

- Arquivar `bootstrap-project` depois que esta revisão da spec `public-shell` for aceita.
- `public-video-catalog` já foi arquivada. A próxima mudança de produto no roadmap é `identity-session`.
