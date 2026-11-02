# Design

## Context

A API grava `video.uploaded.v1` e deixa o vídeo `QUEUED`, `DRAFT` e `PRIVATE`. O relay e o worker ainda não existem. `GET /healthz` é só liveness (ADR 0001) e não consulta PostgreSQL nem storage (ADR 0003). A migration não roda no boot.

O curso descreve escada adaptativa, player e `request_id`. Esta etapa entrega uma rendition, sem reproduzir e sem publicar.

## Goals / Non-Goals

**Goals:**

- Publicar a outbox com confirmação e tolerar a queda entre publicar e marcar.
- Produzir thumbnail e HLS de uma geração e concluir com update condicional.
- Deixar o dono ler o estado e tentar de novo só quando `FAILED`.
- Mostrar esse estado depois da confirmação do envio.

**Non-Goals:**

- Publicar, editar, excluir, autorizar playback ou servir o player.
- Escada 720p/1080p, multipart, varredura de objetos órfãos e deploy em cluster.
- Consumir `video.ready.v1`.

## Decisions

1. **Relay e worker são processos à parte.** `cmd/outbox` reivindica linhas sem `published_at`. `cmd/worker` consome só `video.uploaded.v1`. A API não publica no boot e sobe sem RabbitMQ. Alternativa rejeitada: relay dentro da API, porque derrubaria o boot ou o `/healthz` quando a fila falha.

2. **Claim do relay.** A outbox ganha `claim_until`. Dois relays não pegam a mesma linha ao mesmo tempo. Claim expirado e ainda sem `published_at` pode ser publicado de novo. A duplicata é segura porque o worker é idempotente por geração.

3. **ACK só depois do estado durável.** Prefetch 1. Mensagem inválida é confirmada e não volta para sempre. Falha transitória de storage ou mídia incrementa tentativas e volta à fila até o máximo; depois o vídeo fica `FAILED` e a mensagem é confirmada. O erro guardado tem `code` e mensagem em português, sem object key.

4. **Uma rendition.** ffprobe lê duração, dimensões e áudio. O HLS é VOD em 360p, ou na altura da origem se ela for menor, sem ampliar. Áudio entra se existir. Argumentos do FFmpeg não são concatenados num shell. Há prazo e stderr limitado. Segmentos sobem antes da playlist; a thumbnail JPEG sobe depois. Keys: `videos/{id}/generations/{gen}/hls/...` e `.../thumbnail.jpg`. Saídas são conferidas antes do commit.

5. **Job com cerca.** `processing_jobs` guarda geração, status, `lease_until`, `fencing_token` e tentativas. O update de conclusão exige geração e token vigentes. Sucesso, na mesma transação, marca `READY`, grava `duration_seconds`, insere `video_assets` e um `video.ready.v1`. Entrega repetida de geração já `READY` só confirma a mensagem. Lease de `PROCESSING` expirada volta a ser reivindicável; esgotadas as tentativas, vira `FAILED`.

6. **Geração no vídeo.** `videos.generation` é inteiro, `NOT NULL`, default 1. O retry só em `FAILED` incrementa, volta a `QUEUED` e grava um `video.uploaded.v1` novo, sem mudar `publication_status` nem `visibility`. Vídeos `READY` do seed não têm job e o worker não os varre.

7. **Rotas do dono.** `GET /v1/videos/{id}/processing` devolve status e, se couber, o erro seguro. `POST /v1/videos/{id}/retry-processing` responde 202 só em `FAILED`; nos outros estados, 409 `PROCESSING_NOT_FAILED` sem escrita. 401, 403 `VIDEO_FORBIDDEN` e 404 `VIDEO_NOT_FOUND` seguem o padrão atual. Sem banco, 503. Sem `request_id` e sem `Idempotency-Key`.

8. **Interface.** Depois de “Envio confirmado.”, o fluxo consulta o estado. `FAILED` mostra a mensagem segura e “Tentar de novo”. `READY` pode dizer que o processamento terminou e não pode dizer que está publicado ou que dá para assistir. Os bytes continuam no PUT direto.

9. **Testes e operação.** Fila e FFmpeg têm dublês no teste de estado. Um teste com binário real só roda se `ffprobe` e `ffmpeg` estão no `PATH`. O Compose ganha RabbitMQ com tag fixa. `.env.example` documenta `RABBITMQ_URL` e os caminhos dos binários. A migration é a 005 e não roda no boot. O OpenAPI ganha só as duas rotas; o contrato continua sem `/watch`, `/playback` e `/admin`.

Divergências a registrar no ADR 0004: sem escada adaptativa, sem Wire, `ServeMux` em vez de chi, e `video.ready.v1` só gravado e retransmitido.

## Risks / Trade-offs

- **Uma rendition.** O arquivo fica assistível numa etapa futura, mas sem adaptação de banda. A escada fica para quando houver player.
- **Publicar de novo após queda.** A fila pode ver o mesmo evento duas vezes. O worker trata geração já concluída como no-op.
- **FFmpeg no host.** O laboratório não empacota o worker. O teste de estado não depende do binário; a verificação no browser, sim.
- **`READY` ainda é rascunho.** A interface pode sugerir que já dá para assistir. O texto permitido evita “publicado” e “pronto para assistir”.
