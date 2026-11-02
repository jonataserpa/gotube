# Spec Delta

## ADDED Requirements

### Requirement: Filtro pelo canal
`GET /v1/videos` com `channel_id` reconhecido MUST devolver só vídeos públicos publicados daquele canal, na mesma ordem e com o mesmo cursor. Sem `channel_id`, a lista MUST continuar incluindo os canais públicos publicados que cabem na página. `channel_id` que o contrato não reconhece MUST responder 400. Um identificador bem formado sem vídeos públicos publicados MUST responder 200 com página vazia, não 404.

#### Scenario: Só o canal pedido
- **WHEN** dois canais têm vídeo público publicado e a lista pede o identificador de um deles
- **THEN** a página contém somente vídeos públicos publicados desse canal

#### Scenario: Rascunho do canal pedido
- **WHEN** o canal pedido tem um vídeo público publicado e um rascunho
- **THEN** a página inclui o publicado e omite o rascunho

#### Scenario: Identificador mal formado
- **WHEN** `channel_id` não é um identificador que o contrato reconhece
- **THEN** a resposta é 400 e não uma lista parcial
