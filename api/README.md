# Contratos

`api/openapi.yaml` descreve `GET /healthz`, `GET /v1/videos` e as rotas `POST /v1/auth/*` de cadastro, sessão e recuperação de senha. `/healthz` é a liveness do processo: um método não contratado responde 405 e o corpo não consulta PostgreSQL. `/v1/videos` é a listagem pública; se o banco não atender, a resposta é 503 e não uma página vazia. As rotas de conta também respondem 503 sem banco ou sem segredo de assinatura, e não alteram `/healthz`.

Readiness das dependências ainda não tem endpoint. Quando o catálogo passar a depender do banco, ela entra em outro caminho e nesta spec, sem reutilizar `{"status":"ok"}` como prova de que o banco está pronto.

Adicionar rotas novas junto da mudança correspondente. `api/events` e `api/proto` continuam vazios até serem necessários.
