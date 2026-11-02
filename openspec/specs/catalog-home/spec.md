# catalog-home Specification

## Purpose

Mostra na página inicial os vídeos públicos já persistidos e separa lista com itens, lista vazia, carregamento e falha da API.

## Requirements

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
A leitura inicial MUST ocorrer no servidor, usando apenas o contrato público. Essa leitura MUST NOT enviar credencial de sessão. O catálogo MUST NOT gravar token no navegador. A identidade da sessão, quando existir, pertence ao shell e MUST NOT aparecer nos dados do card.

#### Scenario: Primeira visita
- **WHEN** um visitante abre a página inicial
- **THEN** o HTML já contém a grade, o vazio ou o erro, e a consulta pública segue sem credencial

#### Scenario: Sessão no shell
- **WHEN** a sessão está restaurada na mesma página
- **THEN** o card continua sem credencial e sem dados da conta
