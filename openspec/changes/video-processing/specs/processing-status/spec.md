# Spec Delta

## Purpose

Mostra à conta dona se o vídeo está na fila, em processamento, pronto ou falhou, e permite tentar de novo.

## ADDED Requirements

### Requirement: Leitura do dono
`GET /v1/videos/{id}/processing` com token do dono MUST responder 200 com `processing_status` e, se `FAILED`, um erro com `code` e `message`. O corpo MUST NOT incluir object key nem URL de mídia. Outra conta MUST receber 403. Id desconhecido MUST receber 404. Sem token válido, MUST responder 401.

#### Scenario: Dono consulta
- **WHEN** o dono pede o estado do próprio vídeo
- **THEN** a resposta traz `processing_status` e não traz object key

#### Scenario: Outra conta
- **WHEN** outra conta autenticada pede o estado
- **THEN** a API responde 403

### Requirement: Retry só na falha
`POST /v1/videos/{id}/retry-processing` do dono, com o vídeo `FAILED`, MUST responder 202, voltar o vídeo a `QUEUED`, incrementar a geração e gravar um novo `video.uploaded.v1`. MUST NOT alterar `publication_status`. Se o vídeo não está `FAILED`, MUST responder 409 e MUST NOT gravar evento. Outra conta MUST receber 403.

#### Scenario: Falha do dono
- **WHEN** o dono pede retry de um vídeo `FAILED`
- **THEN** a geração aumenta e existe um evento de upload novo

#### Scenario: Vídeo que não falhou
- **WHEN** o dono pede retry de um vídeo `QUEUED` ou `READY`
- **THEN** a API responde 409 e a geração não muda

### Requirement: Estado sem banco
Se o PostgreSQL não puder atender a leitura ou o retry, a API MUST responder 503. `GET /healthz` MUST continuar 200.

#### Scenario: Banco parado na leitura
- **WHEN** o banco não atende a consulta de estado
- **THEN** a resposta é 503 e `GET /healthz` continua 200
