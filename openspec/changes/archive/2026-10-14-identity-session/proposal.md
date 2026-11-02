# Proposal

## Why

O visitante só navega como anônimo: o botão Conta está desligado e a API não tem cadastro nem sessão. A etapa 3 do roadmap pede conta própria, renovação, saída e recuperação, com isolamento entre sessões, antes de canal e perfil.

## What Changes

- Cadastro com e-mail normalizado e senha com hash adaptativo. O cadastro não abre sessão.
- Login devolve access token curto no corpo e grava refresh opaco em cookie `HttpOnly`. O access token fica só em memória no navegador.
- Refresh rotaciona a sessão de forma atômica. Reapresentar um refresh já rotacionado, depois do prazo curto de corrida, revoga a família. A apresentação perdedora dentro desse prazo recebe 401 e deixa a sessão nova ativa. Logout revoga só a sessão apresentada e limpa o cookie.
- Recuperação responde sempre de forma genérica. O reset consome um token de uso único, troca o hash e revoga os refresh daquela conta.
- A consulta pública de vídeos continua sem credencial. O catálogo não grava token.
- **BREAKING** para `public-shell`: sem sessão, a home deixa de proibir login funcional e passa a oferecer entrada real. Com sessão restaurada, o shell mostra só essa conta e a saída. Busca, Novo vídeo e os demais controles ainda inexistentes continuam desligados. Nada de usuário, vídeo ou canal inventado.

Fora deste escopo: `GET/PATCH /me`, canal, categorias, verificação de e-mail, OAuth, envio de e-mail, limite distribuído de tentativas, upload e qualquer recurso autenticado além da própria sessão. Sem SMTP, o token de reset não sai na resposta padrão; um sinal explícito de desenvolvimento pode devolvê-lo só no laboratório. A página de redefinição aceita o token na URL para um envio futuro.

## Capabilities

### New Capabilities

- `account-registration`: criar conta com e-mail único e senha, sem devolver hash nem abrir sessão.
- `refresh-session`: login, rotação, logout e isolamento entre contas e famílias de refresh.
- `password-recovery`: pedido genérico de recuperação e redefinição de uso único com revogação das sessões da conta.

### Modified Capabilities

- `public-shell`: o requisito que proíbe login funcional passa a exigir entrada e saída reais, sem sessão fictícia.
- `catalog-home`: a leitura pública continua sem credencial; a identidade visível, quando houver, fica no shell e não no card.

## Impact

- Go: rotas `/v1/auth/*` no `ServeMux` já existente, hash de senha, JWT de acesso e persistência de refresh. `GET /healthz` e `GET /v1/videos` não mudam de contrato.
- Next.js: páginas de cadastro, entrada, recuperação e redefinição; o controle Conta em `web/src/components/layout/shell-actions.tsx` deixa de ser um botão desligado. O navegador fala com a origem do Next; o servidor encaminha a API e o cookie.
- Dados: nova migration só com usuário, sessão de refresh e reset de senha. A tabela `channels` atual permanece sem dono.
- Operação: a API continua subindo sem `DATABASE_URL`; as rotas de conta respondem indisponibilidade. Não há fila, worker, storage nem Kubernetes. O contrato em `api/openapi.yaml` ganha só essas rotas de autenticação.
