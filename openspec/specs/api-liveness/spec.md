# api-liveness Specification

## Purpose

Confirma que o processo da API está vivo, configurado e capaz de encerrar sem abandonar o listener, antes de existir qualquer recurso de negócio.

## Requirements

### Requirement: Liveness pública
O processo HTTP MUST responder a `GET /healthz` sem credencial, com status 200 e um JSON cujo campo `status` seja `ok`. Essa resposta MUST NOT consultar PostgreSQL nem afirmar que dependências estão prontas.

#### Scenario: Cliente sem credencial
- **WHEN** um cliente chama `GET /healthz` sem cabeçalho de autorização
- **THEN** a resposta é 200, `Content-Type` JSON e o corpo contém `status` igual a `ok`

#### Scenario: Método não contratado
- **WHEN** um cliente chama `/healthz` com método diferente de GET
- **THEN** a resposta não é o JSON de liveness com status 200

### Requirement: Configuração inválida impede o boot
O processo MUST recusar iniciar o listener quando o endereço ou um prazo HTTP explícito for ausente ou inválido, e MUST apresentar o erro sem usar panic para essa entrada.

#### Scenario: Endereço ausente
- **WHEN** o endereço de escuta não é informado
- **THEN** a configuração falha e nenhum listener é aberto

#### Scenario: Prazo inválido
- **WHEN** um prazo HTTP é informado com valor que não é uma duração positiva
- **THEN** a configuração falha antes de aceitar conexões

### Requirement: Encerramento com prazo
Ao receber sinal de interrupção ou término, o processo MUST deixar de aceitar conexões novas e MUST encerrar as que já estão em curso dentro de um prazo configurado.

#### Scenario: Sinal de término
- **WHEN** o processo em execução recebe cancelamento de ciclo de vida
- **THEN** o listener para de aceitar conexões e a parada conclui dentro do prazo configurado
