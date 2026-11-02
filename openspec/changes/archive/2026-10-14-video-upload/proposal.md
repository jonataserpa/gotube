# Proposal

## Why

A conta com canal ainda não consegue enviar um vídeo. A etapa 5 do roadmap pede o rascunho, a autorização de storage, o progresso no navegador e a confirmação idempotente, antes do processamento.

## What Changes

- `POST /v1/videos` cria um rascunho no canal da conta autenticada. O vídeo nasce `DRAFT`, `PRIVATE` e `WAITING_UPLOAD`. Sem canal próprio, a API recusa e não cria vídeo.
- `POST /v1/videos/{id}/upload-sessions` autoriza um único envio. A object key nasce no servidor, com prazo e tamanho máximo. O cliente não escolhe a key nem envia uma URL de mídia.
- O navegador envia os bytes direto ao storage e mostra o progresso desse envio. Cem por cento dos bytes não significa vídeo processado.
- `POST /v1/upload-sessions/{id}/complete` confere o objeto no storage. A mesma sessão concluída de novo devolve o mesmo sucesso e não pede processamento outra vez. A transação marca `QUEUED` e grava `video.uploaded.v1` na outbox.
- Com sessão, Novo vídeo abre esse fluxo. Busca continua indisponível. O rascunho não entra na listagem pública.

Suposições: um PUT assinado, sem multipart; object storage compatível com S3 no Compose local; duração zero até a inspeção; sem cabeçalho `Idempotency-Key`, porque a idempotência é a da sessão; o relay, a fila, o worker e o FFmpeg ficam na etapa `video-processing`.

Fora deste escopo: publicar, excluir, editar o rascunho depois de criado, reproduzir, inspecionar com ffprobe, gerar thumbnail ou HLS, quota por quantidade de vídeos, e upload no canal de laboratório que não tem dono.

## Capabilities

### New Capabilities

- `video-draft`: criar um rascunho privado no canal da conta autenticada, invisível no catálogo público.
- `upload-session`: autorizar o objeto, verificar a conclusão no storage e registrar o pedido de processamento uma única vez.
- `upload-progress`: mostrar no navegador o progresso do envio direto, separado do processamento.

### Modified Capabilities

- `public-shell`: com sessão, Novo vídeo abre o fluxo de rascunho; busca permanece indisponível.

## Impact

- Go: rotas autenticadas no `ServeMux` existente. Sem token válido, 401. Sem banco ou sem storage, as rotas de upload respondem 503 e `GET /healthz` continua 200.
- Next.js: Novo vídeo deixa de ser um botão morto e passa a criar o rascunho, acompanhar o envio e confirmar a sessão. Os bytes não passam pela API nem por Server Action.
- Dados: migration com `processing_status`, `upload_sessions` e `outbox`. Vídeos já publicados no seed permanecem fora da fila. O rascunho de exemplo continua fora da listagem pública.
- Operação: Compose local ganha object storage. Sem RabbitMQ, worker ou FFmpeg nesta etapa. `api/openapi.yaml` ganha só as rotas deste fluxo.
