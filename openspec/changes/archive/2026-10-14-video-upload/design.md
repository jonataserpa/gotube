# Design

## Context

O catálogo já lista só vídeo `PUBLIC` e `PUBLISHED`. A tabela `videos` ainda não tem `processing_status`; duração é obrigatória. Não há object storage no Compose. O envelope de erro é `error.code` e `error.message`, sem `request_id`. `GET /healthz` não consulta dependências (ADR-0001). Ver `proposal.md` para a motivação.

O curso propõe upload direto e, na mesma transação da conclusão, o estado `QUEUED` com `video.uploaded.v1`. O ADR desse upload ainda não está em `docs/adr/`; esta mudança o registra como decisão local, sem o relay.

## Goals / Non-Goals

**Goals:**

- Rascunho privado só no canal da conta, autorização de um PUT e conclusão que confere o objeto.
- A segunda conclusão da mesma sessão não cria outro evento.
- O navegador mostra progresso do envio e não anuncia processamento.

**Non-Goals:**

- Multipart, relay, RabbitMQ, worker, FFmpeg, publicação, playback e exclusão.
- Cabeçalho `Idempotency-Key` e `request_id` no envelope.

## Decisions

1. **PUT assinado, não proxy pela API.** O browser envia os bytes ao storage. A API só emite a autorização e depois consulta o objeto. Alternativa rejeitada: receber o arquivo na API, porque prende conexão e banda no processo Go. Alternativa rejeitada: disco local servido pela API, pelo mesmo motivo.

2. **Storage compatível com S3, MinIO no Compose.** A API depende de uma porta pequena: assinar PUT e consultar tamanho e existência. Testes usam um dublê. Sem endpoint configurado, a rota responde 503 e `/healthz` continua 200. Alternativa rejeitada: acoplar os testes a um container MinIO obrigatório.

3. **Key gerada no servidor.** Forma `videos/{video_id}/source/{session_id}`. Corpo com `object_key` ou URL é 400. O cliente declara o tamanho previsto; acima do máximo configurado, a sessão não nasce. O padrão de laboratório é 512 MiB, por variável de ambiente.

4. **Uma sessão aberta por vídeo.** Índice único parcial em `video_id` onde `completed_at` é nulo. Sessão expirada não completa. Enquanto o vídeo está `WAITING_UPLOAD`, uma sessão nova substitui a expirada. Vídeo já `QUEUED` responde 409.

5. **Conclusão na mesma transação.** `Head` no storage acontece antes do commit. Objeto ausente ou maior que o autorizado: 409 e o vídeo permanece `WAITING_UPLOAD`. Objeto aceito: `processing_status = QUEUED`, `completed_at` preenchido e uma linha na `outbox` com `video_id`, `generation` e `upload_session_id`. Sem bytes, sem key vinda do cliente e sem URL. A segunda chamada lê a sessão já completa e devolve 202 sem inserir outra linha.

6. **Estados.** `publication_status` continua `DRAFT`. `visibility` nasce `PRIVATE`. `processing_status` novo: os publicados do seed ficam `READY`; o rascunho de exemplo fica `WAITING_UPLOAD`. Duração do rascunho novo é 0 até a inspeção da etapa seguinte. A listagem pública não muda o filtro.

7. **Autorização.** O mesmo Bearer do perfil. Dono é `channels.owner_id`. Canal sem dono, inclusive `laboratorio`, não aceita sessão. Conta sem canal recebe 409 `CHANNEL_REQUIRED` na criação. Vídeo de outra conta: 403 `VIDEO_FORBIDDEN`. Id desconhecido: 404.

8. **Progresso no browser.** `XMLHttpRequest` contra a URL assinada, para ter progresso e aborto. A confirmação volta pela API, no proxy same-origin que já encaminha `Authorization`. Os bytes não passam pelo Next.

9. **CORS do bucket.** Só a origem do Next e o método PUT. O segredo do storage fica no ambiente da API, nunca em `NEXT_PUBLIC_`.

## Risks / Trade-offs

- [Objeto órfão se o browser envia e não completa] → a sessão expira e a etapa de processamento recolhe; esta etapa não apaga o bucket.
- [Head bem-sucedido e queda antes do commit] → o objeto existe e o vídeo segue `WAITING_UPLOAD`; um novo complete da mesma sessão ainda aberta conclui. Se a sessão expirou, outra sessão aponta para outra key.
- [ETag não é hash do arquivo] → a verificação usa existência e tamanho, não ETag como MD5.
- [Seed publicado marcado `READY` sem objeto] → o worker futuro não tenta transcodificar exemplo que não tem arquivo.

## Migration Plan

1. Migration nova, fora do boot da API: coluna `processing_status`, tabelas `upload_sessions` e `outbox`.
2. Backfill: `PUBLISHED` vira `READY`; o restante vira `WAITING_UPLOAD`.
3. Subir MinIO no Compose e documentar as variáveis ao lado de `.env.example`, sem sobrescrever `.env`.
4. Rollback desta etapa: parar as rotas novas. Não apagar objetos já enviados nem linhas `QUEUED` sem uma rotina de limpeza, que fica para o processamento.

## Open Questions

Nenhuma. O tamanho máximo padrão e a ausência de multipart estão fechados acima; mudá-los alteraria o contrato e exigiria outra mudança.
