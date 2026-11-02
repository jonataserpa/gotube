# GoTube — desenvolvimento assistido com Cursor + OpenSpec

Este repositório começa com contexto, regras, curso e ferramentas de preparação.
Não contém ainda aplicação Go/Next executável. O primeiro prompt constrói essa base.

## Começar
1. Instalar/inicializar OpenSpec: `bash scripts/setup-openspec.sh` (requer rede/Node/npm).
2. Abrir esta raiz no Cursor: `cursor .`, ou File > Open Folder.
3. Abrir Agent Chat e usar o conteúdo de `docs/ai/PROMPT_INICIAL_CURSOR.md`.
4. Referenciar com @: `AGENTS.md` e `docs/course/GoTube_Curso_Completo.md`.
5. O prompt autoriza somente bootstrap-project, não o produto inteiro.

Não execute npm dev/go run antes de o bootstrap criar módulos/scripts reais.
O agente deve atualizar este README com os comandos implementados e testados.

## Curso
docs/course/GoTube_Curso_Completo.md contém a versão full stack completa:
32 módulos Go, 24 Next.js, 516 horas estimadas. O texto é referência de aprendizado.

## Instruções e qualidade
AGENTS.md na raiz, internal/ e web/. Regras focadas em .cursor/rules/.
Especificações em openspec/; decisões em docs/adr/; contrato em api/ após bootstrap.
Não há Superpowers instalado; considerar depois de algumas funcionalidades concluídas.

## OpenSpec local
O setup instala somente em .tooling/openspec, sem instalação global nem sudo.
Primeira execução resolve latest; --save-exact e lockfile fixam a versão escolhida.
Reexecução sem argumento usa npm ci do lockfile. Para versão específica:
`bash scripts/setup-openspec.sh VERSAO` (substitua VERSAO por uma release real).
CLI: `bash scripts/openspec.sh --help`.
Se o download falhar, corrija a rede e execute setup novamente; o scaffold permanece.

## Escopo dos scripts
Não criam repositório remoto, não fazem commit/push, não iniciam containers,
não instalam Go/Node/Docker e não implantam aplicação. O prompt inicial orienta
o Cursor a criar e testar a aplicação mínima conforme as ferramentas disponíveis.
