# public-video-list Specification

## Purpose

Permite a um visitante listar somente vídeos publicados e públicos, em páginas estáveis, sem receber rascunho, vídeo privado ou dado de conta.

## Requirements

### Requirement: Listagem pública sem credencial
`GET /v1/videos` MUST responder sem credencial. Cada item MUST ter visibilidade pública e publicação publicada. Rascunho, vídeo privado, não listado, em exclusão, excluído ou bloqueado MUST NOT aparecer.

#### Scenario: Visitante sem sessão
- **WHEN** um cliente chama `GET /v1/videos` sem autorização
- **THEN** a resposta de sucesso contém apenas vídeos públicos publicados

#### Scenario: Rascunho e privado no banco
- **WHEN** o banco contém um vídeo público publicado, um rascunho e um vídeo privado
- **THEN** a lista inclui somente o vídeo público publicado

### Requirement: DTO público
Cada item MUST trazer identificador, título, duração em segundos, nome e handle públicos do canal. A thumbnail MUST ser uma URL ou ausente. A resposta MUST NOT incluir e-mail, segredo, status de processamento, visibilidade interna nem capacidade de edição.

#### Scenario: Item publicado
- **WHEN** a lista devolve um vídeo público publicado
- **THEN** o item contém título, duração, canal público e thumbnail presente ou ausente, sem campos de conta

### Requirement: Cursor estável
A ordem MUST ser data de publicação descendente e, na mesma data, identificador descendente. `limit` ausente MUST valer 12. `limit` acima de 50 ou inválido MUST produzir 400. O cursor MUST ser opaco. A página seguinte MUST continuar depois do último item sem repetir nem pular empates.

#### Scenario: Empate de data
- **WHEN** dois vídeos públicos publicados compartilham a data de publicação e cabem em páginas diferentes
- **THEN** ambos aparecem, o de identificador maior vem antes, e o cursor não repete nem omite um deles

#### Scenario: Cursor inválido
- **WHEN** o cliente envia um cursor que o contrato não reconhece
- **THEN** a resposta é 400 e não uma lista parcial

#### Scenario: Limite acima do máximo
- **WHEN** o cliente pede `limit` maior que 50
- **THEN** a resposta é 400

### Requirement: Banco indisponível
Se o PostgreSQL não puder atender a listagem, a API MUST responder 503. Esse corpo MUST NOT ser uma lista vazia. `GET /healthz` MUST continuar respondendo liveness sem consultar o banco.

#### Scenario: Falha de banco na listagem
- **WHEN** o PostgreSQL está indisponível e o cliente chama `GET /v1/videos`
- **THEN** a resposta é 503 e o corpo não é uma página de itens

#### Scenario: Liveness sem banco
- **WHEN** o PostgreSQL está indisponível e o cliente chama `GET /healthz`
- **THEN** a resposta continua 200 com `status` igual a `ok`

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

### Requirement: Seed de desenvolvimento identificado
O conjunto de dados local de exemplo MUST declarar-se como desenvolvimento. Ele MUST conter ao menos um vídeo público publicado e ao menos um vídeo que a listagem exclui.

#### Scenario: Seed aplicado
- **WHEN** o seed de desenvolvimento é aplicado num banco local vazio
- **THEN** a listagem mostra o vídeo público e omite o vídeo excluído pelo filtro, e o seed está marcado como dado de desenvolvimento
