# Proposal

## Why

O envio já grava `video.uploaded.v1` e deixa o vídeo `QUEUED`, mas nada publica esse evento nem produz thumbnail e HLS. A etapa 6 do roadmap pede a outbox, a fila, o worker e o FFmpeg antes de reproduzir ou publicar.

## What Changes

- Um processo de relay reivindica linhas não publicadas da outbox, publica no RabbitMQ com confirmação e só então marca `published_at`. Republicar o mesmo evento é seguro.
- Um worker consome `video.uploaded.v1`, inspeciona o objeto com ffprobe e gera uma rendition HLS mais uma thumbnail. A conclusão condicional marca `READY` ou `FAILED` e grava `video.ready.v1`. Entrega repetida não produz outro conjunto da mesma geração.
- O vídeo continua `DRAFT` e `PRIVATE` mesmo quando fica `READY`. A listagem pública não muda.
- O dono consulta `GET /v1/videos/{id}/processing` e, se o estado for `FAILED`, pede `POST /v1/videos/{id}/retry-processing`. Outra conta não vê nem reprocessa.
- Depois da confirmação do envio, a interface acompanha esse estado. Falha mostra o erro seguro e a nova tentativa. `READY` não anuncia vídeo publicado nem pronto para assistir.

Suposições: uma rendition, sem escada adaptativa; geração começa em 1 e o retry incrementa; sem `Idempotency-Key` e sem `request_id`; `video.ready.v1` é gravado e retransmitido, sem consumidor nesta etapa; lease expirada volta ao worker, sem varredura de objetos órfãos no bucket; testes de estado usam dublês de fila e de mídia, e o FFmpeg real entra quando o binário existe.

Fora deste escopo: publicar, editar, excluir, reproduzir, player, variantes 720p/1080p, multipart, quota, comentários e deploy em cluster.

## Capabilities

### New Capabilities

- `outbox-relay`: publicar eventos duráveis na fila e tolerar queda entre a publicação e a marcação.
- `media-processing`: inspecionar o objeto, gerar thumbnail e HLS e concluir uma geração uma única vez.
- `processing-status`: o dono lê o estado e o erro seguro, e reprocessa só uma falha.

### Modified Capabilities

- `upload-progress`: depois que o envio é confirmado, a interface mostra o processamento e a falha, sem tratar isso como publicação.

## Impact

- Go: executáveis `cmd/outbox` e `cmd/worker`, separados da API. Duas rotas autenticadas no `ServeMux` existente. Sem banco, essas rotas respondem 503. `GET /healthz` não consulta fila nem FFmpeg.
- Next.js: o fluxo de Novo vídeo passa a consultar o estado e oferecer a nova tentativa. Os bytes continuam fora da API.
- Dados: migration com geração, jobs, assets e o claim do relay. Vídeos `READY` do seed não entram na fila.
- Operação: Compose local ganha RabbitMQ com tag fixa. Sem player e sem publicação nesta etapa. `api/openapi.yaml` ganha só as rotas deste fluxo.
