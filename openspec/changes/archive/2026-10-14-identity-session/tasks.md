# Tasks

## 1. Schema

- [x] 1.1 Criar a migration de `users`, `refresh_sessions` e `password_resets`, sem `owner_id` em `channels` e sem senha no seed; verificar que `go run ./cmd/migrate` aplica o arquivo num PostgreSQL local e que o boot da API não executa migration
- [x] 1.2 Documentar `AUTH_JWT_SECRET`, `AUTH_COOKIE_SECURE` e `AUTH_EXPOSE_RESET_TOKEN` só em `.env.example`, com o sinal de reset desligado e sem segredo de produção; verificar que `.env` e `web/.env` não foram alterados

## 2. Cadastro

- [x] 2.1 Escrever o teste HTTP que prova 201 sem cookie nem hash, 409 para o mesmo e-mail com outra caixa, 400 para senha curta ou campo desconhecido, e 503 sem banco com `GET /healthz` ainda 200; verificar que falha antes do cadastro existir
- [x] 2.2 Implementar `POST /v1/auth/register` com e-mail normalizado e Argon2id; verificar que o teste de 2.1 passa

## 3. Login

- [x] 3.1 Escrever o teste de login que exige access token e e-mail no JSON, refresh só em cookie `HttpOnly`, e o mesmo 401 para e-mail inexistente, senha errada e conta inativa; verificar que falha antes do login existir
- [x] 3.2 Implementar `POST /v1/auth/login` com JWT HS256 de 15 minutos; verificar que o teste de 3.1 passa e que o corpo não contém o valor do refresh

## 4. Refresh e logout

- [x] 4.1 Escrever os testes de rotação, de duas renovações simultâneas em que a perdedora não revoga a vencedora, de reuso depois de 10 segundos revogando a família, de cookie desconhecido sem afetar outra conta, e de logout 204 só da sessão apresentada; verificar que falham antes da rotação existir
- [x] 4.2 Implementar `POST /v1/auth/refresh` e `POST /v1/auth/logout` com a transação e o prazo de corrida do design; verificar que os testes de 4.1 passam

## 5. Recuperação

- [x] 5.1 Escrever os testes do 202 idêntico com e sem conta, da ausência de token com o sinal desligado, da presença só com `AUTH_EXPOSE_RESET_TOKEN` e conta existente, do reset que consome o token e revoga os refresh da conta, do segundo uso recusado e da senha inválida que não consome o token; verificar que falham antes da recuperação existir
- [x] 5.2 Implementar `POST /v1/auth/password/forgot` e `POST /v1/auth/password/reset`; verificar que os testes de 5.1 passam e que o 503 não afirma senha alterada

## 6. Contrato

- [x] 6.1 Acrescentar somente as rotas `/v1/auth/*` desta mudança em `api/openapi.yaml`; verificar que o teste de contrato permite esse prefixo, continua proibindo `/watch`, `/playback` e `/upload`, e que o lint do OpenAPI passa

## 7. Next.js

- [x] 7.1 Escrever testes do shell e das páginas: sem sessão o controle é o link Entrar; com sessão aparecem o e-mail e Sair; busca e Novo vídeo seguem indisponíveis; o card do catálogo não mostra o e-mail; verificar que falham enquanto Conta estiver desligado
- [x] 7.2 Implementar as rotas `web/src/app/api/auth/*`, a sessão em memória com um único refresh em voo, as páginas `/login`, `/register`, `/forgot-password` e `/reset-password`, e o controle no shell; verificar com os testes do `web/`, typecheck, lint e build, e que nenhum token usa `localStorage` nem `NEXT_PUBLIC_`

## 8. Integração

- [x] 8.1 Rodar `go test ./cmd/... ./internal/...` com `DATABASE_URL` do Compose; verificar cadastro, login, refresh, logout, reset e que `/healthz` segue 200 com o banco parado nas rotas de conta
- [x] 8.2 No navegador, cadastrar, entrar, ver o e-mail no shell com o catálogo público ainda carregado, sair, pedir recuperação e redefinir com o token do laboratório; registrar os resultados em `docs/learning/progress.md`

## Workflow follow-up

- Revisar esta proposta antes de implementar. A aplicação começa só com `/opsx-apply` ou pedido explícito.
- Ao arquivar, atualizar o Purpose consolidado de `public-shell`: navegação, tema e sessão real quando existir, sem vídeo ou busca inventados. O delta não carrega Purpose.
- `GET/PATCH /me`, canal e categorias ficam para a próxima etapa do roadmap.
