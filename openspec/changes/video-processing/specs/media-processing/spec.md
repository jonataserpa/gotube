# Spec Delta

## Purpose

Transforma o objeto enviado em thumbnail e HLS e registra o resultado sem publicar o vídeo.

## ADDED Requirements

### Requirement: Artefatos da geração
O worker MUST inspecionar o objeto da sessão e gravar uma thumbnail e uma playlist HLS com ao menos um segmento, em keys da geração criadas no servidor. MUST NOT usar URL de mídia enviada pelo cliente.

#### Scenario: Objeto com vídeo
- **WHEN** o worker recebe `video.uploaded.v1` de um objeto que tem trilha de vídeo
- **THEN** existem thumbnail, playlist e segmento dessa geração

#### Scenario: Sem trilha de vídeo
- **WHEN** a inspeção não encontra vídeo
- **THEN** o vídeo fica `FAILED` e não há playlist nova

### Requirement: Conclusão sem publicar
A conclusão MUST valer só se a geração e o token do job ainda são os vigentes. Sucesso MUST marcar `READY`, gravar a duração inspecionada e um único `video.ready.v1`, e MUST deixar o vídeo `DRAFT` e `PRIVATE`.

#### Scenario: Geração vigente
- **WHEN** os artefatos são válidos e o job ainda é o vigente
- **THEN** o vídeo fica `READY` e `DRAFT`, com um evento `video.ready.v1`

#### Scenario: Geração antiga
- **WHEN** o job não é mais a geração vigente
- **THEN** esse job não altera o estado do vídeo

### Requirement: Entrega repetida
Consumir de novo o evento de uma geração já `READY` MUST NOT criar outro asset nem outro `video.ready.v1`.

#### Scenario: Mensagem reentregue
- **WHEN** o worker recebe outra vez o evento da geração já concluída
- **THEN** continua existindo um único conjunto de assets dessa geração

### Requirement: Falha limitada
Falha transitória MUST parar no máximo de tentativas. Esgotado o máximo, ou diante de falha permanente, o vídeo MUST ficar `FAILED` com erro que não inclui a object key.

#### Scenario: Tentativas esgotadas
- **WHEN** a mesma geração falha até o máximo configurado
- **THEN** `processing_status` é `FAILED` e o erro não contém a object key

### Requirement: Exemplos do seed
Vídeo que já está `READY` e não tem evento de upload MUST NOT ser transcodificado.

#### Scenario: Catálogo de exemplo
- **WHEN** o worker está ativo
- **THEN** os vídeos publicados do seed continuam `READY` sem asset novo
