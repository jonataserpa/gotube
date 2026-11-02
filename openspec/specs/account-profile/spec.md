# account-profile Specification

## Purpose

Permite à conta autenticada consultar e atualizar o próprio nome de exibição, sem mudar e-mail nem senha.

## Requirements

### Requirement: Leitura do próprio perfil
`GET /v1/me` com access token válido MUST responder 200 com o identificador, o e-mail e o nome de exibição da conta desse token. O corpo MUST NOT incluir hash de senha nem refresh. Sem token, token inválido ou token expirado, a API MUST responder 401 e MUST NOT devolver dados de conta.

#### Scenario: Token da própria conta
- **WHEN** a conta apresenta um access token válido
- **THEN** a resposta traz o identificador, o e-mail e o nome de exibição dessa conta, sem hash

#### Scenario: Token ausente
- **WHEN** a requisição não traz access token
- **THEN** a API responde 401 e o corpo não contém e-mail

### Requirement: Nome de exibição editável
`PATCH /v1/me` com access token válido e nome de exibição não vazio MUST gravar esse nome e responder 200. Nome ausente, só com espaços ou campo que o perfil não declara MUST responder 400 e MUST NOT alterar o nome já gravado. O e-mail MUST permanecer o da conta.

#### Scenario: Nome válido
- **WHEN** a conta envia um nome de exibição com texto
- **THEN** a leitura seguinte devolve esse nome e o mesmo e-mail

#### Scenario: Nome em branco
- **WHEN** o nome de exibição contém apenas espaços
- **THEN** a API responde 400 e o nome anterior permanece

### Requirement: Perfil sem banco
Se o PostgreSQL não puder atender a leitura ou a edição, a API MUST responder 503. O corpo MUST NOT afirmar que o nome mudou. `GET /healthz` MUST continuar 200.

#### Scenario: Banco parado na edição
- **WHEN** o banco não atende a edição do perfil
- **THEN** a resposta é 503 e o nome anterior permanece
