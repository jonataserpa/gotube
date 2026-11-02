# Design

## Context

O starter em `template/` foi materializado na raiz porque o pacote foi aberto como o próprio repositório. Não há `go.mod`, app Next nem `api/openapi.yaml`. `docs/adr/` só explica como registrar decisões; ainda não há ADR. A porta `127.0.0.1:8080` já está ocupada nesta máquina. Ver `proposal.md` para a motivação.

## Goals / Non-Goals

**Goals:**

- Um processo Go que valida configuração, escuta HTTP com prazos e para com `Shutdown`.
- Um shell Next que hidrata sem divergir o texto. O aviso de catálogo ausente não permanece como requisito; a listagem é catalog-home.
- Compose de PostgreSQL utilizável na próxima etapa, desacoplado da liveness.
- Contrato, testes e CI alinhados ao que realmente existe.

**Non-Goals:**

- Pool, migration, sqlc ou qualquer query.
- Proxy `/v1`, CORS, cliente HTTP no browser e sessão.
- chi, Viper, Redis, fila, storage, Kubernetes.

## Decisions

### Módulo provisório e biblioteca padrão

O módulo é `example.com/gotube`. Trocar depois significa editar a linha `module` e os imports, depois `go mod tidy`, quando existir um caminho de repositório real. O HTTP fica em `net/http` com padrão de método do ServeMux. chi entra no módulo de API do curso, não aqui: o bootstrap não tem grupo de rotas, middleware de auth nem paginação.

Alternativa rejeitada: esperar o remoto GitHub para escolher o module path. O prompt exige um path local agora.

### Liveness não é readiness

`GET /healthz` devolve `{"status":"ok"}` se o processo aceita a requisição. Não abre banco. Readiness futura, quando existir dependência obrigatória, será outro caminho e não reutilizará este corpo para dizer que o PostgreSQL está pronto. A decisão duradoura fica em `docs/adr/0001-liveness-before-readiness.md`. Não há ADR anterior para divergir.

### Configuração falha no boot

Variáveis de ambiente, sem biblioteca de config. `API_ADDR` é obrigatório e precisa ser `host:port` ou `:port`, nunca uma URL. Prazos ausentes recebem default positivo; prazo explícito inválido ou não positivo é erro. `main` retorna esse erro e não chama `Listen`. PostgreSQL não é configuração da API nesta etapa: as variáveis `POSTGRES_*` pertencem só ao Compose.

### Shutdown cooperativo

`signal.NotifyContext` observa SIGINT e SIGTERM. O dono da goroutine de `ListenAndServe` é `Run`. No cancelamento, `http.Server.Shutdown` usa o prazo configurado e o processo espera a goroutine terminar. Context de request não carrega dependências. Não há goroutine de trabalho de negócio.

### Frontend gerado e ilhas pequenas

`create-next-app` estável, compatível com o Node instalado, em diretório temporário; os arquivos entram em `web/` sem substituir `web/AGENTS.md`. Tailwind permanece o que o gerador emitir. A página inicial é Server Component. O único Client Component do shell é o controle de tema. Um script inline, antes da hidratação, aplica a classe `dark` a partir de `localStorage` ou de `prefers-color-scheme`. O botão não lê armazenamento durante o render, então o texto do HTML não muda entre servidor e cliente. `suppressHydrationWarning` fica só no elemento `html`, que o script altera.

A navegação contém Início. Explorar, inscrições e biblioteca não viram links mortos.

### PostgreSQL à parte

Compose em `deploy/compose/compose.yaml`, imagem com patch fixado, volume nomeado e healthcheck `pg_isready`. A API não recebe DSN. Credenciais de exemplo continuam em `.env.example`; `.env` fica de fora do Git.

### Contrato e verificação

`api/openapi.yaml` descreve apenas `GET /healthz`. Testes Go cobrem handler, configuração inválida e parada. Vitest cobre o shell e a página. CI repete `gofmt`, `go vet`, `go test` e os scripts reais do `web/package.json`.

## Risks / Trade-offs

- [8080 ocupada nesta máquina] → o default do exemplo continua `:8080`; a verificação local usa outra porta e o README explica o conflito.
- [Tema aplicado por script pode divergir da classe no `html`] → o aviso de hidratação é suprimido só nesse elemento; o teste e o navegador conferem o texto, não a classe.
- [Banco sobe sem consumidor] → o volume existe para a próxima etapa; a spec impede usar `/healthz` como readiness.
- [Node local pode ser mais novo que o LTS do CI] → o workflow usa uma versão que o Next instalado declara suportada e o README registra a versão que gerou o lockfile.
- [OpenSpec 1.14.1 reporta avisos de auditoria npm no tooling] → o CLI fica isolado em `.tooling/openspec`; não faz parte do runtime da aplicação.

## Migration Plan

Não há deploy. Rollback desta etapa é não executar os processos. O volume Docker só deve ser apagado com comando explícito de volume, documentado como destrutivo.

## Open Questions

Nenhuma. Versões resolvidas na implementação: Go 1.27, Next.js 16.4.0, React 19.3.0, Tailwind 4, Node `>=20.9.0` (local 26.8.2, CI 24), PostgreSQL `18.1-alpine`, OpenSpec 1.14.1.
