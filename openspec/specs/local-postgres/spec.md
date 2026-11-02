# local-postgres Specification

## Purpose

Disponibiliza um PostgreSQL local reproduzível para as próximas etapas, sem transformar a liveness da API em prova de que o banco está pronto.

## Requirements

### Requirement: Banco local reproduzível
O ambiente de desenvolvimento MUST subir PostgreSQL com imagem de versão fixada, porta publicada, volume nomeado e credenciais lidas de variáveis de ambiente locais. Segredos reais MUST NOT entrar no Git; o exemplo versionado MUST permanecer um valor só de laboratório.

#### Scenario: Configuração do Compose
- **WHEN** a configuração do Compose é renderizada com o ambiente de exemplo
- **THEN** o serviço PostgreSQL declara imagem fixada, volume persistente e variáveis de banco, usuário e senha

#### Scenario: Volume após recreação
- **WHEN** o container do banco é recriado sem remover o volume nomeado
- **THEN** o volume de dados continua associado ao serviço

### Requirement: Liveness independente do banco
A API MUST responder liveness mesmo que o PostgreSQL esteja parado. A saúde do banco MUST permanecer um sinal próprio do container, não o corpo de `/healthz`.

#### Scenario: API sem banco
- **WHEN** o PostgreSQL não está em execução e a API sobe só com o endereço HTTP válido
- **THEN** `GET /healthz` continua retornando 200 e `status` igual a `ok`

#### Scenario: Saúde do container
- **WHEN** o container PostgreSQL fica pronto
- **THEN** a checagem de saúde do serviço confirma que o banco aceita conexões, sem alterar o contrato de `/healthz`
