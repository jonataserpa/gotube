# Tasks

## 1. Dados

- [x] 1.1 Ajustar o teste de schema para exigir `videos.processing_status`, `upload_sessions` e `outbox`; vídeo publicado do seed fica `READY` e o rascunho de exemplo fica `WAITING_UPLOAD`, com o canal de laboratório ainda sem dono; verificar que falha antes da migration nova
- [x] 1.2 Criar a migration e o backfill no seed; verificar que o teste de 1.1 passa e que o boot da API não executa migration
- [x] 1.3 Acrescentar MinIO ao Compose e as variáveis de exemplo, sem sobrescrever `.env`; verificar que `docker compose config` renderiza o serviço

## 2. Rascunho

- [x] 2.1 Escrever o teste HTTP de `POST /v1/videos`: 201 `DRAFT`/`PRIVATE`/`WAITING_UPLOAD` no canal da conta; 409 sem canal; 401 sem token; 400 para título em branco ou campo desconhecido, sem gravar; o rascunho ausente em `GET /v1/videos`; 503 sem banco com `GET /healthz` ainda 200; verificar que falha antes do rascunho existir
- [x] 2.2 Implementar `POST /v1/videos` com Bearer e duração zero; verificar que o teste de 2.1 passa

## 3. Sessão de upload

- [x] 3.1 Escrever os testes com dublê de storage: autorização só do dono, 403 para outra conta, 404 para id desconhecido, 400 quando o cliente envia key, 409 quando o vídeo já não está `WAITING_UPLOAD`, conclusão 202 com um evento, objeto ausente ou grande demais em 409 sem enfileirar, segunda conclusão sem outro evento, e 503 quando o storage não responde com `/healthz` ainda 200; verificar que falham antes da sessão existir
- [x] 3.2 Implementar `POST /v1/videos/{id}/upload-sessions` e `POST /v1/upload-sessions/{id}/complete` com key gerada no servidor, PUT assinado e a transação de `QUEUED` mais `video.uploaded.v1`; verificar que os testes de 3.1 passam

## 4. Contrato

- [x] 4.1 Acrescentar em `api/openapi.yaml` somente as rotas e códigos desta mudança; ajustar o teste de contrato para permitir essas rotas e continuar proibindo `/watch`, `/playback` e `/admin`; verificar que o lint do OpenAPI passa
- [x] 4.2 Registrar em `docs/adr/` a decisão de upload direto, incluindo a divergência de não fazer multipart nem relay nesta etapa; verificar que o ADR cita o PUT assinado e que `/healthz` não consulta o storage

## 5. Next.js

- [x] 5.1 Escrever os testes do shell e do fluxo: com sessão, Novo vídeo abre o rascunho; busca continua indisponível; progresso parcial e bytes concluídos não anunciam vídeo processado ou publicado; falha de envio não anuncia fila; verificar que falham antes do fluxo existir
- [x] 5.2 Implementar o fluxo com `XMLHttpRequest` direto ao storage e a confirmação pelo proxy que encaminha `Authorization`; verificar com os testes do `web/`, typecheck, lint e build, e que o access token não usa `localStorage` nem `NEXT_PUBLIC_`

## 6. Integração

- [x] 6.1 Rodar `go test ./cmd/... ./internal/...` com `DATABASE_URL` do Compose; verificar rascunho, sessão, conclusão idempotente e que `/healthz` segue 200 quando banco ou storage não atendem essas rotas
- [x] 6.2 No navegador, criar o rascunho, ver o progresso do envio, confirmar a sessão, repetir a confirmação e checar que a home não mostra o rascunho; registrar os resultados em `docs/learning/progress.md`

## Workflow follow-up

- Revisar esta proposta antes de implementar. A aplicação começa só com `/opsx-apply` ou pedido explícito.
- Arquivar depois da revisão. O Purpose consolidado de `public-shell` permanece; o delta altera o requisito de sessão. `public-video-list` e `catalog-home` não mudam.
