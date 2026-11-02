# Arquitetura inicial

Status: decisões de partida fornecidas pelo solicitante; validar durante bootstrap.

- Monorepo: Go na raiz, Next.js em web/.
- Monólito modular com API/worker/outbox separados por executável.
- PostgreSQL para dados; object storage para mídia; RabbitMQ/outbox para jobs duráveis.
- FFmpeg supervisionado por Go. MediaMTX para plano de mídia ao vivo.
- Go autoriza operações; Next.js compõe a experiência e usa a API.
- SSR público, sessão privada restaurada no cliente, access token em memória.
- Mesma origem pública por reverse proxy é a topologia desejada; desenvolvimento precisa
  configurar encaminhamento explícito quando integrar o frontend à API.
- Contrato HTTP em api/openapi.yaml. Eventos versionados em api/events/ quando necessários.
- O bootstrap implementa a API com `GET /healthz`, o shell Next e o PostgreSQL local. O contrato completo do produto e a infraestrutura de mídia ainda não existem.

ADRs propostos para formalizar durante execução: monólito modular; DI manual; SQL/sqlc;
upload direto; outbox/idempotência; HLS/MediaMTX; Next App Router; sessão/cache privado.
