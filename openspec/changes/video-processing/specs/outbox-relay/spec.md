# Spec Delta

## Purpose

Publica eventos já gravados na outbox e só os considera entregues depois que a fila confirma.

## ADDED Requirements

### Requirement: Confirmação antes da marcação
O relay MUST publicar o payload persistido e MUST gravar `published_at` só depois da confirmação da fila. A mensagem MUST NOT incluir bytes de vídeo nem URL escolhida pelo cliente.

#### Scenario: Fila confirma
- **WHEN** o relay publica uma linha sem `published_at` e a fila confirma
- **THEN** a linha fica marcada e a mensagem contém `video_id` e `generation`

#### Scenario: Fila indisponível
- **WHEN** a fila não confirma a publicação
- **THEN** `published_at` continua vazio

### Requirement: Republicação segura
Se o processo cair depois de publicar e antes de marcar, outra passagem MUST poder publicar de novo o mesmo evento sem apagar a linha.

#### Scenario: Queda na janela
- **WHEN** a fila já recebeu o evento e `published_at` ainda é nulo
- **THEN** uma passagem seguinte publica de novo essa linha

### Requirement: API sem a fila
`GET /healthz` da API MUST continuar 200 enquanto o relay não consegue publicar.

#### Scenario: Fila parada
- **WHEN** o relay não publica porque a fila não atende
- **THEN** `GET /healthz` responde 200
