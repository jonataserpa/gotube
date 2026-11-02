# Spec Delta

## MODIFIED Requirements

### Requirement: Grade pública
A página inicial MUST renderizar no servidor os itens públicos recebidos da listagem. Cada card MUST mostrar título, duração e o nome público do canal. O nome do canal MUST apontar para a página pública desse canal. Thumbnail presente MUST ser imagem com texto alternativo. Thumbnail ausente MUST reservar um espaço com texto acessível. O card MUST NOT apontar para uma página de reprodução ainda inexistente.

#### Scenario: Lista com um vídeo
- **WHEN** a listagem pública devolve um vídeo com título, duração, canal e thumbnail
- **THEN** a página inicial mostra esses dados e a imagem tem texto alternativo

#### Scenario: Thumbnail ausente
- **WHEN** a listagem devolve um vídeo sem thumbnail
- **THEN** a página reserva a área do quadro e informa que não há imagem, sem simular um arquivo de mídia

#### Scenario: Nome do canal
- **WHEN** o card mostra o nome público de um canal
- **THEN** esse nome leva à página pública do canal e o card não leva a uma página de reprodução
