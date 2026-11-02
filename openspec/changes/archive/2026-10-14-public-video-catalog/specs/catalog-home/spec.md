# Spec Delta

## Purpose

Mostra na página inicial os vídeos públicos já persistidos e separa lista com itens, lista vazia, carregamento e falha da API.

## ADDED Requirements

### Requirement: Grade pública
A página inicial MUST renderizar no servidor os itens públicos recebidos da listagem. Cada card MUST mostrar título, duração e o nome público do canal. Thumbnail presente MUST ser imagem com texto alternativo. Thumbnail ausente MUST reservar um espaço com texto acessível. O card MUST NOT apontar para uma página de reprodução ou de canal ainda inexistente.

#### Scenario: Lista com um vídeo
- **WHEN** a listagem pública devolve um vídeo com título, duração, canal e thumbnail
- **THEN** a página inicial mostra esses dados e a imagem tem texto alternativo

#### Scenario: Thumbnail ausente
- **WHEN** a listagem devolve um vídeo sem thumbnail
- **THEN** a página reserva a área do quadro e informa que não há imagem, sem simular um arquivo de mídia

### Requirement: Vazio diferente de erro
Uma página sem itens e com API disponível MUST informar que não há vídeos públicos. Uma falha de rede, HTTP de erro ou payload inválido MUST mostrar erro recuperável. O erro MUST NOT reutilizar a mensagem de lista vazia nem a mensagem de catálogo ainda não implementado.

#### Scenario: Nenhum vídeo público
- **WHEN** a listagem responde com sucesso e nenhum item
- **THEN** a página explica que ainda não há vídeos públicos

#### Scenario: API indisponível
- **WHEN** a leitura pública da API falha
- **THEN** a página mostra erro e um caminho para tentar de novo, sem grade vazia fingindo sucesso

### Requirement: Carregamento da próxima página
A página seguinte MUST usar o cursor na URL. Enquanto a navegação para outra página não termina, a interface MUST indicar carregamento. A página concluída MUST mostrar só o resultado daquela consulta.

#### Scenario: Próximo cursor
- **WHEN** o visitante pede a página indicada por `next_cursor`
- **THEN** a URL carrega esse cursor e o resultado continua a lista sem repetir o último item da página anterior

#### Scenario: Navegação em andamento
- **WHEN** a troca de página ainda não recebeu resposta
- **THEN** a interface indica carregamento

### Requirement: Consulta pública no servidor
A leitura inicial MUST ocorrer no servidor, usando apenas o contrato público. A página MUST NOT apresentar usuário autenticado, enviar credencial de sessão nem gravar token no navegador.

#### Scenario: Primeira visita
- **WHEN** um visitante abre a página inicial
- **THEN** o HTML já contém a grade, o vazio ou o erro, e não mostra uma sessão autenticada
