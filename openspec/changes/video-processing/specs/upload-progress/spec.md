# Spec Delta

## ADDED Requirements

### Requirement: Estado depois do envio
Depois da confirmação do envio, a interface MUST mostrar se o vídeo está na fila, em processamento, concluído ou falhou. Em `QUEUED` ou `PROCESSING`, MUST NOT afirmar que está publicado ou pronto para assistir. Em `FAILED`, MUST mostrar a falha e uma ação de nova tentativa, e MUST NOT afirmar que foi publicado. Em `READY`, MUST NOT afirmar que foi publicado ou que dá para assistir.

#### Scenario: Ainda processando
- **WHEN** a confirmação já ocorreu e o estado é `QUEUED` ou `PROCESSING`
- **THEN** a interface mostra esse estado e não anuncia publicação

#### Scenario: Processamento falhou
- **WHEN** o estado é `FAILED`
- **THEN** a interface mostra a falha e oferece nova tentativa

#### Scenario: Processamento concluído
- **WHEN** o estado é `READY`
- **THEN** a interface não anuncia vídeo publicado nem pronto para assistir
