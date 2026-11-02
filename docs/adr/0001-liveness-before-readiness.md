# ADR 0001 — Liveness antes de readiness

- Status: aceito em 14/10/2026, durante bootstrap-project.
- Spec: `openspec/changes/bootstrap-project/specs/api-liveness/spec.md`.

## Contexto

A API ainda não consulta PostgreSQL. Um único endpoint de saúde que abre conexão com o banco misturaria “o processo está vivo” com “a dependência está pronta”. O curso separa ciclo de vida do processo (M09) de pool e transação (M11), que entram depois.

## Alternativas

- Não expor saúde: o processo sobe, mas não há sinal observável para teste e operação.
- Fazer `/healthz` consultar o banco: falha o boot da etapa atual quando o Compose está parado e transforma liveness em readiness cedo demais.

## Decisão

`GET /healthz` é liveness. Responde 200 e `{"status":"ok"}` quando o processo HTTP aceita a requisição. Não abre banco, não exige credencial e não promete que uma dependência está pronta.

Readiness, quando uma rota de negócio passar a depender do PostgreSQL, será outro caminho. Não se reutiliza este corpo para dizer que o banco está saudável. A checagem do container continua no healthcheck do Compose.

## Consequências

Um orquestrador que usar `/healthz` como readiness enviará tráfego mesmo com o banco parado. Isso é aceitável enquanto a API não precisa do banco para cumprir o contrato. A revisão acontece quando o catálogo passar a exigir PostgreSQL para responder.

## Critério de revisão

Existe uma rota implementada cuja resposta correta depende do banco. Nesse momento, readiness deixa de ser futura e ganha contrato próprio.
