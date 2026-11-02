# ADR 0004 — Processamento em relay e worker

- Status: aceito em 14/10/2026, durante video-processing.
- Specs: `openspec/changes/video-processing/specs/`.

## Contexto

A conclusão do upload já grava `video.uploaded.v1` e deixa o vídeo `QUEUED`. Falta publicar a outbox, gerar thumbnail e HLS, e mostrar esse estado ao dono. Publicar e reproduzir continuam no item seguinte do roadmap.

`GET /healthz` é liveness (ADR 0001). O upload direto não consulta a fila (ADR 0003). O curso descreve escada adaptativa, player, Wire, chi e `request_id`.

## Alternativas

- Publicar a outbox dentro do processo da API: o boot ou o `/healthz` passariam a depender do RabbitMQ.
- Várias renditions agora: o player que as escolheria ainda não existe.
- Confirmar a mensagem antes de gravar `READY`: uma queda no meio perderia o trabalho ou o publicaria duas vezes sem cerca.

## Decisão

`cmd/outbox` reivindica a linha, publica com confirmação e só então grava `published_at`. `cmd/worker` consome só `video.uploaded.v1`, com prefetch 1, e confirma a mensagem depois do estado durável. A API sobe sem RabbitMQ e sem FFmpeg.

Há uma rendition HLS, em 360p ou na altura da origem se ela for menor, mais uma thumbnail. A conclusão exige geração e fencing token. `READY` não muda `DRAFT` nem `PRIVATE`. `video.ready.v1` é gravado e retransmitido; nada o consome nesta etapa.

O dono lê `GET /v1/videos/{id}/processing` e, em `FAILED`, chama `POST /v1/videos/{id}/retry-processing`. O erro devolvido não traz object key. Sem `request_id` e sem `Idempotency-Key`. O roteador continua sendo `ServeMux`.

## Consequências

Uma queda entre a publicação e a marcação pode entregar o mesmo evento de novo. O worker trata a geração já concluída como no-op. `/healthz` em 200 não significa que a fila ou o FFmpeg estão prontos. A escada adaptativa e o player ficam para video-playback.

## Critério de revisão

Quando o player passar a pedir mais de uma rendition, ou quando o relay precisar de mais de um consumidor, rever a fila e o contrato de liveness.
