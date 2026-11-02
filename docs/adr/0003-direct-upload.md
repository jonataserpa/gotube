# ADR 0003 — Upload direto com PUT assinado

- Status: aceito em 14/10/2026, durante video-upload.
- Spec: `openspec/changes/video-upload/specs/upload-session/spec.md`.

## Contexto

O rascunho precisa de um arquivo no object storage antes de qualquer processamento. Receber os bytes na API prenderia conexão e banda no processo Go. O curso também descreve multipart e um relay para a fila na mesma etapa; esta base ainda não tem worker.

`GET /healthz` já é liveness (ADR 0001). Object storage é mais uma dependência e não pode mudar esse contrato.

## Alternativas

- Receber o arquivo na API ou gravá-lo em disco local servido pelo processo: concentra banda e impede o progresso no browser sem ocupar o processo.
- Multipart e relay para RabbitMQ nesta etapa: exigem worker, FFmpeg e uma fila que o roadmap coloca em video-processing.
- Deixar o cliente escolher a object key ou uma URL de mídia: o servidor perderia o controle do objeto que depois será conferido.

## Decisão

O browser envia os bytes com um PUT assinado direto ao storage compatível com S3. A API só emite a autorização e, na conclusão, consulta existência e tamanho. A object key nasce no servidor, no formato `videos/{video_id}/source/{session_id}`. Sem endpoint de storage, as rotas de upload respondem 503.

`GET /healthz` não consulta o storage. Continua respondendo 200 quando o processo HTTP está vivo, mesmo que o storage não atenda a autorização ou a conclusão.

Nesta etapa não há multipart nem relay. A conclusão idempotente grava `video.uploaded.v1` na outbox, na mesma transação que marca o vídeo `QUEUED`. Quem publica o evento fica para video-processing.

## Consequências

Um objeto pode ficar órfão se o browser enviar os bytes e não concluir a sessão. A sessão expira; a limpeza fica na etapa de processamento. `/healthz` em 200 não significa que o bucket está pronto.

## Critério de revisão

Quando video-processing passar a consumir a outbox, ou quando um arquivo passar do tamanho de um único PUT assinado, rever multipart e o relay sem alterar o contrato de liveness.
