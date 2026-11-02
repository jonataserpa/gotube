# account-registration Specification

## Purpose

Cria uma conta com e-mail único e senha, sem devolver o hash e sem abrir sessão.

## Requirements

### Requirement: Cadastro sem sessão
`POST /v1/auth/register` com e-mail e senha válidos MUST criar a conta e responder 201. A resposta MUST trazer o identificador e o e-mail normalizado. Ela MUST NOT trazer access token, cookie de refresh nem o hash da senha.

#### Scenario: Conta nova
- **WHEN** o visitante envia e-mail e senha que atendem à política
- **THEN** a API responde 201 sem sessão e o corpo não contém o hash

### Requirement: E-mail único normalizado
O e-mail MUST ser comparado sem diferença de maiúsculas e sem espaços nas pontas. Um e-mail já cadastrado MUST responder 409 e MUST NOT criar outra conta.

#### Scenario: Repetição com outra caixa
- **WHEN** já existe conta para um e-mail e o cadastro repete esse e-mail com outra caixa ou espaços nas pontas
- **THEN** a API responde 409 e permanece uma única conta

### Requirement: Senha e corpo recusados
Senha com menos de 8 caracteres, senha ausente, tipo inválido ou campo que o cadastro não declara MUST responder 400. A conta MUST NOT ser gravada.

#### Scenario: Senha curta
- **WHEN** a senha tem menos de 8 caracteres
- **THEN** a API responde 400 e a conta não existe

#### Scenario: Campo desconhecido
- **WHEN** o corpo inclui um campo que o cadastro não declara
- **THEN** a API responde 400 e não cria conta

### Requirement: Cadastro sem banco
Se o PostgreSQL não puder gravar a conta, a API MUST responder 503. O corpo MUST NOT afirmar que a conta existe. `GET /healthz` MUST continuar 200.

#### Scenario: Banco parado
- **WHEN** o banco não atende o cadastro
- **THEN** a resposta é 503 e a liveness permanece `ok`
