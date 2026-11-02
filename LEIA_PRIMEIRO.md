# GoTube — pacote para começar no Cursor

## O que vem pronto

- Script de preparação para Linux, macOS e WSL (bash + Python 3).
- Curso full stack completo: 56 módulos, Go e Next.js.
- AGENTS.md raiz, backend e frontend.
- 3 regras Cursor: contratos, mídia e modo de aprendizado.
- Configuração OpenSpec, contexto de produto/arquitetura e roadmap.
- Instalador local OpenSpec com lockfile, sem sudo ou instalação global.
- Prompt inicial e prompts para retomar e criar o catálogo.

O pacote prepara o repositório e as instruções. **Não contém a aplicação GoTube
implementada**: o prompt orienta o Cursor a construir a base executável primeiro.
Não são criados repositório GitHub, push, commit, deploy ou containers pelo script.

## 1. Extrair

Extraia GoTube_Starter_Cursor.zip e entre na pasta GoTube_Starter_Cursor.
Mantenha template/, bootstrap.py, directories.json e init-gotube.sh juntos.
O script depende desses arquivos; não copie apenas o .sh.

## 2. Preparar o projeto

Com Git, Python 3, bash, Node.js e npm instalados:

```bash
bash init-gotube.sh ../gotube --openspec
```

O destino deve ser novo ou vazio. O script recusa destinos com arquivos.
O OpenSpec exige Node >=20.19 na documentação consultada. Ao criar o Next, o Cursor
deve conferir e registrar o requisito da versão escolhida, que pode ser mais alto.

Sem instalar ferramentas agora (preparação offline):

```bash
bash init-gotube.sh ../gotube
cd ../gotube
bash scripts/setup-openspec.sh
```

Sem Git, acrescente --no-git. Caminhos com espaço devem ser colocados entre aspas.
Para fixar uma release conhecida: --openspec-version VERSAO (trocar por versão real).
A primeira instalação resolve latest e salva versão exata + lockfile em .tooling/openspec.
Repetir setup sem argumento usa npm ci desse lockfile; atualização explícita recebe versão.
Se instalação/download falhar, não rode init sobre a pasta já preenchida: rode setup novamente.

## 3. Abrir no Cursor

```bash
cd ../gotube
cursor .
```

Se cursor não estiver no PATH, abra pelo menu File > Open Folder e selecione gotube.
Não abra a pasta do pacote ou template/ como projeto de desenvolvimento.

## 4. Anexar contexto e começar

No Agent Chat do Cursor, digite @ e selecione:

- AGENTS.md
- docs/ai/PROMPT_INICIAL_CURSOR.md
- docs/course/GoTube_Curso_Completo.md

Cole o conteúdo de PROMPT_INICIAL_CURSOR.md ou diga:

> Execute as instruções de @docs/ai/PROMPT_INICIAL_CURSOR.md neste repositório.
> Use @docs/course/GoTube_Curso_Completo.md como referência e respeite @AGENTS.md.
> O escopo autorizado é somente bootstrap-project. Comece pelo diagnóstico e
> implemente até os critérios de aceite, explicando os conceitos e testando.

O curso já foi copiado para dentro do projeto; não precisa anexar ZIP nem todo o
Markdown a cada mensagem. Mencione as seções relevantes nas próximas tarefas.

## 5. Resultado esperado da primeira sessão

API Go com /healthz; Next.js com shell responsivo/tema; PostgreSQL local via Compose;
OpenAPI mínimo; scripts reais; testes; CI; README atualizado. Nada de vídeos falsamente
funcionais, login incompleto ou todos os serviços ao mesmo tempo.

Quando concluir, revise resultados e use PROMPT_CATALOGO.md para especificar a próxima
feature. Os comandos /opsx-* ficam no chat; scripts/openspec.sh é o wrapper de terminal.
Se comandos não aparecerem, confira a saída do init, os arquivos .cursor gerados,
o modo de entrega configurado e abra uma nova sessão. Não crie slash commands fictícios.

## Limitações e verificação

A geração offline e as proteções de destino são verificadas no relatório VERIFICACAO.md.
Download/execução real de OpenSpec, criação Next, compilação Go e execução de Docker
não foram realizados por este pacote. Serão feitos na sua máquina/Cursor conforme
ferramentas e rede disponíveis. Superpowers não é instalado nesta primeira etapa.

Fontes oficiais e workflow estão em template/docs/ai/workflow.md.
