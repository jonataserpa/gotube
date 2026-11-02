# Tasks

## 1. Contrato e schema

- [x] 1.1 Escrever `docs/adr/0004-media-processing.md` com relay e worker separados, `/healthz` sem fila nem FFmpeg, uma rendition e as divergências do curso.
- [x] 1.2 Escrever um teste de schema que falha porque faltam `videos.generation`, `video_assets`, `processing_jobs` e o claim da outbox. Depois, migration `005` aplicada só por `cmd/migrate`: geração default 1, assets únicos por vídeo, geração e object key, e job com lease e fencing token. O boot da API continua sem migrar.
- [x] 1.3 Acrescentar no OpenAPI só `GET /v1/videos/{id}/processing` e `POST /v1/videos/{id}/retry-processing`, com 401, 403, 404, 409 e 503. O teste de contrato continua sem `/watch`, `/playback` e `/admin`.

## 2. Relay

- [x] 2.1 Teste que falha: uma linha sem `published_at` é publicada num broker falso e só então marcada. Se o broker recusa, `published_at` permanece vazio. A mensagem tem ids e não tem bytes nem URL de cliente.
- [x] 2.2 Teste que falha: dois relays não reivindicam a mesma linha; claim expirado e ainda não marcado pode ser publicado de novo.
- [x] 2.3 Implementar `cmd/outbox`. A API sobe sem `RABBITMQ_URL` e `GET /healthz` continua 200.

## 3. Worker

- [x] 3.1 Teste que falha com runner de mídia falso: objeto com vídeo produz thumbnail, playlist e segmento na geração. Sem trilha de vídeo, o estado fica `FAILED` e não há playlist.
- [x] 3.2 Teste que falha: conclusão vigente marca `READY`, grava a duração, insere os assets e um único `video.ready.v1`, e mantém `DRAFT` e `PRIVATE`.
- [x] 3.3 Teste que falha: token ou geração antigos não alteram o vídeo. Reentrega de geração já `READY` não cria outro asset nem outro evento.
- [x] 3.4 Teste que falha: falha transitória respeita o máximo e então fica `FAILED` com erro sem object key. Mensagem inválida é confirmada e não volta para sempre. Lease `PROCESSING` expirada volta a ser reivindicável. Vídeo `READY` do seed não ganha job.
- [x] 3.5 Implementar o pipeline com ffprobe e uma rendition, sem shell concatenado, e `cmd/worker`. Teste com binário real só quando `ffmpeg` e `ffprobe` estão no `PATH`; caso contrário, o teste é pulado.

## 4. Estado para o dono

- [x] 4.1 Teste que falha: o dono lê 200 sem object key nem URL. Sem token, 401. Outra conta, 403. Id desconhecido, 404.
- [x] 4.2 Teste que falha: retry de `FAILED` responde 202, volta a `QUEUED`, incrementa a geração e grava um evento. `QUEUED` ou `READY` respondem 409 sem escrita. Outra conta responde 403.
- [x] 4.3 Teste que falha: sem banco, as duas rotas respondem 503 e `GET /healthz` continua 200. Implementar as rotas no `ServeMux` e o proxy same-origin no Next.

## 5. Interface

- [x] 5.1 Teste que falha: depois de “Envio confirmado.”, `QUEUED` e `PROCESSING` aparecem e o texto não contém “publicado” nem “pronto para assistir”.
- [x] 5.2 Teste que falha: `FAILED` mostra a mensagem segura e “Tentar de novo”. `READY` não anuncia publicação nem que dá para assistir. As regras de progresso em bytes permanecem.

## 6. Verificação

- [x] 6.1 Compose com RabbitMQ em tag fixa. Documentar `RABBITMQ_URL` e os binários só em `.env.example`.
- [x] 6.2 `go test` dos pacotes Go com `DATABASE_URL`, lint do OpenAPI, e no `web` typecheck, lint, teste e build.
- [x] 6.3 Se o FFmpeg existir, conferir no browser um arquivo curto de verdade até `READY` ou `FAILED`, sem anunciar publicação. Atualizar `docs/learning/progress.md` com o resultado e o próximo item do roadmap.
