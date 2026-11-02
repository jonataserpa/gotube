# public-shell Specification

## Purpose

Oferece navegação, tema e uma sessão real quando ela existir, sem vídeo ou busca inventados.

## Requirements

### Requirement: Shell acessível
A página inicial MUST expor linguagem `pt-BR`, um link para pular ao conteúdo, um landmark de navegação e um landmark principal. O controle de tema MUST ter nome acessível estável na primeira renderização.

#### Scenario: Leitura da página inicial
- **WHEN** a página inicial é renderizada
- **THEN** existem o link de pular conteúdo, a navegação com destino Início e a região principal

#### Scenario: Teclado no tema
- **WHEN** o controle de tema recebe foco
- **THEN** ele é operável por teclado e expõe um nome acessível que não depende só de cor

### Requirement: Sem sessão ou catálogo fictícios
Sem sessão restaurada, a interface MUST NOT apresentar um usuário autenticado. Ela MUST oferecer entrada real para a conta. Ela MUST NOT listar vídeos, canais ou resultados de busca inventados. Com sessão restaurada, o shell MUST mostrar somente essa conta, a ação de saída e um destino real para o perfil. Busca e demais ações ainda inexistentes MUST permanecer indisponíveis. Com sessão, Novo vídeo MUST abrir o fluxo de rascunho. A listagem de vídeos públicos persistidos pertence à capacidade catalog-home. A página pública de um canal existente pertence à capacidade channel-profile.

#### Scenario: Visitante sem sessão
- **WHEN** a página é aberta sem sessão
- **THEN** a interface não apresenta um usuário autenticado e oferece uma ação de entrada funcional

#### Scenario: Sessão restaurada
- **WHEN** o navegador restaura a sessão desta conta
- **THEN** o shell mostra a identidade dessa conta e a saída, sem identidade de outra conta

#### Scenario: Perfil da conta
- **WHEN** a sessão desta conta está restaurada
- **THEN** o shell oferece um destino que abre o perfil dessa conta

#### Scenario: Novo vídeo
- **WHEN** a sessão desta conta está restaurada e a pessoa aciona Novo vídeo
- **THEN** a interface abre o fluxo de rascunho e a busca continua indisponível

#### Scenario: Conteúdo do catálogo
- **WHEN** um visitante abre a página inicial
- **THEN** a página não declara que o catálogo público ainda não existe e não mostra itens inventados

### Requirement: Tema sem divergir a hidratação
O tema claro ou escuro MUST poder ser alternado e MUST manter o mesmo HTML inicial entre servidor e cliente. A preferência persistida MUST ser aplicada sem trocar o texto do primeiro render.

#### Scenario: Primeira pintura coerente
- **WHEN** o documento hidrata no navegador
- **THEN** o texto visível do shell coincide com o HTML enviado pelo servidor

#### Scenario: Alternância
- **WHEN** o visitante aciona o controle de tema
- **THEN** o documento passa a usar o outro tema e a escolha fica disponível na visita seguinte neste navegador
