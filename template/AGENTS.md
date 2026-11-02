# GoTube — instruções para agentes

## Contexto
- Formação full stack em Go e Next.js; monorepo com Go na raiz e web/ para Next.
- Curso: docs/course/GoTube_Curso_Completo.md. Consulte apenas seções relevantes.
- Produto: docs/product.md. Sequência: docs/roadmap.md. Decisões: docs/adr/.
- O curso descreve objetivos; não prova que o código já implementa esses objetivos.
- No scaffold inicial ainda não existem go.mod, app Next ou CI executável.

## Arquitetura
- Go é autoridade de identidade, autorização e regras de negócio.
- Next.js não acessa diretamente as tabelas do Go.
- API, worker e outbox são executáveis de um monólito modular Go.
- Upload vai ao object storage; FFmpeg é supervisionado por worker limitado.
- Live usa servidor de mídia, com controle/autorização no Go.
- SSR público e sessão privada restaurada no browser são a decisão inicial.
- Não introduza microsserviços, BFF ou novos protocolos sem necessidade e ADR.

## Fontes e escopo
- Leia o código, testes, mudança ativa e contratos antes de editar.
- openspec/specs/ contém comportamento consolidado; changes/ contém propostas.
- Não trate checklist de tarefas como evidência de teste.
- Contrato HTTP é api/openapi.yaml, a criar na primeira implementação.
- Se houver divergência entre código, ADR, spec e contrato, explicite e resolva.
- Preserve trabalho existente e evite refatorações alheias à tarefa.
- Não invente endpoints, payloads, scripts de package.json ou resultados.
- Não edite código gerado manualmente; altere a fonte e regenere.
- Não modifique arquivos gerados do OpenSpec para mudar regras do projeto.

## Autonomia e fluxo
- Implemente a etapa autorizada até seu aceite, sem pedir confirmação a cada arquivo.
- Pergunte apenas por decisão bloqueante que não possa ser inferida com segurança.
- Propostas de funcionalidades novas passam por revisão antes de implementação.
- O prompt inicial autoriza proposta e implementação de bootstrap-project apenas.
- Execute sequencialmente; agentes paralelos somente se solicitados, com escopos claros.
- Não faça push, merge, deploy, ações destrutivas ou configuração de contas sem autorização.

## Qualidade
- Para regras e bugs, escreva testes que demonstrem comportamento relevante.
- Reproduza bugs antes da correção quando viável.
- Teste autorização por recurso, concorrência e falhas de fronteira aplicáveis.
- Não enfraqueça testes para obter sucesso.
- Execute verificações focadas e depois os gates definidos do projeto.
- Se um comando não existir ou ferramenta faltar, informe e corrija o setup quando autorizado.
- Nunca declare build/teste/deploy executado sem evidência.
- Mantenha segredos fora de Git, prompts, logs e fixtures.

## Entrega e aprendizado
- Explique o comportamento alterado, decisões, arquivos e verificações reais.
- Explique brevemente o conceito Go/React/Next necessário nesta tarefa.
- Registre o próximo passo em docs/learning/progress.md.
- Só marque tarefa concluída quando seu critério de aceite estiver demonstrado.
