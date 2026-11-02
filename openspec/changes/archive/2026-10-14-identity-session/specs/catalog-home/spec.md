# Spec Delta

## MODIFIED Requirements

### Requirement: Consulta pública no servidor
A leitura inicial MUST ocorrer no servidor, usando apenas o contrato público. Essa leitura MUST NOT enviar credencial de sessão. O catálogo MUST NOT gravar token no navegador. A identidade da sessão, quando existir, pertence ao shell e MUST NOT aparecer nos dados do card.

#### Scenario: Primeira visita
- **WHEN** um visitante abre a página inicial
- **THEN** o HTML já contém a grade, o vazio ou o erro, e a consulta pública segue sem credencial

#### Scenario: Sessão no shell
- **WHEN** a sessão está restaurada na mesma página
- **THEN** o card continua sem credencial e sem dados da conta
