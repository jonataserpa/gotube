# Desenvolvimento assistido no Cursor

## Ferramentas
Cursor + OpenSpec + AGENTS.md e regras focadas. Superpowers não é instalado por este starter.
Os arquivos de integração em .cursor/commands e .cursor/skills são gerados pelo OpenSpec.
Não copie/adapte skills manualmente para substituir a inicialização oficial.

## Fluxo por mudança
1. Ler contexto relevante; esclarecer somente decisões bloqueantes.
2. /opsx-propose nome-da-mudanca no chat Cursor, se disponível na instalação.
3. Revisar proposta, cenários, design e tasks. Bootstrap já está autorizado no prompt inicial.
4. /opsx-apply nome-da-mudanca. Implementar fatias pequenas e testar.
5. Verificar contrato e jornada. /opsx-verify depende de workflow expandido habilitado.
6. Rever diff e integrar conforme autorização. Arquivar após conclusão real, não só checklist.

Os slash commands são do chat, não do shell. Para CLI local use bash scripts/openspec.sh.
Se só skills estiverem disponíveis, use a skill exposta pelo Cursor. Não alegue que comando
existe sem verificar arquivos/autocomplete da instalação.

## Perfis opcionais
bash scripts/openspec.sh config profile
bash scripts/openspec.sh update

Escolha verify no perfil desejado. Esta configuração pode alterar preferências globais
do OpenSpec para o usuário; faça conscientemente. Core basta para começar.

## Uso diário
Uma mudança ativa de cada vez. Uma fonte de requisitos; sem segunda cópia de tasks em
docs/plans. ADR explica decisão; OpenAPI explica transporte; spec explica comportamento.
Modo didático: @learning-mode. Para retomar, use docs/ai/PROMPT_CONTINUAR.md.

## Referências oficiais verificadas em 14/10/2026
- https://github.com/Fission-AI/OpenSpec
- https://github.com/Fission-AI/OpenSpec/blob/main/docs/cli.md
- https://github.com/Fission-AI/OpenSpec/blob/main/docs/supported-tools.md
- https://github.com/Fission-AI/OpenSpec/blob/main/docs/customization.md
- https://cursor.com/docs/rules
- https://github.com/obra/superpowers

Rules são orientação; CI/testes são verificações executáveis. Revisar comportamento da
versão instalada antes de atualizar. Não há integração automática com Superpowers aqui.
