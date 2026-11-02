# Spec Delta

## Purpose

Permite à conta com canal criar um rascunho privado que ainda não entrou no catálogo público.

## ADDED Requirements

### Requirement: Rascunho no próprio canal
`POST /v1/videos` com access token válido e título não vazio MUST criar um vídeo `DRAFT`, `PRIVATE` e `WAITING_UPLOAD` no canal dessa conta e responder 201. Sem canal da conta, a API MUST responder 409 e MUST NOT criar vídeo. Sem token válido, MUST responder 401. Título vazio ou campo desconhecido MUST responder 400 e MUST NOT criar vídeo.

#### Scenario: Conta com canal
- **WHEN** a conta autenticada que já tem canal envia um título
- **THEN** a API responde 201 com o identificador de um rascunho privado desse canal

#### Scenario: Conta sem canal
- **WHEN** a conta autenticada ainda não tem canal tenta criar um vídeo
- **THEN** a API responde 409 e não existe vídeo novo

#### Scenario: Título em branco
- **WHEN** o título contém apenas espaços
- **THEN** a API responde 400 e não cria vídeo

### Requirement: Rascunho fora da listagem pública
O rascunho criado MUST NOT aparecer em `GET /v1/videos`. O canal de laboratório sem dono MUST NOT receber vídeo por essa rota.

#### Scenario: Lista pública depois do rascunho
- **WHEN** a conta cria um rascunho e um visitante pede a listagem pública
- **THEN** a página não contém esse rascunho

### Requirement: Rascunho sem banco
Se o PostgreSQL não puder atender a criação, a API MUST responder 503. O corpo MUST NOT afirmar que o vídeo foi criado. `GET /healthz` MUST continuar 200.

#### Scenario: Banco parado na criação
- **WHEN** o banco não atende a criação do rascunho
- **THEN** a resposta é 503 e a conta continua sem esse vídeo
