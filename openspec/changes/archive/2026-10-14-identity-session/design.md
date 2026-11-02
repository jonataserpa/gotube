# Design

## Context

A API usa `ServeMux`, envelope `{"error":{"code","message"}}` e migration explícita em `go run ./cmd/migrate`. `GET /healthz` não consulta o banco (ADR 0001). `GET /v1/videos` já depende do PostgreSQL e responde 503 sem virar readiness. O shell Obsidian (ADR 0002) deixa o controle Conta desligado até a etapa que o implementa. O catálogo é lido no servidor por `GO_API_INTERNAL_URL`. Não há usuário, cookie nem SMTP.

O curso (M12 e o contrato da seção 6) pede access token curto e refresh opaco. Também cita o roteador chi. Esta mudança permanece no `ServeMux`.

## Goals / Non-Goals

**Goals:**

- Conta, login, refresh, logout e reset testáveis sem canal e sem e-mail.
- O navegador guarda o refresh só em cookie `HttpOnly` da origem do Next. O access token não entra em `localStorage` nem em cookie.
- Rotação concorrente não apaga a sessão que venceu.

**Non-Goals:**

- Readiness HTTP. ADR 0001 previa outro caminho quando uma rota dependesse do banco; o catálogo já cobriu esse caso com 503 na própria rota. Identidade repete esse padrão.
- Perfil, canal, dono em `channels`, OAuth, verificação de e-mail e limite distribuído de tentativas.
- Seed com senha. Os testes criam as contas.

## Decisions

### Rotas e códigos

Prefixo `/v1/auth`. O contrato de teste que hoje proíbe `/auth/` passa a permitir só esse prefixo e continua proibindo `/watch`, `/playback` e `/upload`.

| Rota | Sucesso | Erro de negócio |
| --- | --- | --- |
| `POST /register` | 201 `{id,email}` | 400 `VALIDATION_ERROR`, 409 `EMAIL_ALREADY_REGISTERED` |
| `POST /login` | 200 `{access_token,token_type,expires_in,user}` | 401 `INVALID_CREDENTIALS` |
| `POST /refresh` | 200 no mesmo formato do login | 401 `INVALID_REFRESH` |
| `POST /logout` | 204 | — |
| `POST /password/forgot` | 202 `{message}` | — |
| `POST /password/reset` | 204 | 400 `INVALID_RESET_TOKEN` ou `VALIDATION_ERROR` |

Banco indisponível: 503 `DATABASE_UNAVAILABLE`. Segredo de assinatura ausente: 503 `AUTH_UNAVAILABLE`. Nos dois casos o corpo não confirma conta criada, login ou senha alterada. `GET /healthz` segue 200.

Mensagens em português, no mesmo envelope do catálogo, sem `request_id`. O curso mostra `request_id`; o contrato já publicado não tem esse campo, e esta mudança não o introduz.

E-mail: `TrimSpace` e minúsculas. Senha: mínimo 8 caracteres. Campo desconhecido ou tipo errado é 400. Cadastro e login não criam sessão no cadastro.

### Senha e tokens

Hash Argon2id (`golang.org/x/crypto/argon2`), parâmetros OWASP: memória 19 MiB, 2 iterações, paralelismo 1, salt de 16 bytes. Alternativa bcrypt foi descartada porque o curso pede hash adaptativo com parâmetros revisáveis.

Access token: JWT HS256 (`github.com/golang-jwt/jwt/v5`), 15 minutos. Claims `sub`, `email`, `iss=gotube`, `aud=gotube-web`, `iat`, `exp`. A verificação recusa outro algoritmo, emissor ou audiência. Segredo em `AUTH_JWT_SECRET`. Não vai para o Git; `.env.example` ganha um valor só de laboratório.

Refresh: 32 bytes aleatórios, guardados como SHA-256. Cookie `gotube_refresh`, `HttpOnly`, `SameSite=Lax`, `Path=/api/auth`, 30 dias. `Secure` quando `AUTH_COOKIE_SECURE=true` ou a requisição é HTTPS; no HTTP local fica desligado para o navegador gravar o cookie. O valor não aparece no JSON.

Reset: 32 bytes aleatórios, hash SHA-256, prazo de 30 minutos, um uso. Novo pedido invalida resets ainda não usados da mesma conta. O 202 padrão não traz o token. `AUTH_EXPOSE_RESET_TOKEN=true` inclui `reset_token` só se a conta existe. O exemplo versionado deixa esse sinal desligado.

### Rotação e reuso

Uma transação faz `SELECT … FOR UPDATE` pelo hash do refresh vigente, marca `revoked_at` e insere outra linha com o mesmo `family_id`.

O prazo de corrida é 10 segundos. Se o token apresentado já está revogado e o sucessor da família foi criado dentro desses 10 segundos, a resposta é 401 e a família permanece. Depois dos 10 segundos, a mesma apresentação revoga a família inteira. Cookie desconhecido é 401 sem tocar em outra família.

Logout revoga só a linha daquele hash. Reset revoga todas as linhas da conta.

### Tabelas

Migration nova, separada do catálogo. `channels` não ganha `owner_id`.

- `users`: `id`, `email_normalized` único, `password_hash`, `status` (`ACTIVE` nesta etapa), `created_at`.
- `refresh_sessions`: `id`, `user_id`, `token_hash` único, `family_id`, `expires_at`, `revoked_at`, `created_at`.
- `password_resets`: `id`, `user_id`, `token_hash` único, `expires_at`, `used_at`.

A API não aplica migration na subida. O comando de seed existente não cria usuário.

### Borda Next

O navegador chama rotas de mesma origem em `web/src/app/api/auth/*`. O servidor encaminha para `GO_API_INTERNAL_URL` e repassa `Set-Cookie`. O access token fica numa variável de módulo, fora de `localStorage` e de cookie. Ao hidratar, um único refresh em voo restaura a sessão; chamadas seguintes esperam essa promessa.

`SameSite=Lax` impede o POST cruzado de levar o cookie. Se o cabeçalho `Origin` vier preenchido e não for a origem da página, a rota do Next responde 403 sem chamar a API.

Páginas `/login`, `/register`, `/forgot-password` e `/reset-password` (o token vai na query). Copy em português. O controle Conta vira link “Entrar” sem sessão e, com sessão, mostra o e-mail e “Sair”. Busca, microfone, Novo vídeo, notificações e o restante da lateral seguem desligados, como na ADR 0002. A ADR dizia que a área principal anunciava catálogo ausente; isso já foi substituído pela grade e não volta.

O catálogo continua no servidor, sem cookie e sem o e-mail no card.

### Purpose no archive

O delta de capacidade existente não carrega Purpose. Ao arquivar, o Purpose consolidado de `public-shell` deve passar a dizer que a superfície tem navegação, tema e sessão real quando existir, sem vídeo ou busca inventados.

## Risks / Trade-offs

- [Access token segue válido até expirar depois do logout] → Prazo de 15 minutos. Não há lista de revogação nesta etapa; a spec registra isso.
- [Replay do refresh antigo dentro de 10 segundos não mata a família] → Também não emite sessão nova. Fora da janela, a família cai.
- [Sem SMTP o visitante não recebe o link] → A página de reset aceita o token na URL. O laboratório só vê o token com o sinal explícito.
- [Cookie sem `Secure` no HTTP local] → Ligado em HTTPS ou por `AUTH_COOKIE_SECURE`. O exemplo não liga a exposição do reset.
- [Argon2id deixa o teste de senha mais lento] → Parâmetros OWASP, não um perfil de 64 MiB.

## Migration Plan

1. Publicar a migration e as rotas. Sem `DATABASE_URL` ou sem `AUTH_JWT_SECRET`, a API sobe e as rotas de conta respondem 503.
2. O volume do Compose permanece. Rollback é remover as rotas e parar de aplicar a migration nova; as tabelas de identidade podem ficar sem uso.
3. Não há dado de produção para converter.

## Open Questions

Nenhuma que mude a spec, o desenho ou as tarefas. Prazo de 15 minutos, janela de 10 segundos e ausência de SMTP ficam como decisões acima.
