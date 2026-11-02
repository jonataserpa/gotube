# active-categories Specification

## Purpose

Lista as categorias ativas e só deixa o dono associá-las ao próprio canal, recusando categoria inativa ou desconhecida.

## Requirements

### Requirement: Lista só de categorias ativas
`GET /v1/categories` MUST responder sem credencial. Cada item MUST estar ativo e MUST trazer identificador, slug e nome. Categoria inativa MUST NOT aparecer.

#### Scenario: Catálogo misto
- **WHEN** existem uma categoria ativa e uma inativa
- **THEN** a lista contém a ativa e omite a inativa

### Requirement: Associação só com categorias ativas
Criar ou editar o próprio canal com uma lista de categorias MUST gravar exatamente esse conjunto quando todos os identificadores estão ativos. Lista vazia MUST remover as associações. Identificador desconhecido ou inativo MUST responder 400 e MUST NOT alterar o conjunto já gravado. A resposta do canal MUST incluir as categorias ativas associadas.

#### Scenario: Conjunto válido
- **WHEN** o dono envia somente identificadores de categorias ativas
- **THEN** a leitura do canal devolve esse conjunto e nenhuma outra categoria

#### Scenario: Categoria inativa
- **WHEN** a lista inclui uma categoria inativa
- **THEN** a API responde 400 e as categorias anteriores do canal permanecem

### Requirement: Associação só do dono
Outra conta autenticada MUST NOT alterar as categorias de um canal que não é seu. Canal sem dono MUST NOT receber categorias por essa edição.

#### Scenario: Conta alheia
- **WHEN** outra conta envia categorias para o canal de um dono
- **THEN** a API recusa e as categorias públicas desse canal permanecem

### Requirement: Categorias na leitura pública
`GET /v1/channels/{handle}` MUST incluir somente as categorias ativas associadas àquele canal. Categoria inativa associada MUST NOT aparecer nessa leitura.

#### Scenario: Categoria desativada depois
- **WHEN** uma categoria associada ao canal deixa de estar ativa
- **THEN** a leitura pública do canal não inclui essa categoria

### Requirement: Categorias sem banco
Se o PostgreSQL não puder atender a lista ou a associação, a API MUST responder 503. O corpo da lista MUST NOT ser uma lista vazia de sucesso. `GET /healthz` MUST continuar 200.

#### Scenario: Banco parado na lista
- **WHEN** o banco não atende `GET /v1/categories`
- **THEN** a resposta é 503 e o corpo não é uma lista de categorias
