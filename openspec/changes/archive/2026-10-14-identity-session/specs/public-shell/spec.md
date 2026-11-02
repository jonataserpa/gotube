# Spec Delta

## MODIFIED Requirements

### Requirement: Sem sessão ou catálogo fictícios
Sem sessão restaurada, a interface MUST NOT apresentar um usuário autenticado. Ela MUST oferecer entrada real para a conta. Ela MUST NOT listar vídeos, canais ou resultados de busca inventados. Com sessão restaurada, o shell MUST mostrar somente essa conta e a ação de saída. Busca, Novo vídeo e demais ações ainda inexistentes MUST permanecer indisponíveis. A listagem de vídeos públicos persistidos pertence à capacidade catalog-home.

#### Scenario: Visitante sem sessão
- **WHEN** a página é aberta sem sessão
- **THEN** a interface não apresenta um usuário autenticado e oferece uma ação de entrada funcional

#### Scenario: Sessão restaurada
- **WHEN** o navegador restaura a sessão desta conta
- **THEN** o shell mostra a identidade dessa conta e a saída, sem identidade de outra conta

#### Scenario: Conteúdo do catálogo
- **WHEN** um visitante abre a página inicial
- **THEN** a página não declara que o catálogo público ainda não existe e não mostra itens inventados
