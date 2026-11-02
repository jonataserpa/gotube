# password-recovery Specification

## Purpose

Permite pedir e concluir a troca de senha com token de uso único, sem revelar se o e-mail existe.

## Requirements

### Requirement: Pedido genérico
`POST /v1/auth/password/forgot` MUST responder 202 com o mesmo corpo público para e-mail cadastrado e não cadastrado. A resposta padrão MUST NOT incluir o token nem revelar se a conta existe. E-mail desconhecido MUST NOT gravar token.

#### Scenario: E-mail desconhecido
- **WHEN** o e-mail não tem conta
- **THEN** a API responde 202 com o mesmo corpo público de um e-mail cadastrado e não cria token

#### Scenario: E-mail cadastrado
- **WHEN** o e-mail tem conta e o sinal de laboratório está desligado
- **THEN** a API responde 202 sem o token e grava um reset de uso único ainda não consumido

### Requirement: Token só no laboratório
Somente um sinal explícito de desenvolvimento MUST incluir o token de reset no corpo do 202, e só quando a conta existe. Com o sinal desligado, o corpo MUST NOT trazê-lo.

#### Scenario: Sinal desligado
- **WHEN** o pedido é feito sem o sinal de laboratório
- **THEN** o 202 não contém token

#### Scenario: Sinal ligado
- **WHEN** o sinal de laboratório está ligado e o e-mail tem conta
- **THEN** o 202 inclui o token de uso único

### Requirement: Reset de uso único
`POST /v1/auth/password/reset` com token vigente e senha válida MUST trocar a senha, consumir o token e revogar todos os refresh daquela conta. O segundo uso do mesmo token MUST falhar e MUST NOT trocar a senha outra vez.

#### Scenario: Token válido
- **WHEN** o token ainda não foi usado nem expirou e a senha nova é válida
- **THEN** a senha muda, o token fica consumido e os refresh da conta ficam revogados

#### Scenario: Token repetido
- **WHEN** o mesmo token é apresentado outra vez
- **THEN** a API recusa e a senha permanece a que o primeiro uso gravou

### Requirement: Reset recusado
Token ausente, expirado ou desconhecido MUST ser recusado e MUST NOT alterar a senha. Senha inválida com token vigente MUST NOT consumir o token.

#### Scenario: Token expirado
- **WHEN** o token passou do prazo
- **THEN** a API recusa e a senha não muda

#### Scenario: Senha inválida
- **WHEN** o token está vigente e a senha nova é recusada
- **THEN** a senha não muda e o token continua utilizável

### Requirement: Reset sem banco
Se o PostgreSQL não puder atender o pedido ou o reset, a API MUST responder 503. O corpo MUST NOT afirmar que a senha mudou. `GET /healthz` MUST continuar 200.

#### Scenario: Banco parado no reset
- **WHEN** o banco não atende o reset
- **THEN** a resposta é 503 e a senha anterior permanece
