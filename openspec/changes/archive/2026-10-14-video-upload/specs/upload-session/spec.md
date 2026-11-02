# Spec Delta

## Purpose

Autoriza um envio direto ao storage e só pede processamento depois de verificar o objeto, uma vez por sessão.

## ADDED Requirements

### Requirement: Autorização do próprio rascunho
`POST /v1/videos/{id}/upload-sessions` com access token do dono MUST responder com autorização temporária para uma object key gerada no servidor, com prazo e tamanho máximo. Outra conta MUST receber 403. Identificador desconhecido MUST receber 404. Vídeo que já não está `WAITING_UPLOAD` MUST receber 409. Key ou URL enviada pelo cliente MUST receber 400 e MUST NOT ser usada.

#### Scenario: Dono pede autorização
- **WHEN** o dono pede uma sessão para o próprio rascunho
- **THEN** a resposta traz autorização limitada a uma key que o cliente não escolheu

#### Scenario: Outra conta
- **WHEN** outra conta autenticada pede sessão para esse vídeo
- **THEN** a API responde 403 e não emite autorização

#### Scenario: Key do cliente
- **WHEN** o corpo tenta informar a object key
- **THEN** a API responde 400 e não grava essa key

### Requirement: Conclusão verificada
`POST /v1/upload-sessions/{id}/complete` MUST consultar o objeto no storage. Objeto presente e dentro do tamanho autorizado MUST responder 202, marcar o vídeo `QUEUED` e gravar um único evento `video.uploaded.v1`. Objeto ausente ou maior que o máximo MUST responder 409 e MUST deixar o vídeo em `WAITING_UPLOAD`.

#### Scenario: Objeto aceito
- **WHEN** o dono completa a sessão e o objeto existe dentro do limite
- **THEN** a resposta é 202, o vídeo fica `QUEUED` e existe um evento de upload

#### Scenario: Objeto ausente
- **WHEN** o dono completa a sessão e o storage não tem o objeto
- **THEN** a API responde 409 e o vídeo continua `WAITING_UPLOAD`

### Requirement: Conclusão idempotente
Completar de novo a mesma sessão já concluída MUST responder 202 e MUST NOT gravar outro evento nem mudar o vídeo outra vez.

#### Scenario: Segunda confirmação
- **WHEN** o dono completa outra vez a sessão que já respondeu 202
- **THEN** a resposta é 202 e continua existindo um único evento de upload

### Requirement: Sessão sem banco ou storage
Se o PostgreSQL ou o storage não puder atender a autorização ou a conclusão, a API MUST responder 503. O corpo MUST NOT afirmar que o vídeo foi enfileirado. `GET /healthz` MUST continuar 200.

#### Scenario: Storage parado na conclusão
- **WHEN** o storage não atende a verificação do objeto
- **THEN** a resposta é 503 e o vídeo continua `WAITING_UPLOAD`
