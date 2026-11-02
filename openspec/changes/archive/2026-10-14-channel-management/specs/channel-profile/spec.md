# Spec Delta

## Purpose

Permite à conta criar e editar um único canal e a qualquer visitante ler os metadados públicos desse canal pelo handle.

## ADDED Requirements

### Requirement: Ausência explícita do canal
`GET /v1/me/channel` com access token válido MUST responder 200. Sem canal da conta, o corpo MUST declarar ausência e MUST NOT inventar handle. Com canal, o corpo MUST trazer identificador, handle, nome e descrição. Sem token válido, a API MUST responder 401.

#### Scenario: Conta sem canal
- **WHEN** a conta autenticada ainda não criou canal
- **THEN** a resposta é 200 e declara que não há canal

#### Scenario: Conta com canal
- **WHEN** a conta autenticada já tem canal
- **THEN** a resposta traz o handle, o nome e a descrição desse canal

### Requirement: Um canal por conta
`POST /v1/channels` com access token válido, handle aceito e nome não vazio MUST criar o canal dessa conta e responder 201. A conta que já tem canal MUST receber 409 e MUST NOT criar outro. Handle já usado, ignorando maiúsculas e espaços nas pontas, MUST responder 409. Handle vazio, só com espaços ou com caractere fora de letras, dígitos e hífen MUST responder 400 e MUST NOT criar canal.

#### Scenario: Primeiro canal
- **WHEN** a conta sem canal envia handle e nome válidos
- **THEN** a API responde 201 e a leitura da conta passa a devolver esse canal

#### Scenario: Segundo canal
- **WHEN** a conta que já tem canal tenta criar outro
- **THEN** a API responde 409 e permanece um único canal

#### Scenario: Handle repetido
- **WHEN** outra conta envia um handle já usado, com outra caixa ou espaços nas pontas
- **THEN** a API responde 409 e não cria canal

### Requirement: Edição só do dono
`PATCH /v1/channels/{id}` MUST alterar nome, handle ou descrição somente quando o token é da conta dona. Canal de outra conta ou canal sem dono MUST responder 403 e MUST NOT mudar. Identificador desconhecido MUST responder 404. Campo desconhecido ou nome vazio MUST responder 400 e MUST NOT gravar a alteração.

#### Scenario: Dono edita o nome
- **WHEN** o dono envia um nome novo válido
- **THEN** a leitura pública passa a mostrar esse nome

#### Scenario: Outra conta
- **WHEN** outra conta autenticada envia alteração para esse canal
- **THEN** a API responde 403 e o nome público permanece

#### Scenario: Canal sem dono
- **WHEN** uma conta autenticada tenta editar o canal de laboratório que não tem dono
- **THEN** a API responde 403 e os metadados públicos permanecem

### Requirement: Leitura pública por handle
`GET /v1/channels/{handle}` MUST responder sem credencial. O corpo MUST trazer identificador, handle, nome e descrição públicos. Ele MUST NOT trazer e-mail, identificador da conta nem capacidade de edição. Handle desconhecido MUST responder 404.

#### Scenario: Visitante
- **WHEN** um cliente sem sessão pede um handle existente
- **THEN** a resposta traz os metadados públicos e não contém e-mail

#### Scenario: Handle inexistente
- **WHEN** o handle não corresponde a nenhum canal
- **THEN** a API responde 404

### Requirement: Canal sem banco
Se o PostgreSQL não puder atender criação, edição ou leitura do canal, a API MUST responder 503. O corpo MUST NOT afirmar que o canal foi criado ou alterado. `GET /healthz` MUST continuar 200.

#### Scenario: Banco parado na criação
- **WHEN** o banco não atende a criação do canal
- **THEN** a resposta é 503 e a conta continua sem canal
