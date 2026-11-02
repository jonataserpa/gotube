# Spec Delta

## Purpose

Oferece a primeira superfície pública do GoTube, com navegação e tema utilizáveis, sem simular sessão, vídeos ou busca.

## ADDED Requirements

### Requirement: Shell acessível
A página inicial MUST expor linguagem `pt-BR`, um link para pular ao conteúdo, um landmark de navegação e um landmark principal. O controle de tema MUST ter nome acessível estável na primeira renderização.

#### Scenario: Leitura da página inicial
- **WHEN** a página inicial é renderizada
- **THEN** existem o link de pular conteúdo, a navegação com destino Início e a região principal

#### Scenario: Teclado no tema
- **WHEN** o controle de tema recebe foco
- **THEN** ele é operável por teclado e expõe um nome acessível que não depende só de cor

### Requirement: Sem sessão ou catálogo fictícios
A página inicial MUST NOT apresentar um usuário autenticado nem uma ação de login funcional. Ela MUST NOT listar vídeos, canais ou resultados de busca inventados. A listagem de vídeos públicos persistidos pertence à capacidade catalog-home.

#### Scenario: Visitante sem sessão
- **WHEN** a página é aberta sem sessão
- **THEN** a interface não apresenta um usuário autenticado nem uma ação de login funcional

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
