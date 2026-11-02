# Spec Delta

## Purpose

Mostra o progresso do envio direto ao storage e separa bytes enviados de vídeo processado.

## ADDED Requirements

### Requirement: Progresso do envio direto
Durante o envio autorizado, a interface MUST mostrar o quanto dos bytes já foi aceito pelo storage. Ao atingir o total dos bytes, ela MUST NOT afirmar que o vídeo está processado, publicado ou pronto para assistir.

#### Scenario: Envio pela metade
- **WHEN** metade dos bytes do arquivo já foi aceita pelo storage
- **THEN** a interface mostra progresso parcial e não diz que o vídeo está pronto

#### Scenario: Bytes concluídos
- **WHEN** o storage aceitou todos os bytes e a confirmação ainda não voltou
- **THEN** a interface não anuncia vídeo processado nem publicado

### Requirement: Falha do envio
Se o envio direto falha ou é interrompido, a interface MUST mostrar a falha e MUST NOT dizer que o vídeo foi enfileirado.

#### Scenario: Storage recusa o envio
- **WHEN** o storage rejeita a transferência antes do fim
- **THEN** a interface mostra a falha e não anuncia processamento
